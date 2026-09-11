// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// newTestServer creates a test server with automatic cleanup.
// This helper ensures that all servers created in tests are properly
// shut down to prevent goroutine leaks.
func newTestServer(t testing.TB) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir()) // Use fast bcrypt for tests
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	// Entitlement is a test input: install a Professional trial so these tests
	// exercise handler behaviour rather than whatever activation state the
	// developer's ~/.config/stem happens to hold.
	mgr, _, mgrErr := license.LoadFromDir(t.TempDir())
	if mgrErr != nil {
		t.Fatalf("LoadFromDir() error: %v", mgrErr)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}
	s.licenseManager = mgr
	return s
}
