// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

func newReflectorUpdateServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "reflectorupdate")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectorupdate123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// The autostart request arrives through the config endpoint, so it has to
// survive the commit that endpoint performs. The earlier test for this set
// s.reflectorConfig directly and so never exercised the path an operator
// actually uses — the fields were accepted and silently dropped, and the
// daemon never learned it should start the reflector at boot.
func TestReflectorConfigUpdateKeepsInterfaceAndAutostart(t *testing.T) {
	s := newReflectorUpdateServer(t)

	s.commitReflectorConfigUpdate(&ReflectorConfig{
		Profile:    "netally",
		PortFilter: 3842,
		Interface:  "eth0",
		Autostart:  true,
	})

	s.statsMu.RLock()
	got := s.reflectorConfig
	s.statsMu.RUnlock()

	if got.Interface != "eth0" {
		t.Errorf("interface = %q, want eth0 — the daemon cannot autostart without it", got.Interface)
	}
	if !got.Autostart {
		t.Error("autostart was dropped by the config update")
	}
	if got.Profile != "netally" || got.PortFilter != 3842 {
		t.Errorf("profile/port = %s/%d, want netally/3842", got.Profile, got.PortFilter)
	}
}

// An operator turning autostart back off must be able to: a field that only
// ever latches true cannot be switched off through the same endpoint.
func TestReflectorConfigUpdateCanDisableAutostart(t *testing.T) {
	s := newReflectorUpdateServer(t)

	s.commitReflectorConfigUpdate(&ReflectorConfig{Interface: "eth0", Autostart: true})
	s.commitReflectorConfigUpdate(&ReflectorConfig{Interface: "eth0", Autostart: false})

	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	if s.reflectorConfig.Autostart {
		t.Error("autostart could not be turned off")
	}
}

// The descriptor is read by a CLI running from wherever the operator
// happens to be, so a relative certificate path — which is what the daemon
// resolves against its own working directory — is unusable to them.
func TestPublishedCertificatePathIsAbsolute(t *testing.T) {
	s := newReflectorUpdateServer(t)
	s.tlsConfig.Enabled = true
	s.tlsConfig.CertFile = ""
	s.tlsConfig.CertsDir = "certs"

	got := s.publishedCertPath()
	if got == "" {
		t.Fatal("publishedCertPath() = \"\", want a path")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("publishedCertPath() = %q, want an absolute path", got)
	}
	if !strings.HasSuffix(got, filepath.Join("certs", "server.crt")) {
		t.Errorf("publishedCertPath() = %q, want it to end in certs/server.crt", got)
	}
}

// An already-absolute configured certificate is published unchanged.
func TestPublishedCertificatePathKeepsAnAbsoluteSetting(t *testing.T) {
	s := newReflectorUpdateServer(t)
	s.tlsConfig.Enabled = true
	s.tlsConfig.CertFile = "/etc/stem/tls/server.crt"

	if got := s.publishedCertPath(); got != "/etc/stem/tls/server.crt" {
		t.Errorf("publishedCertPath() = %q, want it unchanged", got)
	}
}

// The path has to be absolute in the descriptor a client actually reads,
// not merely available from a helper: this is the failure that stopped
// `stem reflect` working from an operator's shell on CT307.
func TestPublishedDescriptorCarriesAnAbsoluteCertificatePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "descriptorpath")
	t.Setenv("STEM_AUTH_PASSWORD", "descriptorpath123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	s.dataDir = dir
	s.tlsConfig.Enabled = true
	s.tlsConfig.CertFile = ""
	s.tlsConfig.CertsDir = "certs"

	if publishErr := s.publishConnection("https://localhost:8444"); publishErr != nil {
		t.Fatalf("publishConnection: %v", publishErr)
	}

	descriptor, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !filepath.IsAbs(descriptor.CAFile) {
		t.Errorf("published caFile = %q, want an absolute path a client can open from any directory", descriptor.CAFile)
	}
}
