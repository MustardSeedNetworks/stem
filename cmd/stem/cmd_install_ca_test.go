// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/truststore"
)

// writeSelfSignedCert writes a PEM certificate of the shape `stem web`
// generates (a single-tier self-signed root) into dir and returns its path
// alongside the DER bytes the fingerprint is taken over.
func writeSelfSignedCert(t *testing.T, dir string) (string, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "stem-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	path := filepath.Join(dir, "server.crt")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if writeErr := os.WriteFile(path, pemBytes, 0o600); writeErr != nil {
		t.Fatalf("write certificate: %v", writeErr)
	}
	return path, der
}

func TestParseInstallCAFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want installCAFlags
	}{
		{
			name: "defaults",
			args: nil,
			want: installCAFlags{certPath: defaultCertPath},
		},
		{
			name: "explicit cert path",
			args: []string{"--cert", "/tmp/other.crt"},
			want: installCAFlags{certPath: "/tmp/other.crt"},
		},
		{
			name: "uninstall",
			args: []string{"--uninstall"},
			want: installCAFlags{certPath: defaultCertPath, uninstall: true},
		},
		{
			name: "print fingerprint",
			args: []string{"--print-fingerprint"},
			want: installCAFlags{certPath: defaultCertPath, printFingerprint: true},
		},
		{
			name: "all three together",
			args: []string{"--cert", "a.crt", "--uninstall", "--print-fingerprint"},
			want: installCAFlags{certPath: "a.crt", uninstall: true, printFingerprint: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInstallCAFlags(tt.args)
			if err != nil {
				t.Fatalf("parseInstallCAFlags(%v) error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("parseInstallCAFlags(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

// TestParseInstallCAFlagsRejectsUnknownFlags: install-ca parses with
// ContinueOnError so a typo returns to the caller instead of exiting.
func TestParseInstallCAFlagsRejectsUnknownFlags(t *testing.T) {
	if _, err := parseInstallCAFlags([]string{"--not-a-flag"}); err == nil {
		t.Error("parseInstallCAFlags(--not-a-flag) = nil error, want a parse error")
	}
}

func TestParseInstallCAFlagsHelpIsNotAnError(t *testing.T) {
	_, err := parseInstallCAFlags([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("parseInstallCAFlags(-h) error = %v, want flag.ErrHelp", err)
	}
}

func TestResolveCertPath(t *testing.T) {
	dir := t.TempDir()
	path, _ := writeSelfSignedCert(t, dir)

	abs, err := resolveCertPath(path)
	if err != nil {
		t.Fatalf("resolveCertPath(%q) error: %v", path, err)
	}
	if !filepath.IsAbs(abs) {
		t.Errorf("resolveCertPath returned %q, which is not absolute", abs)
	}
}

func TestResolveCertPathRejectsEmpty(t *testing.T) {
	if _, err := resolveCertPath(""); err == nil {
		t.Error("resolveCertPath(\"\") = nil error, want a refusal")
	}
}

// TestResolveCertPathMissingFileTellsTheOperatorWhatToRun: the file is absent
// on every host that has not started the web UI yet, so this is the common
// case and the message has to name the fix.
func TestResolveCertPathMissingFileTellsTheOperatorWhatToRun(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "server.crt")

	_, err := resolveCertPath(missing)
	if err == nil {
		t.Fatal("resolveCertPath on a missing file = nil error")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("error %q does not name the path it looked at", err)
	}
	if !strings.Contains(err.Error(), "stem web") {
		t.Errorf("error %q does not tell the operator to run 'stem web'", err)
	}
}

// TestCertFingerprintMatchesTheCertificateDER pins the value the operator is
// told to compare against /__version's tlsFingerprint.
func TestCertFingerprintMatchesTheCertificateDER(t *testing.T) {
	path, der := writeSelfSignedCert(t, t.TempDir())

	got, err := certFingerprint(path)
	if err != nil {
		t.Fatalf("certFingerprint: %v", err)
	}

	sum := sha256.Sum256(der)
	if want := formatColonHex(sum[:]); got != want {
		t.Errorf("certFingerprint = %q, want %q", got, want)
	}
}

func TestCertFingerprintRejectsNonCertificates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.crt")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := certFingerprint(path); !errors.Is(err, truststore.ErrInvalidCertificate) {
		t.Errorf("certFingerprint on junk = %v, want ErrInvalidCertificate", err)
	}
}

func TestFormatColonHex(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{"single byte", []byte{0x0a}, "0A"},
		{"two bytes", []byte{0x00, 0xff}, "00:FF"},
		{"nibble order", []byte{0xab, 0xcd}, "AB:CD"},
		{"four bytes", []byte{0xde, 0xad, 0xbe, 0xef}, "DE:AD:BE:EF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatColonHex(tt.in); got != tt.want {
				t.Errorf("formatColonHex(%x) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestFormatColonHexLengthForASHA256 is the shape the operator compares by
// eye: 32 uppercase pairs, 31 separators.
func TestFormatColonHexLengthForASHA256(t *testing.T) {
	sum := sha256.Sum256([]byte("stem"))

	got := formatColonHex(sum[:])
	if want := 32*2 + 31; len(got) != want {
		t.Errorf("formatColonHex(sha256) length = %d, want %d", len(got), want)
	}
	if strings.Count(got, ":") != 31 {
		t.Errorf("formatColonHex(sha256) = %q, want 31 colons", got)
	}
}

// TestInstallCACmdPrintFingerprintDoesNotTouchTheTrustStore: --print-fingerprint
// is documented as the read-only verification path, so it must return before
// truststore.Install is reached (which would need root).
func TestInstallCACmdPrintFingerprintDoesNotTouchTheTrustStore(t *testing.T) {
	path, der := writeSelfSignedCert(t, t.TempDir())

	var err error
	out := captureStdout(t, func() {
		err = installCACmd([]string{"--cert", path, "--print-fingerprint"})
	})
	if err != nil {
		t.Fatalf("installCACmd --print-fingerprint: %v", err)
	}

	sum := sha256.Sum256(der)
	if want := formatColonHex(sum[:]) + "\n"; out != want {
		t.Errorf("installCACmd printed %q, want exactly %q", out, want)
	}
}

func TestInstallCACmdReportsAMissingCertificate(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "server.crt")

	if err := installCACmd([]string{"--cert", missing, "--print-fingerprint"}); err == nil {
		t.Error("installCACmd on a missing certificate = nil error")
	}
}

func TestPrintResult(t *testing.T) {
	out := captureStdout(t, func() {
		printResult(truststore.Result{
			Stores:  []string{"System Keychain"},
			Skipped: []string{"Firefox NSS"},
		})
	})

	if !strings.Contains(out, "modified: System Keychain") {
		t.Errorf("printResult output %q omits the modified store", out)
	}
	if !strings.Contains(out, "skipped:  Firefox NSS") {
		t.Errorf("printResult output %q omits the skipped store", out)
	}
}
