// SPDX-License-Identifier: BUSL-1.1

package api

// server_license_status_internal_test.go covers the daemon half of #1312: the
// one place an operator learns that the licence file on disk is not one this
// build can stand behind is the startup log, because the entitlement
// consequence is silent — hasFeature grants the Free set either way. The
// `_internal_test.go` suffix is what the testpackage linter accepts for an
// internal-package test.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// licenseStartupEnv gives NewServer the credentials it requires and a HOME of
// this case's own, so the licence file under test is the one the daemon reads.
func licenseStartupEnv(t *testing.T) {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "licensetest")
	t.Setenv("STEM_AUTH_PASSWORD", "licensepass123")

	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config", "stem"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

// startupLog runs NewServer with the default logger writing into a buffer and
// returns what it wrote.
func startupLog(t *testing.T) string {
	t.Helper()
	var buf bytes.Buffer
	if err := logging.InitWithWriter(logging.DefaultConfig(), &buf); err != nil {
		t.Fatalf("InitWithWriter: %v", err)
	}
	t.Cleanup(logging.Reset)

	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })

	return buf.String()
}

// TestNewServerReportsAnUnusableLicenceFile: a licence file the process cannot
// read is named once at startup, with the status and the path to replace.
func TestNewServerReportsAnUnusableLicenceFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 file; the case cannot be built")
	}
	licenseStartupEnv(t)

	mgr, err := license.Load()
	if err != nil {
		t.Fatalf("license.Load: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}
	path := license.DefaultLicensePath()
	if chmodErr := os.Chmod(path, 0o000); chmodErr != nil {
		t.Fatalf("chmod: %v", chmodErr)
	}

	out := startupLog(t)

	for _, want := range []string{"license.unusable", "unreadable", path} {
		if !strings.Contains(out, want) {
			t.Errorf("startup log does not report the unusable licence (%q missing):\n%s", want, out)
		}
	}
}

// TestNewServerIsSilentOnAUsableLicence is the control: the event above is a
// signal, so a licence the daemon can use must not raise it.
func TestNewServerIsSilentOnAUsableLicence(t *testing.T) {
	licenseStartupEnv(t)

	mgr, err := license.Load()
	if err != nil {
		t.Fatalf("license.Load: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}

	if out := startupLog(t); strings.Contains(out, "license.unusable") {
		t.Errorf("startup log reports an unusable licence on an active trial:\n%s", out)
	}
}
