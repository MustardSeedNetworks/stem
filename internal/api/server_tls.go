// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api/tlsutil"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// acmeReadHeaderTimeoutSec is the timeout for reading ACME HTTP-01 challenge
// request headers on the port-80 challenge server.
const acmeReadHeaderTimeoutSec = 10

// startTLS starts the server with TLS encryption on the already-bound
// listener. Priority order: ACME → manual certificates → self-signed.
func (s *Server) startTLS(ln net.Listener) error {
	// Priority 1: ACME/Let's Encrypt automatic certificates
	if s.tlsConfig.ACME.Enabled {
		if s.tlsConfig.ACME.Domain == "" {
			return errors.New("ACME enabled but no domain specified")
		}
		return s.startTLSWithACME(ln)
	}

	// Priority 2: Manual certificates from config
	certFile := s.tlsConfig.CertFile
	keyFile := s.tlsConfig.KeyFile

	// Priority 3: Self-signed certificate (fallback)
	if certFile == "" || keyFile == "" {
		var err error
		certFile, keyFile, err = tlsutil.EnsureSelfSignedCert(s.tlsConfig.CertsDir)
		if err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
	}

	// Configure TLS 1.3 minimum.
	s.httpServer.TLSConfig = tlsutil.ServerConfig()

	logging.Info("Starting HTTPS server",
		"addr", s.httpServer.Addr,
		"tls_version", "1.3",
		"cert_file", certFile,
	)

	listenErr := s.httpServer.ServeTLS(ln, certFile, keyFile)
	if listenErr != nil {
		return fmt.Errorf("serve TLS: %w", listenErr)
	}
	return nil
}

// startTLSWithACME starts the server with automatic Let's Encrypt certificates
// on the already-bound listener. Ported from Seed project for automatic
// certificate management.
func (s *Server) startTLSWithACME(ln net.Listener) error {
	manager, err := tlsutil.NewACMEManager(s.tlsConfig.ACME)
	if err != nil {
		return fmt.Errorf("create ACME manager: %w", err)
	}

	// Configure TLS with ACME
	s.httpServer.TLSConfig = tlsutil.ACMETLSConfig(manager)

	logging.Info("Starting HTTPS server with ACME",
		"addr", s.httpServer.Addr,
		"domain", s.tlsConfig.ACME.Domain)

	// Start HTTP-01 challenge handler on port 80
	// This is required for Let's Encrypt domain validation
	challengeServer := &http.Server{
		Addr:              ":80",
		Handler:           manager.HTTPHandler(nil),
		ReadHeaderTimeout: acmeReadHeaderTimeoutSec * time.Second,
	}
	s.acmeChallengeServer = challengeServer
	go func() {
		if listenErr := s.acmeChallengeServer.ListenAndServe(); listenErr != nil &&
			!errors.Is(listenErr, http.ErrServerClosed) {
			logging.Error("ACME challenge server error", "error", listenErr)
		}
	}()

	// ServeTLS with empty cert/key paths uses GetCertificate from TLSConfig.
	if listenErr := s.httpServer.ServeTLS(ln, "", ""); listenErr != nil {
		return fmt.Errorf("https server with ACME: %w", listenErr)
	}
	return nil
}

// activeCertPath returns the cert file path the server will use, or "" if the
// server is running in HTTP mode. Mirrors the priority order of startTLS so
// /__version reports the same cert that is actually served.
func (s *Server) activeCertPath() string {
	if !s.tlsConfig.Enabled {
		return ""
	}
	if s.tlsConfig.ACME.Enabled {
		// ACME certs live in the autocert cache; they are not a single
		// stable file path we can fingerprint here. Return empty rather
		// than guessing.
		return ""
	}
	if s.tlsConfig.CertFile != "" {
		return s.tlsConfig.CertFile
	}
	// Fall back to the self-signed default path used by EnsureSelfSignedCert.
	certsDir := s.tlsConfig.CertsDir
	if certsDir == "" {
		certsDir = tlsutil.DefaultCertsDir
	}
	return filepath.Join(certsDir, "server.crt")
}

// tlsFingerprintForResponse returns the cached fingerprint (computing it on
// first call). Errors are swallowed and reported as an empty string so
// /__version always returns a stable shape even if the cert is missing or
// unreadable; an empty value is a signal to the operator to investigate.
func (s *Server) tlsFingerprintForResponse() string {
	path := s.activeCertPath()
	if path == "" {
		return ""
	}
	fp, err := s.tlsFingerprint.Get(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(fp)
}
