// SPDX-License-Identifier: BUSL-1.1

// Package update installs a newer stem release in place (ADR-0012).
//
// This file is the part the daemon shares with the `stem update` verb: the
// run lock that keeps an update and a measurement apart.
package update

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RunLockName is the advisory lock file in the daemon's data directory. The
// daemon holds it while any test or reflector run is active; `stem update`
// holds it from before it resolves a release until the new version is
// healthy. A run therefore never starts against a binary that is being
// replaced, and an update never starts under a run: stem is a measurement
// instrument, and a run that straddles two builds measured neither.
const RunLockName = "run.lock"

// ErrRunLockHeld reports that another process holds the run lock.
var ErrRunLockHeld = errors.New("the run lock is held by another process")

// RunGuard holds the run lock while at least one holder in this process
// needs it. The daemon needs the count: a cancelled run plan keeps its hold
// until its executor has actually stopped, which can be after the next run
// has started.
type RunGuard struct {
	path string

	mu      sync.Mutex
	file    *os.File
	holders int
}

// NewRunGuard returns a guard for the run lock in dir. It touches nothing
// until the first Acquire.
func NewRunGuard(dir string) *RunGuard {
	return &RunGuard{path: filepath.Join(dir, RunLockName)}
}

// Acquire adds a holder, taking the lock if this is the first. It never
// waits: it returns ErrRunLockHeld when another process holds the lock. The
// returned release may be called more than once; only the first call counts.
func (g *RunGuard) Acquire() (func(), error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.holders == 0 {
		file, openErr := os.OpenFile(g.path, os.O_RDWR|os.O_CREATE, 0o600)
		if openErr != nil {
			return nil, fmt.Errorf("open run lock: %w", openErr)
		}
		locked, lockErr := tryLock(file)
		if lockErr != nil {
			_ = file.Close()
			return nil, fmt.Errorf("lock %s: %w", g.path, lockErr)
		}
		if !locked {
			_ = file.Close()
			return nil, ErrRunLockHeld
		}
		g.file = file
	}
	g.holders++
	return sync.OnceFunc(g.release), nil
}

func (g *RunGuard) release() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.holders--
	if g.holders > 0 {
		return
	}
	_ = unlock(g.file)
	_ = g.file.Close()
	g.file = nil
}
