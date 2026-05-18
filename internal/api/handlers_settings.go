// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/netif"
)

// handleSettings handles settings get/update.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.statsMu.RLock()
		mode := s.mode
		iface := s.selectedIface
		s.statsMu.RUnlock()

		writeJSON(w, SettingsResponse{
			Mode:      mode,
			Interface: iface,
			Theme:     "system",
		})

	case http.MethodPost:
		var update SettingsUpdate
		if !decodeJSONStrict(w, r, &update) {
			return
		}

		if update.Interface != "" {
			// Validate that the interface exists.
			if !validateInterfaceExists(w, update.Interface) {
				return
			}

			s.statsMu.Lock()
			s.selectedIface = update.Interface
			s.statsMu.Unlock()
			logging.Info("interface selected", "interface", update.Interface)
		}

		writeJSON(w, StatusResponse{Status: "updated"})

	default:
		WriteMethodNotAllowed(w)
	}
}

// handleMode handles mode switching between reflector and test_master.
func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.statsMu.RLock()
		mode := s.mode
		s.statsMu.RUnlock()
		writeJSON(w, ModeResponse{Mode: mode})

	case http.MethodPost:
		var req ModeRequest
		if !decodeJSONStrict(w, r, &req) {
			return
		}

		if req.Mode != modeReflector && req.Mode != modeTestMaster {
			logging.Warn("mode update failed: invalid mode", "mode", req.Mode)
			WriteInvalidRequest(w, "Invalid mode (must be 'reflector' or 'test_master')")
			return
		}

		s.statsMu.Lock()
		oldMode := s.mode
		s.mode = req.Mode
		s.statsMu.Unlock()
		logging.Info("mode changed", "from", oldMode, "to", req.Mode)
		writeJSON(w, ModeUpdateResponse{Status: "updated", Mode: req.Mode})

	default:
		WriteMethodNotAllowed(w)
	}
}

// validateInterfaceExists checks if an interface exists and returns false if not (writing error to response).
func validateInterfaceExists(w http.ResponseWriter, ifaceName string) bool {
	ifaces, detectErr := netif.DetectInterfaces()
	if detectErr != nil {
		logging.Error("failed to detect interfaces for validation", "error", detectErr)
		WriteInternalError(w, detectErr)
		return false
	}

	for _, iface := range ifaces {
		if iface.Name == ifaceName {
			return true
		}
	}

	WriteInvalidRequest(w, "Specified network interface not found")
	return false
}
