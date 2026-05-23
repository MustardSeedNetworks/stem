// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"net/http"

	"github.com/krisarmstrong/stem/internal/license"
)

// handleLicense returns current license status.
func (s *Server) handleLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	if s.licenseManager == nil {
		writeJSON(w, LicenseStatus{
			Activated:     false,
			IsTrialMode:   false,
			Tier:          0,
			TierName:      "",
			DaysRemaining: 0,
			Features:      nil,
			DeviceHash:    "",
			LicenseKey:    "",
			Message:       "License manager not initialized",
		})
		return
	}

	state := s.licenseManager.GetState()
	fp := s.licenseManager.GetFingerprint()

	var status LicenseStatus

	switch {
	case state == nil:
		status = LicenseStatus{
			Activated:     false,
			IsTrialMode:   false,
			Tier:          0,
			TierName:      "",
			DaysRemaining: 0,
			Features:      nil,
			DeviceHash:    fp.Hash(),
			LicenseKey:    "",
			Message:       "No license. Start a trial or enter a license key.",
		}
	case state.IsTrialMode:
		status = LicenseStatus{
			Activated:     true,
			IsTrialMode:   true,
			Tier:          int(license.TierProfessional),
			TierName:      "Trial",
			DaysRemaining: s.licenseManager.TrialDaysRemaining(),
			Features:      state.Features,
			DeviceHash:    fp.Hash(),
			LicenseKey:    "",
			Message:       fmt.Sprintf("Trial mode: %d days remaining", s.licenseManager.TrialDaysRemaining()),
		}
	default:
		activated := s.licenseManager.IsActivated()
		message := "License expired or invalid"
		if activated {
			message = fmt.Sprintf("Licensed: %s", state.Tier)
		}
		status = LicenseStatus{
			Activated:     activated,
			IsTrialMode:   false,
			Tier:          int(state.Tier),
			TierName:      state.Tier.String(),
			DaysRemaining: 0,
			Features:      state.Features,
			DeviceHash:    fp.Hash(),
			LicenseKey:    license.FormatKey(state.LicenseKey),
			Message:       message,
		}
	}

	writeJSON(w, status)
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
	if !decodeJSONStrict(w, r, &req) {
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
