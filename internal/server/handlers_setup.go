// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"net/http"
	"os"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/logging"
)

const (
	// suggestedPasswordLength is the length of auto-generated passwords.
	suggestedPasswordLength = 24
)

// SetupStatusResponse is the response from the setup status endpoint.
type SetupStatusResponse struct {
	NeedsSetup        bool   `json:"needsSetup"`
	Username          string `json:"username,omitempty"`
	SuggestedPassword string `json:"suggestedPassword,omitempty"`
	SetupToken        string `json:"setupToken,omitempty"`
}

// SetupCompleteRequest is the request body for completing setup.
type SetupCompleteRequest struct {
	Password   string `json:"password"`
	SetupToken string `json:"setupToken"`
}

// handleSetupStatus checks if initial setup is required.
// Returns setup token and suggested password if setup is needed.
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	// Check if setup is needed by checking environment variables.
	needsSetup := s.needsInitialSetup()

	resp := SetupStatusResponse{
		NeedsSetup:        needsSetup,
		Username:          os.Getenv("STEM_AUTH_USERNAME"),
		SuggestedPassword: "",
		SetupToken:        "",
	}

	// If setup is needed, generate setup token and suggested password.
	if needsSetup {
		// Generate a suggested password.
		suggestedPassword, err := auth.GenerateSecurePassword(suggestedPasswordLength)
		if err == nil {
			resp.SuggestedPassword = suggestedPassword
		}

		// Generate one-time setup token.
		if s.setupTokenManager != nil {
			setupToken, tokenErr := s.setupTokenManager.GenerateToken()
			if tokenErr != nil {
				logging.Error("Failed to generate setup token", "error", tokenErr)
				WriteAPIError(w, ErrAPIInternalError)
				return
			}
			resp.SetupToken = setupToken
			logging.Info("Setup token generated", "event", "auth.setup.token_generated")
		}
	}

	writeJSON(w, resp)
}

// handleSetupComplete completes initial setup by setting admin password.
// Requires valid setup token to prevent CSRF attacks.
func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Check if setup is actually needed.
	if !s.needsInitialSetup() {
		http.Error(w, "Setup has already been completed", http.StatusForbidden)
		return
	}

	// Decode request.
	var req SetupCompleteRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	// Validate setup token.
	if s.setupTokenManager == nil || !s.setupTokenManager.ValidateToken(req.SetupToken) {
		logging.Warn("Invalid setup token provided", "event", "auth.setup.invalid_token")
		http.Error(w, "Invalid or expired setup token", http.StatusForbidden)
		return
	}

	// Validate password strength.
	passwordErr := auth.ValidatePasswordStrength(req.Password)
	if passwordErr != nil {
		http.Error(w, passwordErr.Error(), http.StatusBadRequest)
		return
	}

	// Hash the new password.
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		logging.Error("Failed to hash password", "error", err)
		WriteAPIError(w, ErrAPIInternalError)
		return
	}

	// Update the auth manager with the new password hash.
	s.authManager.UpdatePasswordHash(r.Context(), hash)

	// Invalidate the setup token.
	if s.setupTokenManager != nil {
		s.setupTokenManager.Invalidate()
	}

	// Mark setup as complete.
	s.markSetupComplete()

	// Log successful setup.
	logging.Info("Initial setup completed - admin password configured",
		"event", "auth.setup.complete",
		"username", os.Getenv("STEM_AUTH_USERNAME"))

	writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Setup completed successfully",
	})
}

// needsInitialSetup checks if the initial setup wizard should be shown.
// Returns true if no password has been configured yet.
func (s *Server) needsInitialSetup() bool {
	// Check if setup has been explicitly marked as complete.
	if s.setupComplete {
		return false
	}

	// Check for the setup mode flag from environment.
	// If STEM_SETUP_MODE=true, show the setup wizard.
	if os.Getenv("STEM_SETUP_MODE") == "true" {
		return true
	}

	// If password hash is default/empty, setup is needed.
	if s.authManager != nil {
		hash := s.authManager.GetPasswordHash()
		if auth.IsDefaultPasswordHash(hash) {
			return true
		}
	}

	return false
}

// markSetupComplete marks the initial setup as complete.
func (s *Server) markSetupComplete() {
	s.setupComplete = true
	s.setupModeStartTime = time.Time{} // Reset setup mode timer.
}
