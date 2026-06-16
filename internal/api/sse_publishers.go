// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"context"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api/sse"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// reflectorStatsInterval is how often the reflector-stats publisher
// computes and broadcasts a stats frame. 1 Hz matches the polling
// cadence the UI was using before this PR; the SSE channel just
// removes the round-trip overhead and tightens the latency.
const reflectorStatsInterval = time.Second

// runReflectorStatsPublisher periodically publishes the current reflector
// stats to all SSE subscribers until ctx is cancelled. It blocks for the
// lifetime of the publisher; [BackgroundComponents] owns the goroutine and
// the cancellation/await (see background.go).
//
// The loop is cheap when nobody's subscribed (it short-circuits without
// computing stats), so it's always-on rather than gated by subscriber count.
//
// Only broadcasts when stats actually exist (i.e., reflector mode is active
// and the executor is running). Subscribers in test-master mode just receive
// heartbeats until the mode flips.
func (s *Server) runReflectorStatsPublisher(ctx context.Context) {
	logging.Debug("SSE reflector-stats publisher started",
		"interval", reflectorStatsInterval.String())
	ticker := time.NewTicker(reflectorStatsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.broadcastReflectorStatsIfActive()
		}
	}
}

// broadcastReflectorStatsIfActive reads current reflector stats and
// publishes them as an SSE frame, but only when the reflector executor
// is active. We skip the broadcast (no frame at all) when there's
// nothing to report — subscribers see heartbeats only until reflector
// mode is engaged.
func (s *Server) broadcastReflectorStatsIfActive() {
	s.statsMu.RLock()
	exec := s.reflectorExec
	elapsed := time.Since(s.startTime).Seconds()
	s.statsMu.RUnlock()

	if exec == nil {
		return
	}

	stats := s.buildActiveReflectorStats(exec, elapsed)
	s.sseBroadcaster.Publish(sse.Frame{
		Type:    "reflector_stats",
		Payload: stats,
	})
}

// PublishTestProgress is the seam test runners call to push progress
// updates over SSE. The payload shape is left to the caller so each
// module (RFC 2544 throughput sweep, Y.1564 service test, etc.) can
// emit its own progress structure without a central type that grows
// fields for every test variant.
//
// Today this is exported for use by future test-runner integrations
// (#296 follow-up); the function exists so the SSE wiring is complete
// and the runners just need to call it once they're ready.
func (s *Server) PublishTestProgress(testID string, progress any) {
	if s.sseBroadcaster == nil {
		return
	}
	s.sseBroadcaster.Publish(sse.Frame{
		Type: "test_progress",
		Payload: map[string]any{
			"testId":   testID,
			"progress": progress,
		},
	})
}
