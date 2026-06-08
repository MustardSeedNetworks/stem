package api

import (
	"maps"
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

// handleBuildVersion serves GET /__version with build metadata for deployment
// validation. Unauthenticated by design — operators need to verify which
// binary is running without holding a session. Required by the Universal
// Build Contract (CLAUDE.md): all three sibling projects (seed/stem/niac)
// expose this endpoint with lowercase JSON keys: version, commit, buildTime,
// uiBuildHash.
//
// Stem additionally exposes `tlsFingerprint` — the SHA-256 fingerprint of
// the active TLS certificate, formatted as colon-separated uppercase hex
// pairs (the same format browsers show in the certificate dialog). The
// field is always present; in HTTP mode (or when the cert is not yet on
// disk) it is the empty string. This lets operators verify the cert
// installed via `stem install-ca` matches the one actually being served.
func (s *Server) handleBuildVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	info := version.Info()
	resp := make(map[string]string, len(info)+1)
	maps.Copy(resp, info)
	resp["tlsFingerprint"] = s.tlsFingerprintForResponse()
	writeJSON(w, resp)
}
