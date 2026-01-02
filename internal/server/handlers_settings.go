// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/netif"
)

// handleSettings handles settings get/update.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, SettingsResponse{
			Mode:      s.mode,
			Interface: s.selectedIface,
			Theme:     "system",
		})

	case http.MethodPost:
		var update SettingsUpdate
		decodeErr := json.NewDecoder(r.Body).Decode(&update)
		if decodeErr != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if update.Interface != "" {
			// Validate that the interface exists.
			ifaces, detectErr := netif.DetectInterfaces()
			if detectErr != nil {
				logging.Error("failed to detect interfaces for validation", "error", detectErr)
				http.Error(w, "Failed to validate interface", http.StatusInternalServerError)
				return
			}

			found := false
			for _, iface := range ifaces {
				if iface.Name == update.Interface {
					found = true
					break
				}
			}
			if !found {
				http.Error(w, fmt.Sprintf("Interface '%s' not found", update.Interface), http.StatusBadRequest)
				return
			}

			s.selectedIface = update.Interface
			logging.Info("interface selected", "interface", update.Interface)
		}

		writeJSON(w, StatusResponse{Status: "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMode handles mode switching between reflector and test_master.
func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, ModeResponse{Mode: s.mode})

	case http.MethodPost:
		var req ModeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logging.Warn("mode update failed: invalid JSON", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Mode != modeReflector && req.Mode != modeTestMaster {
			logging.Warn("mode update failed: invalid mode", "mode", req.Mode)
			http.Error(w, "Invalid mode (must be 'reflector' or 'test_master')", http.StatusBadRequest)
			return
		}

		oldMode := s.mode
		s.mode = req.Mode
		logging.Info("mode changed", "from", oldMode, "to", s.mode)
		writeJSON(w, ModeUpdateResponse{Status: "updated", Mode: s.mode})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
