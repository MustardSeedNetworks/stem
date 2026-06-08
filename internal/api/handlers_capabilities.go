// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
)

// CapabilityInfo describes whether a single stem capability is
// available on this binary. When Supported is false, Reason carries a
// short, operator-facing explanation (e.g. "CGO + Linux required") so
// the UI can render a precise platform-guard banner without parsing
// raw error strings.
type CapabilityInfo struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

// CapabilitiesResponse is the payload returned by
// GET /api/v1/capabilities. It is unauthenticated by design so the
// frontend can gate UX before login. Mirrors the pattern of /__version
// (build metadata, also unauthenticated).
type CapabilitiesResponse struct {
	Reflector  CapabilityInfo `json:"reflector"`
	TestMaster CapabilityInfo `json:"testMaster"`
}

// handleCapabilities serves GET /api/v1/capabilities — a small,
// unauthenticated capability descriptor used by the frontend to
// surface platform-unsupported states (e.g. the macOS / Windows builds
// of stem ship without the CGO + Linux reflector dataplane).
//
// Test Master mode is supported on every platform stem builds for
// today, so its Supported flag is hard-coded true; if that changes the
// flag should be wired to a real probe in the same way the reflector
// flag delegates to [reflectorDP.Available].
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reflectorAvailable := reflectorDP.Available()
	resp := CapabilitiesResponse{
		Reflector: CapabilityInfo{
			Supported: reflectorAvailable,
		},
		TestMaster: CapabilityInfo{
			Supported: true,
		},
	}
	if !reflectorAvailable {
		resp.Reflector.Reason = reflectorDP.UnsupportedReason()
	}

	writeJSON(w, resp)
}
