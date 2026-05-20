// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"github.com/krisarmstrong/stem/internal/logging"
)

// TLS configuration constants.
const (
	rsaKeyBits          = 4096
	certValidYears      = 1
	defaultCertsDir     = "certs"
	defaultACMECacheDir = "certs/acme"
	serialNumberBitSize = 128
	refreshMultiplier   = 24 // Refresh token lasts 24x longer than access token

	// acmeReadHeaderTimeoutSec is the timeout for reading ACME challenge request headers.
	acmeReadHeaderTimeoutSec = 10
)

// ACMEConfig contains ACME/Let's Encrypt certificate settings.
// Ported from Seed project for automatic certificate management.
type ACMEConfig struct {
	// Enabled enables automatic certificate management via Let's Encrypt.
	Enabled bool

	// Domain is the domain name for the certificate (e.g., "stem.example.com").
	// Required when ACME is enabled.
	Domain string

	// Email is the contact email for Let's Encrypt notifications.
	Email string

	// CacheDir is the directory to cache certificates.
	// Defaults to "certs/acme" if empty.
	CacheDir string

	// Staging uses Let's Encrypt staging server (for testing).
	// Staging certificates are not trusted by browsers.
	Staging bool
}

// TLSConfig holds TLS configuration options.
type TLSConfig struct {
	// Enabled enables HTTPS mode.
	Enabled bool

	// CertFile is the path to the TLS certificate file.
	// If empty, a self-signed certificate will be generated.
	CertFile string

	// KeyFile is the path to the TLS private key file.
	// If empty, a self-signed certificate will be generated.
	KeyFile string

	// CertsDir is the directory for storing generated certificates.
	// Defaults to "certs".
	CertsDir string

	// ACME contains Let's Encrypt/ACME configuration.
	// When enabled, takes priority over CertFile/KeyFile and self-signed certs.
	ACME ACMEConfig
}

// DefaultTLSConfig returns secure TLS defaults with auto-generated certificates.
func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		Enabled:  true,
		CertFile: "",
		KeyFile:  "",
		CertsDir: defaultCertsDir,
	}
}

// createTLSConfig creates a [tls.Config] with secure settings.
// Uses TLS 1.3 minimum for best security.
func createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
}

// ensureSelfSignedCert generates a self-signed certificate if needed.
// Returns paths to the certificate and key files.
func ensureSelfSignedCert(certsDir string) (string, string, error) {
	if certsDir == "" {
		certsDir = defaultCertsDir
	}

	// Sanitize directory path to prevent directory traversal.
	certsDir = filepath.Clean(certsDir)

	certFile, keyFile := certPaths(certsDir)
	if existingCertFiles(certFile, keyFile) {
		logging.Info("Using existing TLS certificates", "cert", certFile, "key", keyFile)
		return certFile, keyFile, nil
	}

	// Ensure certs directory exists.
	if err := os.MkdirAll(certsDir, 0o700); err != nil {
		return "", "", fmt.Errorf("create certs directory: %w", err)
	}
	certRoot, err := os.OpenRoot(certsDir)
	if err != nil {
		return "", "", fmt.Errorf("open certs directory: %w", err)
	}
	defer func() { _ = certRoot.Close() }()

	// Generate private key with 4096-bit RSA.
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return "", "", fmt.Errorf("generate RSA key: %w", err)
	}

	serialNumber, err := newSerialNumber()
	if err != nil {
		return "", "", fmt.Errorf("generate serial number: %w", err)
	}

	template := newSelfSignedTemplate(serialNumber)
	certDER, err := createSelfSignedCert(&template, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("create certificate: %w", err)
	}

	// Write certificate.
	if writeErr := writeCertificate(certRoot, filepath.Base(certFile), certDER); writeErr != nil {
		return "", "", writeErr
	}

	// Write private key with restricted permissions.
	if writeErr := writePrivateKey(certRoot, filepath.Base(keyFile), privateKey); writeErr != nil {
		return "", "", writeErr
	}

	logging.Info("Generated self-signed TLS certificate",
		"cert_file", certFile,
		"key_file", keyFile,
		"valid_until", template.NotAfter.Format(time.RFC3339),
	)

	return certFile, keyFile, nil
}

func certPaths(certsDir string) (string, string) {
	return filepath.Join(certsDir, "server.crt"), filepath.Join(certsDir, "server.key")
}

func existingCertFiles(certFile, keyFile string) bool {
	if _, err := os.Stat(certFile); err != nil {
		return false
	}
	if _, err := os.Stat(keyFile); err != nil {
		return false
	}
	return true
}

func newSerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialNumberBitSize))
}

func newSelfSignedTemplate(serialNumber *big.Int) x509.Certificate {
	// The cert is a single-tier self-signed CA: it acts as both the root
	// (Issuer == Subject) and the leaf the TLS listener serves. This lets
	// `stem install-ca` install the same file into the OS trust store so
	// browsers stop showing the self-signed warning. Without IsCA=true and
	// KeyUsageCertSign, OS trust stores will reject the cert as not
	// eligible to act as a root.
	//
	// Existing certs on disk are not regenerated automatically; they will
	// continue to work for TLS but cannot be installed as roots until they
	// are deleted and stem regenerates them.
	return x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"The Stem"},
			CommonName:   "The Stem Self-Signed",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(certValidYears, 0, 0),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost", "stem.local"},
	}
}

func createSelfSignedCert(template *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
	return x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&privateKey.PublicKey,
		privateKey,
	)
}

func writeCertificate(root *os.Root, certFile string, certDER []byte) error {
	certOut, err := root.Create(certFile)
	if err != nil {
		return fmt.Errorf("create cert file: %w", err)
	}
	defer func() { _ = certOut.Close() }()

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}
	if encodeErr := pem.Encode(certOut, certBlock); encodeErr != nil {
		return fmt.Errorf("encode certificate PEM: %w", encodeErr)
	}
	return nil
}

func writePrivateKey(root *os.Root, keyFile string, privateKey *rsa.PrivateKey) error {
	keyOut, err := root.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer func() { _ = keyOut.Close() }()

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if encodeErr := pem.Encode(keyOut, keyBlock); encodeErr != nil {
		return fmt.Errorf("encode private key PEM: %w", encodeErr)
	}
	return nil
}

// createACMEManager creates an autocert.Manager for Let's Encrypt certificate management.
// Ported from Seed project for automatic certificate management.
func createACMEManager(config ACMEConfig) (*autocert.Manager, error) {
	cacheDir := config.CacheDir
	if cacheDir == "" {
		cacheDir = defaultACMECacheDir
	}

	// Ensure cache directory exists with secure permissions
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("create ACME cache dir: %w", err)
	}

	manager := &autocert.Manager{}
	manager.Prompt = autocert.AcceptTOS
	manager.HostPolicy = autocert.HostWhitelist(config.Domain)
	manager.Cache = autocert.DirCache(cacheDir)
	manager.Email = config.Email

	// Use Let's Encrypt staging server for testing (certs won't be trusted by browsers)
	if config.Staging {
		client := &acme.Client{}
		client.DirectoryURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
		manager.Client = client
		logging.Warn("ACME: Using Let's Encrypt STAGING server (certificates will not be trusted)")
	}

	return manager, nil
}

// createACMETLSConfig creates a TLS config using the ACME manager.
func createACMETLSConfig(manager *autocert.Manager) *tls.Config {
	tlsConfig := manager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS13
	return tlsConfig
}
