// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// TestACMEConfig tests the ACME configuration struct.
func TestACMEConfig(t *testing.T) {
	tests := []struct {
		name   string
		config api.ACMEConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: api.ACMEConfig{
				Enabled:  true,
				Domain:   "stem.example.com",
				Email:    "admin@example.com",
				CacheDir: "certs/acme",
				Staging:  false,
			},
			valid: true,
		},
		{
			name: "staging mode",
			config: api.ACMEConfig{
				Enabled:  true,
				Domain:   "test.example.com",
				Email:    "test@example.com",
				CacheDir: "",
				Staging:  true,
			},
			valid: true,
		},
		{
			name: "disabled config",
			config: api.ACMEConfig{
				Enabled: false,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Enabled && tt.config.Domain == "" {
				t.Error("Expected domain to be required when ACME is enabled")
			}
		})
	}
}

// TestCreateACMEManager tests the ACME manager creation.
func TestCreateACMEManager(t *testing.T) {
	// Create temporary directory for cache
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "acme-cache")

	config := api.ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: cacheDir,
		Staging:  true,
	}

	manager, err := api.CreateACMEManagerForTest(config)
	if err != nil {
		t.Fatalf("api.CreateACMEManagerForTest() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify cache directory was created
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		t.Error("Expected cache directory to be created")
	}
}

// TestCreateACMEManagerDefaultCache tests default cache directory.
func TestCreateACMEManagerDefaultCache(t *testing.T) {
	// Skip if we can't create the default cache dir
	if err := os.MkdirAll(api.DefaultACMECacheDirForTest(), 0o700); err != nil {
		t.Skip("Cannot create default cache dir")
	}
	defer func() { _ = os.RemoveAll("certs") }()

	config := api.ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: "", // Use default
		Staging:  true,
	}

	manager, err := api.CreateACMEManagerForTest(config)
	if err != nil {
		t.Fatalf("api.CreateACMEManagerForTest() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
}

// TestCreateACMETLSConfig tests TLS config creation with ACME.
func TestCreateACMETLSConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := api.ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: tmpDir,
		Staging:  true,
	}

	manager, err := api.CreateACMEManagerForTest(config)
	if err != nil {
		t.Fatalf("api.CreateACMEManagerForTest() error: %v", err)
	}

	tlsConfig := api.CreateACMETLSConfigForTest(manager)
	if tlsConfig == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	// TLS 1.3 minimum should be set
	if tlsConfig.MinVersion != 0x0304 { // tls.VersionTLS13
		t.Errorf("Expected MinVersion TLS 1.3, got %x", tlsConfig.MinVersion)
	}
}

// TestTLSConfigWithACME tests api.TLSConfig with ACME enabled.
func TestTLSConfigWithACME(t *testing.T) {
	config := api.TLSConfig{
		Enabled:  true,
		CertFile: "",
		KeyFile:  "",
		CertsDir: "certs",
		ACME: api.ACMEConfig{
			Enabled:  true,
			Domain:   "stem.example.com",
			Email:    "admin@example.com",
			CacheDir: "certs/acme",
			Staging:  false,
		},
	}

	// When ACME is enabled, CertFile/KeyFile should be ignored
	if !config.Enabled {
		t.Error("Expected TLS to be enabled")
	}
	if config.CertFile != "" {
		t.Errorf("CertFile = %q, want empty", config.CertFile)
	}
	if config.KeyFile != "" {
		t.Errorf("KeyFile = %q, want empty", config.KeyFile)
	}
	if config.CertsDir != "certs" {
		t.Errorf("CertsDir = %q, want %q", config.CertsDir, "certs")
	}
	if !config.ACME.Enabled {
		t.Error("Expected ACME to be enabled")
	}

	if config.ACME.Domain == "" {
		t.Error("Expected domain to be set")
	}
}
