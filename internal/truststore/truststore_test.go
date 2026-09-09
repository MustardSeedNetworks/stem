// SPDX-License-Identifier: BUSL-1.1

// Lifted from the seed project (internal/truststore); keep in sync.

package truststore_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/truststore"
)

// makeFixtureCert writes a self-signed certificate PEM to a temp file and
// returns the path. The cert carries IsCA=true and KeyUsageCertSign so it
// looks like the cert stem generates for HTTPS.
func makeFixtureCert(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(7),
		Subject:               pkix.Name{CommonName: "stem-truststore-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	var buf bytes.Buffer
	if encErr := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: der}); encErr != nil {
		t.Fatalf("pem encode: %v", encErr)
	}
	path := filepath.Join(t.TempDir(), "stem-test.crt")
	if writeErr := os.WriteFile(path, buf.Bytes(), 0o600); writeErr != nil {
		t.Fatalf("write fixture: %v", writeErr)
	}
	return path
}

func TestValidateCertFile_OK(t *testing.T) {
	t.Parallel()
	path := makeFixtureCert(t)
	cert, err := truststore.ValidateCertFile(path)
	if err != nil {
		t.Fatalf("ValidateCertFile: %v", err)
	}
	if cert == nil {
		t.Fatal("ValidateCertFile returned nil cert without error")
	}
	if cert.Subject.CommonName != "stem-truststore-test" {
		t.Errorf("unexpected subject CN: %q", cert.Subject.CommonName)
	}
}

func TestValidateCertFile_EmptyPath(t *testing.T) {
	t.Parallel()
	_, err := truststore.ValidateCertFile("")
	if !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("expected ErrInvalidCertificate, got %v", err)
	}
}

func TestValidateCertFile_NotPEM(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-cert.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, err := truststore.ValidateCertFile(path)
	if !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("expected ErrInvalidCertificate, got %v", err)
	}
}

func TestValidateCertFile_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := truststore.ValidateCertFile(filepath.Join(t.TempDir(), "does-not-exist.crt"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// Install and Uninstall validate the certificate before touching the host
// trust store. These cases stop at that guard: none of them reaches
// installPlatform, which would shell out to the real system tooling.
func TestInstall_RejectsNonCertificate(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "not-a-cert.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	res, err := truststore.Install(t.Context(), path)
	if !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("expected ErrInvalidCertificate, got %v", err)
	}
	if len(res.Stores) != 0 || len(res.Skipped) != 0 {
		t.Errorf("a rejected certificate reported trust stores: %+v", res)
	}
}

func TestInstall_RejectsMissingFile(t *testing.T) {
	t.Parallel()
	_, err := truststore.Install(t.Context(), filepath.Join(t.TempDir(), "absent.crt"))
	if err == nil {
		t.Fatal("expected an error for a missing certificate file, got nil")
	}
	if errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("a missing file should report the read error, not ErrInvalidCertificate: %v", err)
	}
}

func TestUninstall_RejectsNonCertificate(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "not-a-cert.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	res, err := truststore.Uninstall(t.Context(), path)
	if !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("expected ErrInvalidCertificate, got %v", err)
	}
	if len(res.Stores) != 0 || len(res.Skipped) != 0 {
		t.Errorf("a rejected certificate reported trust stores: %+v", res)
	}
}

func TestUninstall_RejectsEmptyPath(t *testing.T) {
	t.Parallel()
	_, err := truststore.Uninstall(t.Context(), "")
	if !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("expected ErrInvalidCertificate, got %v", err)
	}
}
