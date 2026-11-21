// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/httpserver"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// resolveCertDir is the certificate directory as an absolute path:
// STEM_TLS_CERTS_DIR, else certs under the working directory. foundation's
// listener checks the directory with securedir, which never resolves a
// relative path and so refuses one.
func resolveCertDir() (string, error) {
	dir := os.Getenv("STEM_TLS_CERTS_DIR")
	if dir == "" {
		dir = "certs"
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve the certificate directory %q: %w", dir, err)
	}
	return abs, nil
}

// initTransport sets where the daemon's one HTTPS listener binds, which
// certificate it serves, and how its cookies and cross-origin policy behave.
func (s *Server) initTransport(listen ListenAddr) error {
	certDir, err := resolveCertDir()
	if err != nil {
		return err
	}
	s.cookieConfig = auth.DefaultCookieConfig()
	s.corsAllowPrivate = corsAllowPrivateEnabled()
	s.listenConfig = httpserver.Config{
		Addr:     net.JoinHostPort(listen.Host, strconv.Itoa(listen.Port)),
		Explicit: listen.Explicit,
		CertFile: os.Getenv("STEM_TLS_CERT"),
		KeyFile:  os.Getenv("STEM_TLS_KEY"),
		CertDir:  certDir,
		Cert: httpserver.CertOptions{
			CommonName: "The Stem Self-Signed",
			DNSNames:   []string{"localhost", "stem.local"},
		},
		Logger: logging.Get(),
	}
	return nil
}

// activeCertPath returns the certificate file the listener serves: the
// operator's STEM_TLS_CERT when both it and STEM_TLS_KEY are set, otherwise
// the self-signed pair httpserver.Listen keeps in the certs directory.
// It mirrors httpserver's own choice so /__version fingerprints the
// certificate that is actually served.
func (s *Server) activeCertPath() string {
	if s.listenConfig.CertFile != "" && s.listenConfig.KeyFile != "" {
		return s.listenConfig.CertFile
	}
	dir := s.listenConfig.CertDir
	if dir == "" {
		dir = "certs"
	}
	return filepath.Join(dir, httpserver.DefaultCertFileName)
}

// tlsFingerprintForResponse returns the cached fingerprint (computing it on
// first call). Errors are swallowed and reported as an empty string so
// /__version always returns a stable shape even if the cert is missing or
// unreadable; an empty value is a signal to the operator to investigate.
func (s *Server) tlsFingerprintForResponse() string {
	fp, err := s.tlsFingerprint.Get(s.activeCertPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(fp)
}
