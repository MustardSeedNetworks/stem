// SPDX-License-Identifier: BUSL-1.1

package tlsutil_test

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/httpserver"

	"github.com/MustardSeedNetworks/stem/internal/api/tlsutil"
)

// generatedPair is a self-signed pair on disk and the certificate's
// fingerprint computed independently of tlsutil.
type generatedPair struct {
	dir, certPath, want string
}

// generated writes the self-signed pair the listener would.
func generated(t *testing.T) generatedPair {
	t.Helper()
	dir := t.TempDir()
	certPath := filepath.Join(dir, httpserver.DefaultCertFileName)
	if _, err := httpserver.EnsureCertificate(nil, certPath, filepath.Join(dir, httpserver.DefaultKeyFileName),
		httpserver.CertOptions{}); err != nil {
		t.Fatalf("EnsureCertificate: %v", err)
	}
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(certPEM)
	sum := sha256.Sum256(block.Bytes)
	pairs := make([]string, 0, len(sum))
	for _, b := range sum {
		pairs = append(pairs, fmt.Sprintf("%02X", b))
	}
	return generatedPair{dir: dir, certPath: certPath, want: strings.Join(pairs, ":")}
}

// TestFingerprintIsTheCertificateSHA256: /__version's tlsFingerprint is the
// SHA-256 of the certificate the listener generated, colon-separated, the
// form a browser's certificate dialog shows.
func TestFingerprintIsTheCertificateSHA256(t *testing.T) {
	pair := generated(t)
	var cache tlsutil.FingerprintCache
	if got, err := cache.Get(pair.certPath); err != nil || got != pair.want {
		t.Fatalf("Get = %q, %v; want %q", got, err, pair.want)
	}
}

// TestFingerprintIsCached: /__version does not re-read the file per request.
func TestFingerprintIsCached(t *testing.T) {
	pair := generated(t)
	var cache tlsutil.FingerprintCache
	if _, err := cache.Get(pair.certPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pair.certPath); err != nil {
		t.Fatal(err)
	}
	if got, err := cache.Get(pair.certPath); err != nil || got != pair.want {
		t.Errorf("cached Get = %q, %v; want %q", got, err, pair.want)
	}
}

// TestFingerprintRefusesWhatIsNotACertificate: a file holding no CERTIFICATE
// block, or no file at all, is an error rather than a fingerprint of
// something else.
func TestFingerprintRefusesWhatIsNotACertificate(t *testing.T) {
	dir := generated(t).dir
	var cache tlsutil.FingerprintCache
	for _, path := range []string{
		filepath.Join(dir, httpserver.DefaultKeyFileName),
		filepath.Join(dir, "absent.crt"),
	} {
		if fp, err := cache.Get(path); err == nil {
			t.Errorf("%s fingerprinted as %q", filepath.Base(path), fp)
		}
	}
}
