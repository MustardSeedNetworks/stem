// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// LicenseStatusOf renders a manager's entitlement state in the one shape
// every reader sees. The CLI renders the same struct whether it read it from
// a running daemon or, with no daemon on the host, built it here from its own
// manager — two renderers would be two descriptions of one licence.
func LicenseStatusOf(mgr *license.Manager) LicenseStatus {
	if mgr == nil {
		return LicenseStatus{Message: "License manager not initialized"}
	}

	state := mgr.GetState()
	fp := mgr.GetFingerprint()

	switch {
	case state == nil:
		return LicenseStatus{
			DeviceHash: fp.Hash(),
			Platform:   fp.Platform,
			Message:    "No license. Start a trial or enter a license key.",
		}
	case state.IsTrialMode:
		return LicenseStatus{
			Activated:     true,
			IsTrialMode:   true,
			Tier:          int(license.TierProfessional),
			TierName:      "Trial",
			DaysRemaining: mgr.TrialDaysRemaining(),
			Features:      state.Features,
			DeviceHash:    fp.Hash(),
			Platform:      fp.Platform,
			Message:       fmt.Sprintf("Trial mode: %d days remaining", mgr.TrialDaysRemaining()),
		}
	default:
		activated := mgr.IsActivated()
		message := "License expired or invalid"
		if activated {
			message = fmt.Sprintf("Licensed: %s", license.Tier(state.Tier))
		}
		return LicenseStatus{
			Activated:  activated,
			Tier:       state.Tier,
			TierName:   license.Tier(state.Tier).String(),
			Features:   state.Features,
			DeviceHash: fp.Hash(),
			Platform:   fp.Platform,
			LicenseKey: license.FormatKey(state.LicenseKey),
			ExpiresAt:  state.ExpiresAt,
			Message:    message,
		}
	}
}

// handleLicense reports the current licence state, and on DELETE removes it.
// Deactivation lives on this route rather than the CLI's own file write: the
// daemon holds the manager, so a licence removed anywhere else would go on
// being served from memory until a restart (#1335).
func (s *Server) handleLicense(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, LicenseStatusOf(s.licenseManager))
	case http.MethodDelete:
		s.deactivateLicense(w)
	default:
		WriteMethodNotAllowed(w)
	}
}

// deactivateLicense removes the activation the daemon is holding.
func (s *Server) deactivateLicense(w http.ResponseWriter) {
	if s.licenseManager == nil {
		writeJSON(w, ErrorResponse{Success: false, Message: "License manager not initialized"})
		return
	}
	if err := s.licenseManager.Deactivate(); err != nil {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to deactivate: %v", err),
		})
		return
	}
	writeJSON(w, ErrorResponse{Success: true, Message: "License deactivated"})
}

// handleLicenseActivate activates a license key.
func (s *Server) handleLicenseActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	if s.licenseManager == nil {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License manager not initialized",
		})
		return
	}

	var req LicenseActivateRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	if req.LicenseKey == "" {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License key is required",
		})
		return
	}

	result := s.licenseManager.Activate(req.LicenseKey)
	writeJSON(w, result)
}

// handleLicenseTrial starts or checks trial status.
func (s *Server) handleLicenseTrial(w http.ResponseWriter, r *http.Request) {
	if s.licenseManager == nil {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License manager not initialized",
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Check trial status.
		if s.licenseManager.IsTrialValid() {
			writeJSON(w, TrialStatusResponse{
				Active:        true,
				DaysRemaining: s.licenseManager.TrialDaysRemaining(),
			})
		} else {
			writeJSON(w, TrialStatusResponse{Active: false, DaysRemaining: 0})
		}

	case http.MethodPost:
		// Start trial.
		result := s.licenseManager.StartTrial()
		writeJSON(w, result)

	default:
		WriteMethodNotAllowed(w)
	}
}
