// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api/sse"
)

// TestBackgroundComponentsStartStop verifies the holder starts its goroutine
// and that Stop blocks until it has exited (no leak under -race) and returns
// promptly. A bare Server is sufficient: the publisher short-circuits when no
// reflector executor is set, so it never touches uninitialised fields.
func TestBackgroundComponentsStartStop(t *testing.T) {
	t.Parallel()

	s := &Server{sseBroadcaster: sse.New()}
	bg := newBackgroundComponents(s)

	bg.Start(context.Background())

	done := make(chan struct{})
	go func() {
		bg.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("BackgroundComponents.Stop did not return; publisher goroutine leaked")
	}
}

// TestBackgroundComponentsStopBeforeStart verifies Stop is a no-op when Start
// was never called (Server.Shutdown calls it on servers that never ran).
func TestBackgroundComponentsStopBeforeStart(t *testing.T) {
	t.Parallel()

	bg := newBackgroundComponents(&Server{})
	bg.Stop() // must not panic or block
}

// TestBackgroundComponentsStopIdempotent verifies a second Stop is harmless.
func TestBackgroundComponentsStopIdempotent(t *testing.T) {
	t.Parallel()

	s := &Server{sseBroadcaster: sse.New()}
	bg := newBackgroundComponents(s)
	bg.Start(context.Background())
	bg.Stop()
	bg.Stop() // second call is a no-op
}

// TestBackgroundComponentsCtxCancelStops verifies cancelling the parent context
// terminates the goroutine even without an explicit Stop, so a SIGINT/SIGTERM
// that cancels Run's context cleanly winds the publisher down.
func TestBackgroundComponentsCtxCancelStops(t *testing.T) {
	t.Parallel()

	s := &Server{sseBroadcaster: sse.New()}
	bg := newBackgroundComponents(s)

	ctx, cancel := context.WithCancel(context.Background())
	bg.Start(ctx)
	cancel()

	// After the parent ctx is cancelled the goroutine exits on its own; Stop
	// then only has to observe the already-finished WaitGroup and return.
	done := make(chan struct{})
	go func() {
		bg.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("goroutine did not exit after parent context cancellation")
	}
}

// A background loop reads live dataplane state (the reflector executor and
// its stats), so a fault there is a Go panic in a goroutine no caller is
// waiting on. Before #1336 that panic killed the daemon. Supervised, the
// worker is restarted and the daemon keeps serving.
func TestBackgroundPanicIsRestartedAndTheDaemonStaysUp(t *testing.T) {
	t.Parallel()

	panics := make(chan struct{}, backgroundRestarts+1)
	bg := newBackgroundComponents(&Server{sseBroadcaster: sse.New()})
	bg.workers = []backgroundWorker{{
		name: "panicking-probe",
		run: func(context.Context) {
			select {
			case panics <- struct{}{}:
			default:
			}
			panic("forced background fault")
		},
	}}

	bg.Start(context.Background())

	// One entry is the first call; a second proves the supervisor restarted
	// the worker instead of letting the panic escape.
	for range 2 {
		select {
		case <-panics:
		case <-time.After(5 * time.Second):
			t.Fatal("the panicking worker was not restarted")
		}
	}

	done := make(chan struct{})
	go func() {
		bg.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return after a worker exhausted its restarts")
	}
}
