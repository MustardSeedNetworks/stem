// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/truststore"
)

const (
	// installCACommandName is the CLI subcommand name. Used for both
	// command dispatch and operator-facing hint messages so they stay
	// in sync.
	installCACommandName = "install-ca"

	// defaultCertPath is the file ensureSelfSignedCert (in internal/api)
	// writes when stem generates its own HTTPS certificate. install-ca
	// uses this path unless overridden with --cert.
	defaultCertPath = "certs/server.crt"

	// installCATimeoutSeconds is the timeout for individual platform
	// utility invocations (`security`, `update-ca-certificates`, etc.).
	installCATimeoutSeconds = 30
)

// installCAFlags holds the parsed command-line flags for install-ca.
type installCAFlags struct {
	certPath         string
	uninstall        bool
	printFingerprint bool
}

// parseInstallCAFlags extracts the install-ca command flags.
func parseInstallCAFlags(args []string) (installCAFlags, error) {
	fs := flag.NewFlagSet(installCACommandName, flag.ContinueOnError)
	certPath := fs.String("cert", defaultCertPath,
		"Path to the PEM-encoded certificate to install")
	uninstall := fs.Bool("uninstall", false,
		"Remove stem's certificate from the OS trust store")
	printFingerprint := fs.Bool("print-fingerprint", false,
		"Print the SHA-256 fingerprint of the certificate and exit without modifying the trust store")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, `Usage: stem install-ca [flags]

Install stem's self-signed root certificate into the operating system's
trust store so browsers stop showing the "not secure" warning when
visiting the stem UI over HTTPS.

The cert stem generates on first launch (at certs/server.crt) is a
single-tier self-signed root: it is both the leaf served on the TLS
handshake and its own issuer. Installing it as a trusted root tells the
OS to accept it for SSL.

Run 'stem web' at least once before install-ca so the certificate file
exists.

Supported platforms:
  macOS    System Keychain (requires sudo)
  Linux    System CA bundle via update-ca-certificates / update-ca-trust
  Windows  LocalMachine\Root (requires elevated shell)

Verification:
  stem install-ca --print-fingerprint
  curl -k https://localhost:8444/__version | jq -r .tlsFingerprint
The two values must match.

Flags:`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return installCAFlags{}, err
	}
	return installCAFlags{
		certPath:         *certPath,
		uninstall:        *uninstall,
		printFingerprint: *printFingerprint,
	}, nil
}

// resolveCertPath returns the absolute path to the cert file, or an error
// with operator-friendly hints if the file does not exist.
func resolveCertPath(p string) (string, error) {
	if p == "" {
		return "", errors.New("--cert must not be empty")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve cert path: %w", err)
	}
	_, statErr := os.Stat(abs)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", fmt.Errorf(
				"certificate not found at %s\n"+
					"Start 'stem web' once to generate its self-signed certificate, "+
					"then re-run install-ca", abs)
		}
		return "", fmt.Errorf("stat %s: %w", abs, statErr)
	}
	return abs, nil
}

// certFingerprint reads the cert at path and returns its SHA-256 fingerprint
// formatted as 32 uppercase hex pairs separated by colons.
func certFingerprint(path string) (string, error) {
	cert, err := truststore.ValidateCertFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cert.Raw)
	return formatColonHex(sum[:]), nil
}

// formatColonHex converts a byte slice into the standard browser-style
// fingerprint format (uppercase hex pairs separated by colons).
func formatColonHex(b []byte) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(b)*3-1)
	for i, x := range b {
		if i > 0 {
			out = append(out, ':')
		}
		out = append(out, hex[x>>4], hex[x&0x0f])
	}
	return string(out)
}

// installCACmd is the entry point for the `stem install-ca` subcommand.
func installCACmd(args []string) error {
	flags, err := parseInstallCAFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	abs, err := resolveCertPath(flags.certPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	fp, err := certFingerprint(abs)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if flags.printFingerprint {
		_, _ = fmt.Fprintln(os.Stdout, fp)
		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), installCATimeoutSeconds*time.Second)
	defer cancel()

	if flags.uninstall {
		return runUninstallCA(ctx, abs, fp)
	}
	return runInstallCAInstall(ctx, abs, fp)
}

func runInstallCAInstall(ctx context.Context, certPath, fingerprint string) error {
	_, _ = fmt.Fprintf(os.Stdout, "Installing %s into the OS trust store...\n", certPath)
	res, err := truststore.Install(ctx, certPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printInstallHints(err)
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "[ok] Installed.\n")
	printResult(res)
	_, _ = fmt.Fprintf(os.Stdout, "Certificate SHA-256 fingerprint:\n  %s\n", fingerprint)
	_, _ = fmt.Fprintf(os.Stdout, "Compare this against the value reported by:\n")
	_, _ = fmt.Fprintf(os.Stdout, "  curl -k https://localhost:8444/__version | jq -r .tlsFingerprint\n")
	return nil
}

func runUninstallCA(ctx context.Context, certPath, fingerprint string) error {
	_, _ = fmt.Fprintf(os.Stdout, "Removing %s from the OS trust store...\n", certPath)
	res, err := truststore.Uninstall(ctx, certPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "[ok] Removed.\n")
	printResult(res)
	_, _ = fmt.Fprintf(os.Stdout, "Certificate SHA-256 fingerprint:\n  %s\n", fingerprint)
	return nil
}

func printResult(res truststore.Result) {
	for _, s := range res.Stores {
		_, _ = fmt.Fprintf(os.Stdout, "  modified: %s\n", s)
	}
	for _, s := range res.Skipped {
		_, _ = fmt.Fprintf(os.Stdout, "  skipped:  %s\n", s)
	}
}

// printInstallHints adds operator-friendly guidance for common failure
// modes (missing sudo, unsupported platform).
func printInstallHints(err error) {
	if errors.Is(err, truststore.ErrUnsupportedPlatform) {
		_, _ = fmt.Fprintln(os.Stderr,
			"Hint: install-ca supports macOS, Linux, and Windows. "+
				"On other systems you must manually import the certificate "+
				"into your trust store.")
		return
	}
	uid := os.Getuid()
	if uid != 0 && uid != -1 {
		_, _ = fmt.Fprintln(os.Stderr,
			"Hint: modifying the system trust store usually requires root. "+
				"Try: sudo stem install-ca")
	}
}
