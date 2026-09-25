// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"path/filepath"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/httpserver"
)

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
