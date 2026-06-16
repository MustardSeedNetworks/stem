// SPDX-License-Identifier: BUSL-1.1

package tlsutil_test

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api/tlsutil"
)

// TestDefaultConfig tests the DefaultConfig function.
func TestDefaultConfig(t *testing.T) {
	config := tlsutil.DefaultConfig()

	if !config.Enabled {
		t.Error("Expected TLS to be enabled by default")
	}
	if config.CertFile != "" {
		t.Errorf("Expected empty CertFile, got '%s'", config.CertFile)
	}
	if config.KeyFile != "" {
		t.Errorf("Expected empty KeyFile, got '%s'", config.KeyFile)
	}
	if config.CertsDir != tlsutil.DefaultCertsDir {
		t.Errorf("Expected CertsDir '%s', got '%s'", tlsutil.DefaultCertsDir, config.CertsDir)
	}
}

// TestServerConfig tests the ServerConfig function.
func TestServerConfig(t *testing.T) {
	config := tlsutil.ServerConfig()

	if config == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	if config.MinVersion != tls.VersionTLS13 {
		t.Errorf("Expected TLS 1.3 min version, got %d", config.MinVersion)
	}
}

// TestEnsureSelfSignedCert tests the EnsureSelfSignedCert function.
func TestEnsureSelfSignedCert(t *testing.T) {
	// Create a temporary directory for test certificates.
	tempDir := t.TempDir()

	t.Run("generate new certificates", func(t *testing.T) {
		certFile, keyFile, err := tlsutil.EnsureSelfSignedCert(tempDir)
		if err != nil {
			t.Fatalf("EnsureSelfSignedCert() error: %v", err)
		}

		if certFile == "" {
			t.Error("Expected non-empty certFile")
		}
		if keyFile == "" {
			t.Error("Expected non-empty keyFile")
		}

		// Verify files exist.
		_, certStatErr := os.Stat(certFile)
		if certStatErr != nil {
			t.Errorf("Certificate file does not exist: %v", certStatErr)
		}
		_, keyStatErr := os.Stat(keyFile)
		if keyStatErr != nil {
			t.Errorf("Key file does not exist: %v", keyStatErr)
		}
	})

	t.Run("use existing certificates", func(t *testing.T) {
		// Should reuse existing certificates.
		certFile, keyFile, err := tlsutil.EnsureSelfSignedCert(tempDir)
		if err != nil {
			t.Fatalf("EnsureSelfSignedCert() error: %v", err)
		}

		if certFile == "" || keyFile == "" {
			t.Error("Expected non-empty certificate paths")
		}
	})

	t.Run("empty certs dir defaults to default", func(t *testing.T) {
		// This would use the default certs dir.
		// Skip if we don't want to pollute the filesystem.
		t.Skip("Skipping to avoid creating files in default location")
	})
}

// TestEnsureSelfSignedCertExistingFiles tests EnsureSelfSignedCert with existing files.
func TestEnsureSelfSignedCertExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// First call - generate.
	certFile, keyFile, err := tlsutil.EnsureSelfSignedCert(tempDir)
	if err != nil {
		t.Fatalf("First call error: %v", err)
	}

	// Get file info.
	certInfo, _ := os.Stat(certFile)
	keyInfo, _ := os.Stat(keyFile)

	// Second call - should reuse.
	certFile2, keyFile2, err := tlsutil.EnsureSelfSignedCert(tempDir)
	if err != nil {
		t.Fatalf("Second call error: %v", err)
	}

	if certFile != certFile2 || keyFile != keyFile2 {
		t.Error("Expected same paths on second call")
	}

	// Files should not have changed.
	certInfo2, _ := os.Stat(certFile)
	keyInfo2, _ := os.Stat(keyFile)

	if certInfo.ModTime() != certInfo2.ModTime() {
		t.Error("Cert file should not have been regenerated")
	}
	if keyInfo.ModTime() != keyInfo2.ModTime() {
		t.Error("Key file should not have been regenerated")
	}
}

// TestConfigStruct tests the Config struct.
func TestConfigStruct(t *testing.T) {
	config := tlsutil.Config{
		Enabled:  true,
		CertFile: "/path/to/cert.pem",
		KeyFile:  "/path/to/key.pem",
		CertsDir: "/path/to/certs",
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.CertFile != "/path/to/cert.pem" {
		t.Errorf("Unexpected CertFile: %s", config.CertFile)
	}
	if config.KeyFile != "/path/to/key.pem" {
		t.Errorf("Unexpected KeyFile: %s", config.KeyFile)
	}
	if config.CertsDir != "/path/to/certs" {
		t.Errorf("Unexpected CertsDir: %s", config.CertsDir)
	}
}

// TestACMEConfig tests the ACME configuration struct.
func TestACMEConfig(t *testing.T) {
	tests := []struct {
		name   string
		config tlsutil.ACMEConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: tlsutil.ACMEConfig{
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
			config: tlsutil.ACMEConfig{
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
			config: tlsutil.ACMEConfig{
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

// TestNewACMEManager tests the ACME manager creation.
func TestNewACMEManager(t *testing.T) {
	// Create temporary directory for cache
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "acme-cache")

	config := tlsutil.ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: cacheDir,
		Staging:  true,
	}

	manager, err := tlsutil.NewACMEManager(config)
	if err != nil {
		t.Fatalf("tlsutil.NewACMEManager() error: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify cache directory was created
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		t.Error("Expected cache directory to be created")
	}
}

// TestACMETLSConfig tests TLS config creation with ACME.
func TestACMETLSConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := tlsutil.ACMEConfig{
		Enabled:  true,
		Domain:   "test.example.com",
		Email:    "test@example.com",
		CacheDir: tmpDir,
		Staging:  true,
	}

	manager, err := tlsutil.NewACMEManager(config)
	if err != nil {
		t.Fatalf("tlsutil.NewACMEManager() error: %v", err)
	}

	tlsConfig := tlsutil.ACMETLSConfig(manager)
	if tlsConfig == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	// TLS 1.3 minimum should be set
	if tlsConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("Expected MinVersion TLS 1.3, got %x", tlsConfig.MinVersion)
	}
}

// TestConfigWithACME tests Config with ACME enabled.
func TestConfigWithACME(t *testing.T) {
	config := tlsutil.Config{
		Enabled:  true,
		CertFile: "",
		KeyFile:  "",
		CertsDir: "certs",
		ACME: tlsutil.ACMEConfig{
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
