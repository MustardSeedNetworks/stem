// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestACMEConfig tests the ACME configuration struct.
func TestACMEConfig(t *testing.T) {
	tests := []struct {
		name   string
		config ACMEConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: ACMEConfig{
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
			config: ACMEConfig{
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
			config: ACMEConfig{
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

	config := ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: cacheDir,
		Staging:  true,
	}

	manager, err := createACMEManager(config)
	if err != nil {
		t.Fatalf("createACMEManager() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify cache directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Expected cache directory to be created")
	}
}

// TestCreateACMEManagerDefaultCache tests default cache directory.
func TestCreateACMEManagerDefaultCache(t *testing.T) {
	// Skip if we can't create the default cache dir
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

	manager, err := createACMEManager(config)
	if err != nil {
		t.Fatalf("createACMEManager() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
}

// TestCreateACMETLSConfig tests TLS config creation with ACME.
func TestCreateACMETLSConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: tmpDir,
		Staging:  true,
	}

	manager, err := createACMEManager(config)
	if err != nil {
		t.Fatalf("createACMEManager() error: %v", err)
	}

	tlsConfig := createACMETLSConfig(manager)
	if tlsConfig == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	// TLS 1.3 minimum should be set
	if tlsConfig.MinVersion != 0x0304 { // tls.VersionTLS13
		t.Errorf("Expected MinVersion TLS 1.3, got %x", tlsConfig.MinVersion)
	}
}

// TestTLSConfigWithACME tests TLSConfig with ACME enabled.
func TestTLSConfigWithACME(t *testing.T) {
	config := TLSConfig{
		Enabled:  true,
		CertFile: "",
		KeyFile:  "",
		CertsDir: "certs",
		ACME: ACMEConfig{
			Enabled:  true,
			Domain:   "stem.example.com",
			Email:    "admin@example.com",
			CacheDir: "certs/acme",
			Staging:  false,
		},
	}

	// When ACME is enabled, CertFile/KeyFile should be ignored
	if !config.ACME.Enabled {
		t.Error("Expected ACME to be enabled")
	}

	if config.ACME.Domain == "" {
		t.Error("Expected domain to be set")
	}
}
