// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"errors"
	"sync"
	"time"
)

// Account lockout matches seed's: five consecutive failures lock the
// account for fifteen minutes. The per-client limit (a per-IP token bucket
// in internal/api/ratelimit) is separate and still applies in front of this;
// the lockout is what stops a guess spread across many addresses.
const (
	// MaxFailedLogins is the number of consecutive failed logins that lock
	// the administrator account.
	MaxFailedLogins = 5
	// LockoutDuration is how long a locked account refuses password logins.
	LockoutDuration = 15 * time.Minute
)

// ErrAccountLocked is returned in place of a credential check while the
// account is locked. The HTTP layer answers it exactly like a wrong
// password, so a client learns nothing it could not from the timing; the
// distinct error exists so the audit trail can tell the two apart.
var ErrAccountLocked = errors.New("account locked after repeated failed logins")

// lockout counts failures against the one real administrator account. It is
// never keyed on a submitted username, so a flood of invented names cannot
// grow it. A failure keeps counting while locked, so each attempt that gets
// through after the lock expires re-locks at once.
type lockout struct {
	failures int
	until    time.Time
	// inflight counts attempts that passed the gate and have not reported
	// yet. They are charged against the limit up front: the check itself is
	// slow (a password hash), so counting only finished failures would let a
	// burst of parallel requests all pass before the first one is recorded.
	inflight int
}

// LoginLocked reports whether password logins are currently refused.
func (m *Manager) LoginLocked() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.now().Before(m.lock.until)
}

// TryBeginLoginAttempt is the single gate in front of every password or
// second-factor check. It refuses while the account is locked, or when the
// failures already counted plus the attempts still running could reach the
// limit; otherwise it reserves a slot. The caller must call release exactly
// once. release(false) records a failure; release(true) only frees the slot,
// because a correct password is not a completed login while a second factor
// is pending, so the count is cleared by [Manager.ResetLoginFailures] when
// the login completes.
func (m *Manager) TryBeginLoginAttempt() (func(success bool), bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	if !m.lock.until.IsZero() && !now.Before(m.lock.until) {
		// The lock has run out. One strike is left, so the first failure
		// after it re-locks at once, as seed's does.
		m.lock.until = time.Time{}
		m.lock.failures = MaxFailedLogins - 1
	}
	if now.Before(m.lock.until) || m.lock.failures+m.lock.inflight >= MaxFailedLogins {
		return nil, false
	}
	m.lock.inflight++
	var once sync.Once
	return func(success bool) {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.lock.inflight--
			if !success {
				m.recordFailureLocked()
			}
		})
	}, true
}

// RecordLoginFailure counts one failed login and reports whether the
// account is now locked.
func (m *Manager) RecordLoginFailure() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recordFailureLocked()
}

func (m *Manager) recordFailureLocked() bool {
	m.lock.failures++
	if m.lock.failures >= MaxFailedLogins {
		m.lock.until = m.now().Add(LockoutDuration)
	}
	return m.now().Before(m.lock.until)
}

// ResetLoginFailures clears the count and any lock. A completed login resets
// it, and so does a password set through setup or recovery, which already
// proved control of the host.
func (m *Manager) ResetLoginFailures() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lock = lockout{failures: 0, until: time.Time{}, inflight: m.lock.inflight}
}
