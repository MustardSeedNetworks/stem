// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"os"
	"reflect"
	"testing"
)

func newReflectorStateServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "reflectorstate")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectorstate123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	s.dataDir = dir
	return s, dir
}

// An operator configures the reflector once. Losing that on restart is what
// kept the reflector in a hand-written systemd unit running `stem reflect`
// (#1166): the daemon has to remember what it was told.
func TestReflectorConfigSurvivesARestart(t *testing.T) {
	first, dir := newReflectorStateServer(t)

	first.reflectorConfig = ReflectorConfig{
		Profile:    "netally",
		OUIFilter:  "00:c0:17",
		PortFilter: 3842,
		Interface:  "eth0",
		Autostart:  true,
	}
	if err := first.persistReflectorState(); err != nil {
		t.Fatalf("persistReflectorState: %v", err)
	}

	// A second server over the same data directory stands in for a restart.
	second, _ := newReflectorStateServer(t)
	second.dataDir = dir
	if err := second.loadReflectorState(); err != nil {
		t.Fatalf("loadReflectorState: %v", err)
	}

	got := second.reflectorConfig
	if got.Profile != "netally" || got.PortFilter != 3842 {
		t.Errorf("profile/port = %s/%d, want netally/3842", got.Profile, got.PortFilter)
	}
	if got.OUIFilter != "00:c0:17" {
		t.Errorf("ouiFilter = %q, want 00:c0:17", got.OUIFilter)
	}
	if got.Interface != "eth0" {
		t.Errorf("interface = %q, want eth0", got.Interface)
	}
	if !got.Autostart {
		t.Error("autostart was not restored")
	}
}

// A fresh install has no state file, and that is not an error — the daemon
// starts on its defaults.
func TestLoadReflectorStateWithoutAFile(t *testing.T) {
	s, _ := newReflectorStateServer(t)
	before := s.reflectorConfig

	if err := s.loadReflectorState(); err != nil {
		t.Fatalf("loadReflectorState on a fresh install: %v", err)
	}
	if !reflect.DeepEqual(s.reflectorConfig, before) {
		t.Errorf("config = %+v, want the defaults %+v", s.reflectorConfig, before)
	}
}

// A corrupt state file must not take the daemon down, and must not be read
// as "autostart a reflector on an interface nobody chose".
func TestLoadReflectorStateRejectsACorruptFile(t *testing.T) {
	s, dir := newReflectorStateServer(t)
	before := s.reflectorConfig

	if err := os.WriteFile(reflectorStatePath(dir), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := s.loadReflectorState(); err == nil {
		t.Error("loadReflectorState on a corrupt file = nil error, want a report")
	}
	if !reflect.DeepEqual(s.reflectorConfig, before) {
		t.Errorf("config = %+v, want the defaults left intact", s.reflectorConfig)
	}
	if s.reflectorConfig.Autostart {
		t.Error("a corrupt file enabled autostart")
	}
}

// Autostart only fires when an operator asked for it and named an interface;
// anything else would bring a dataplane up on a host nobody configured.
func TestShouldAutostartReflector(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg  ReflectorConfig
		want bool
	}{
		"enabled with an interface": {cfg: ReflectorConfig{Autostart: true, Interface: "eth0"}, want: true},
		"enabled without one":       {cfg: ReflectorConfig{Autostart: true}, want: false},
		"disabled with an interface": {
			cfg: ReflectorConfig{Interface: "eth0"}, want: false,
		},
		"neither": {cfg: ReflectorConfig{}, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tc.cfg.shouldAutostart(); got != tc.want {
				t.Errorf("shouldAutostart(%+v) = %v, want %v", tc.cfg, got, tc.want)
			}
		})
	}
}

// The state file carries no credential, but it is daemon state an operator
// should not be able to rewrite without privilege.
func TestPersistReflectorStateIsOwnerWritable(t *testing.T) {
	s, dir := newReflectorStateServer(t)
	if err := s.persistReflectorState(); err != nil {
		t.Fatalf("persistReflectorState: %v", err)
	}
	info, err := os.Stat(reflectorStatePath(dir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm&0o022 != 0 {
		t.Errorf("mode = %#o, want no group or world write", perm)
	}
}

// An unconfigured daemon must not touch the interface at boot. This is the
// guard that keeps autostart from being a surprise on every other host.
func TestAutostartReflectorDoesNothingWhenNotAskedFor(t *testing.T) {
	s, _ := newReflectorStateServer(t)
	s.reflectorConfig = ReflectorConfig{Profile: "netally", Interface: "eth0"} // no Autostart

	s.statsMu.RLock()
	before := s.testRunID
	s.statsMu.RUnlock()

	s.autostartReflector()

	// The run counter is the precise signal: only beginTestRun advances it,
	// so this distinguishes "never attempted" from "attempted and failed" —
	// which a status check alone cannot do.
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	if s.testRunID != before {
		t.Errorf("a run was begun (runID %d -> %d) without an autostart request", before, s.testRunID)
	}
	if s.currentTest != "" {
		t.Errorf("currentTest = %q, want none", s.currentTest)
	}
}

// Autostart asked for but impossible here (no privileged interface in a test
// process) must leave the daemon serving rather than wedged mid-run: the
// operator needs the API and UI up to see why the reflector did not start.
func TestAutostartReflectorFailureLeavesNoRunning(t *testing.T) {
	s, _ := newReflectorStateServer(t)
	s.reflectorConfig = ReflectorConfig{
		Profile:   "netally",
		Interface: "stem-nonexistent-if0",
		Autostart: true,
	}

	s.autostartReflector()

	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	if s.testStatus == statusRunning {
		t.Errorf("testStatus = %q after a failed autostart, want a non-running state", s.testStatus)
	}
}
