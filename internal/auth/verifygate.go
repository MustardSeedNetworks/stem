// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Password hashing is a 64 MiB Argon2id derivation, so concurrent logins are
// bounded. Two caps, both applied before any username is looked at, so the
// outcome is the same for a real name and an invented one:
//
//   - per client: one address cannot hold more than MaxVerifiesPerClient
//     hashes at once, so a flood from one place cannot crowd out anyone else;
//     over it, the request is refused at once.
//   - global backstop: MaxConcurrentVerifies bounds memory. Worst case it is
//     MaxConcurrentVerifies x 64 MiB = 512 MiB resident for hashing. A request
//     over it waits up to verifyWait for a slot, then is refused.
//
// There is deliberately no pool reserved for the real username: that would
// make the answer depend on whether the name exists.
const (
	// MaxVerifiesPerClient is the concurrent hashes one client address may hold.
	MaxVerifiesPerClient = 2
	// MaxConcurrentVerifies is the global backstop, sized for 512 MiB worst case.
	MaxConcurrentVerifies = 8

	defaultVerifyWait = 2 * time.Second
)

// ErrTooManyAttempts is returned when a login cannot get a hashing slot. The
// HTTP layer answers 429 with Retry-After.
var ErrTooManyAttempts = errors.New("too many concurrent login attempts")

type verifyGate struct {
	slots chan struct{}
	wait  time.Duration

	mu      sync.Mutex
	clients map[string]int
}

func newVerifyGate() verifyGate {
	return verifyGate{
		slots:   make(chan struct{}, MaxConcurrentVerifies),
		wait:    defaultVerifyWait,
		mu:      sync.Mutex{},
		clients: make(map[string]int),
	}
}

// AcquireVerify reserves a hashing slot for a client, identified the way the
// per-address rate limiter identifies it (proxy-aware). The caller must call
// the returned release once the password check has finished, and must call
// this before it examines the submitted username.
func (m *Manager) AcquireVerify(ctx context.Context, client string) (func(), error) {
	g := &m.gate
	g.mu.Lock()
	if g.clients[client] >= MaxVerifiesPerClient {
		g.mu.Unlock()
		return nil, ErrTooManyAttempts
	}
	g.clients[client]++
	g.mu.Unlock()

	forget := func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		if g.clients[client]--; g.clients[client] <= 0 {
			delete(g.clients, client)
		}
	}

	timer := time.NewTimer(g.wait)
	defer timer.Stop()
	select {
	case g.slots <- struct{}{}:
	case <-timer.C:
		forget()
		return nil, ErrTooManyAttempts
	case <-ctx.Done():
		forget()
		return nil, ErrTooManyAttempts
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			<-g.slots
			forget()
		})
	}, nil
}
