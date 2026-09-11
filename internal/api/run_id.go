// SPDX-License-Identifier: BUSL-1.1

package api

// Run identity. The CLI and the web UI are both clients of this daemon
// (#1166) and follow a run by the ID reported when it started, so that ID
// has to name exactly one run. A bare counter does not: it restarts at 1
// with the daemon, and a CLI reconnecting after a restart would accept a
// different run's progress and result as its own.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// runIDInstanceBytes is the random part identifying one daemon lifetime.
// Eight bytes make an accidental collision between two daemons — the only
// way an ID could repeat — not worth defending against further.
const runIDInstanceBytes = 8

// newRunInstanceID returns the per-daemon segment of every run ID it issues.
func newRunInstanceID() (string, error) {
	buf := make([]byte, runIDInstanceBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate run instance ID: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// newRunIDLocked issues the next run ID. Callers must hold statsMu; the
// counter it reads is the same one the run-generation guards compare.
func (s *Server) newRunIDLocked() string {
	s.testRunID++
	return fmt.Sprintf("stem-%s-%d", s.runInstanceID, s.testRunID)
}
