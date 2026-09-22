// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"

	"github.com/MustardSeedNetworks/foundation/pkg/supervise"

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
	srv     *Server
	workers []backgroundWorker
	group   *supervise.Group
}

// backgroundWorker is one named long-lived loop. The name is what an operator
// reads in the log line when the worker faults, so it is the component's
// name, not the function's.
type backgroundWorker struct {
	name string
	run  func(context.Context)
}

// newBackgroundComponents returns a holder bound to srv. Nothing is started
// until Start is called.
func newBackgroundComponents(srv *Server) *BackgroundComponents {
	return &BackgroundComponents{
		srv: srv,
		workers: []backgroundWorker{
			{name: "reflector-stats", run: srv.runReflectorStatsPublisher},
			{name: "connection-refresher", run: srv.runConnectionRefresher},
		},
	}
}

// backgroundRestarts is how many times a faulted background loop is restarted
// before the group gives up on it. These loops are resumable — a ticker and a
// token rotation, both stateless between iterations — so a transient fault
// should not cost the daemon its server-push channel for the rest of its life.
// A loop that faults every time stops after this many tries rather than
// spinning, and the supervisor logs the one line that says so.
const backgroundRestarts = 3

// Start launches the background goroutines. Each runs under a context derived
// from ctx, so cancelling ctx (server shutdown signal) or calling Stop both
// terminate them; the WaitGroup lets Stop block until they have fully exited.
// Start is not safe to call twice.
func (b *BackgroundComponents) Start(ctx context.Context) {
	group := supervise.New(logging.WithComponentLogger("background"))
	for _, w := range b.workers {
		group.Add(w.name, supervise.RestartN(backgroundRestarts), func(workerCtx context.Context) error {
			w.run(workerCtx)
			return nil
		})
	}
	b.group = group
	group.Start(ctx)

	logging.Debug("background components started")
}

// Stop cancels the background goroutines and blocks until they have exited.
// It is safe to call when Start was never invoked (no-op) and idempotent.
func (b *BackgroundComponents) Stop() {
	if b.group == nil {
		return
	}
	group := b.group
	b.group = nil
	// The caller's deadline for the ordered stop is the server's own
	// shutdown timeout, applied by Server.Shutdown around this call; the
	// workers here all return on context cancellation, so a bound of their
	// own would only invent a second timeout to disagree with it.
	if err := group.Stop(context.Background()); err != nil {
		logging.Warn("background components did not stop cleanly", "error", err.Error())
	}
	logging.Debug("background components stopped")
}
