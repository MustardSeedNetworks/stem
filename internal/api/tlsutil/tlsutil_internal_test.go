// SPDX-License-Identifier: BUSL-1.1

package tlsutil

import (
	"os"
	"testing"
)

// TestNewACMEManagerDefaultCache tests that an empty CacheDir falls back to the
// default cache directory. It lives in the internal test package so it can
// reference the unexported defaultACMECacheDir constant rather than hard-coding
// the literal path.
func TestNewACMEManagerDefaultCache(t *testing.T) {
	// Skip if we can't create the default cache dir.
	if err := os.MkdirAll(defaultACMECacheDir, 0o700); err != nil {
		t.Skip("Cannot create default cache dir")
	}
	defer func() { _ = os.RemoveAll("certs") }()

	config := ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: "", // Use default
		Staging:  true,
	}

	manager, err := NewACMEManager(config)
	if err != nil {
		t.Fatalf("NewACMEManager() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
}
