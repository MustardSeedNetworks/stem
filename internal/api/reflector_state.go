// SPDX-License-Identifier: BUSL-1.1

package api

// Reflector state that outlives the process (#1166). The daemon owns the
// reflector now, so it has to remember what an operator configured and
// whether they asked for it at boot — that is what the hand-written
// `stem reflect` systemd unit used to encode.
//
// This is a file in the data directory, next to the licence and the daemon
// descriptor, rather than a row in internal/database: no daemon code opens
// the database today, and adding SQLite to startup for one small record
// would buy a migration path and a new failure mode for nothing.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MustardSeedNetworks/stem/internal/logging"
)

const (
	// reflectorStateFile holds the persisted reflector configuration.
	reflectorStateFile = "reflector.json"

	// reflectorStateMode keeps the file writable only by the daemon's own
	// account; it is daemon state, not operator-editable configuration.
	reflectorStateMode os.FileMode = 0o600
)

// reflectorStatePath returns the state file inside dataDir.
func reflectorStatePath(dataDir string) string {
	return filepath.Join(dataDir, reflectorStateFile)
}

// persistReflectorState writes the current reflector configuration.
func (s *Server) persistReflectorState() error {
	s.statsMu.RLock()
	cfg := s.reflectorConfig
	s.statsMu.RUnlock()

	encoded, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode reflector state: %w", err)
	}

	path := reflectorStatePath(s.dataDir)
	tmp, err := os.CreateTemp(s.dataDir, reflectorStateFile+"-*")
	if err != nil {
		return fmt.Errorf("create reflector state: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	if chmodErr := tmp.Chmod(reflectorStateMode); chmodErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("restrict reflector state: %w", chmodErr)
	}
	if _, writeErr := tmp.Write(encoded); writeErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("write reflector state: %w", writeErr)
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return fmt.Errorf("close reflector state: %w", closeErr)
	}
	if renameErr := os.Rename(tmpName, path); renameErr != nil {
		return fmt.Errorf("publish reflector state: %w", renameErr)
	}
	return nil
}

// loadReflectorState restores a previously persisted configuration. A file
// that is absent is a fresh install, not a failure. A file that cannot be
// read is reported and the defaults are left alone: a daemon must not infer
// "bring a dataplane up on some interface" from a damaged file.
func (s *Server) loadReflectorState() error {
	data, err := os.ReadFile(reflectorStatePath(s.dataDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read reflector state: %w", err)
	}

	var cfg ReflectorConfig
	if unmarshalErr := json.Unmarshal(data, &cfg); unmarshalErr != nil {
		return fmt.Errorf("decode reflector state: %w", unmarshalErr)
	}

	s.statsMu.Lock()
	s.reflectorConfig = cfg
	s.statsMu.Unlock()
	return nil
}

// autostartReflector brings the reflector up at boot when an operator asked
// for it. A failure is logged, not fatal: the daemon's API and UI must come
// up so the operator can see why the reflector did not.
func (s *Server) autostartReflector() {
	s.statsMu.RLock()
	cfg := s.reflectorConfig
	s.statsMu.RUnlock()

	if !cfg.shouldAutostart() {
		return
	}

	runID, err := s.beginTestRun(testTypeReflect, "reflector")
	if err != nil {
		logging.Error("reflector autostart could not reserve a run",
			"event", "reflector.autostart.failed", "error", err)
		return
	}
	if execErr := s.executeReflector(cfg.Interface, cfg.Profile); execErr != nil {
		s.statsMu.Lock()
		s.testStatus = statusError
		s.currentTest = ""
		s.statsMu.Unlock()
		logging.Error("reflector autostart failed",
			"event", "reflector.autostart.failed",
			"interface", cfg.Interface, "profile", cfg.Profile, "error", execErr)
		return
	}
	logging.Info("reflector started at boot",
		"event", "reflector.autostart",
		"suiteId", runID, "interface", cfg.Interface, "profile", cfg.Profile)
}

// initStateFromDataDir prepares the per-daemon state that depends on the
// data directory: what an operator configured before the last restart, and
// the run-ID segment identifying this daemon lifetime.
//
// A reflector state file that cannot be read is reported and the defaults
// stand — a damaged file must not stop the daemon serving, and must not be
// read as "bring a dataplane up on some interface". A run instance that
// cannot be generated is fatal: without it run IDs would repeat.
func (s *Server) initStateFromDataDir() error {
	if stateErr := s.loadReflectorState(); stateErr != nil {
		logging.Error("failed to restore reflector config; using defaults",
			"event", "reflector.state.unusable", "error", stateErr)
	}

	instanceID, err := newRunInstanceID()
	if err != nil {
		return err
	}
	s.runInstanceID = instanceID
	return nil
}
