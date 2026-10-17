// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"time"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

// runHasEnded reports whether a run in this state is over. The four terminal
// states are the ones the operator's history is made of; "starting" and
// "running" are a run still in flight, and "idle" is no run at all.
func runHasEnded(status string) bool {
	switch status {
	case statusCompleted, statusError, statusStopped, statusCancelled:
		return true
	default:
		return false
	}
}

// stampRunRecordLocked fills in the build that produced the result, when the
// run started and, once it has ended, when it finished and how long it took.
//
// stem updates in place (ADR-0012), so two results are comparable only when
// the builds behind them are known; the result names its own.
//
// stem is a measurement instrument, so the run's clock belongs to the side
// that ran it: before this the daemon tracked no per-run timing at all
// (`s.startTime` is the process), the UI was written against fields the wire
// never carried, and the History page could therefore never show a run
// (#1333). A browser stamping its own `completedAt` would have made the page
// fill up with a clock that measured nothing.
//
// The duration is milliseconds — the unit HistoryPage formats.
//
// Callers must hold s.statsMu.
func (s *Server) stampRunRecordLocked(result *TestResultResponse) {
	if result == nil {
		return
	}
	result.StemVersion = version.GetVersion()
	result.StemCommit = version.GetCommit()
	if s.runStartedAt.IsZero() {
		return
	}
	started := s.runStartedAt
	result.StartedAt = &started
	if !runHasEnded(result.Status) {
		return
	}
	completed := time.Now()
	elapsed := completed.Sub(started).Milliseconds()
	result.CompletedAt = &completed
	result.DurationMs = &elapsed
}
