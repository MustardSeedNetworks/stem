// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"
	"sync"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// BackgroundComponents owns the long-lived background goroutines attached to a
// running Server — the lifecycle scoped to Run rather than to construction.
// Today that is the always-on reflector-stats SSE publisher
// (runReflectorStatsPublisher); future server-push producers (e.g. the
// test-progress publisher, #296 follow-up) register here so there is a single
// ordered Start/Stop seam instead of goroutines launched ad hoc inside Run.
//
// Unlike seed's BackgroundComponents — which holds detached feature services
// (the reporting scheduler, the Wi-Fi visibility loop) — stem's one background
// producer reads live Server state (reflector executor, stats), so the holder
// keeps a back-reference to the Server it coordinates rather than owning
// standalone services. See ADR-0005.
//
// Construction-scoped cleanup goroutines (the rate limiters, CSRF manager and
// auth manager created in NewServer) are NOT owned here: they outlive Run and
// must be stopped by Server.Shutdown even when Run was never called (the test
// suite constructs servers and calls Shutdown directly). Mixing them in would
// leak those goroutines whenever Start was skipped.
type BackgroundComponents struct {
	srv    *Server
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// newBackgroundComponents returns a holder bound to srv. Nothing is started
// until Start is called.
func newBackgroundComponents(srv *Server) *BackgroundComponents {
	return &BackgroundComponents{srv: srv}
}

// Start launches the background goroutines. Each runs under a context derived
// from ctx, so cancelling ctx (server shutdown signal) or calling Stop both
// terminate them; the WaitGroup lets Stop block until they have fully exited.
// Start is not safe to call twice.
func (b *BackgroundComponents) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	b.wg.Go(func() {
		b.srv.runReflectorStatsPublisher(runCtx)
	})

	logging.Debug("background components started")
}

// Stop cancels the background goroutines and blocks until they have exited.
// It is safe to call when Start was never invoked (no-op) and idempotent.
func (b *BackgroundComponents) Stop() {
	if b.cancel == nil {
		return
	}
	b.cancel()
	b.cancel = nil
	b.wg.Wait()
	logging.Debug("background components stopped")
}
