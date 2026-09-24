// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
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

	needsSetup := s.needsInitialSetup()

	resp := SetupStatusResponse{
		NeedsSetup:        needsSetup,
		Username:          s.authManager.GetUsername(),
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
				WriteError(w, ErrInternalError)
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

	// One claim at a time: a second request racing the first must see the
	// credential the first one stored, not an unclaimed daemon.
	s.setupMu.Lock()
	defer s.setupMu.Unlock()

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

	username := s.authManager.GetUsername()
	prevAlgorithm := detectHashAlgorithm(s.authManager.GetPasswordHash())

	// Validate password strength.
	if !validatePasswordOrReject(w, r, req.Password, username, prevAlgorithm) {
		return
	}

	// Hash the new password.
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		logging.Error("Failed to hash password", "error", err)
		logging.AuditPasswordChange(r.Context(), r, username,
			logging.PasswordChangeRejected, "hash_failed", prevAlgorithm, err.Error())
		WriteError(w, ErrInternalError)
		return
	}

	if saveErr := s.saveCredential(r.Context(), hash); saveErr != nil {
		logging.Error("Failed to store the administrator credential", "error", saveErr)
		logging.AuditPasswordChange(r.Context(), r, username,
			logging.PasswordChangeRejected, "store_failed", prevAlgorithm, saveErr.Error())
		WriteError(w, ErrInternalError)
		return
	}

	// Invalidate the setup token.
	if s.setupTokenManager != nil {
		s.setupTokenManager.Invalidate()
	}

	// Log successful setup.
	logging.Info("Initial setup completed - admin password configured",
		"event", "auth.setup.complete",
		"username", username)
	logging.AuditPasswordChange(r.Context(), r, username,
		logging.PasswordChangeSuccess, "", prevAlgorithm, "initial setup")

	writeJSON(w, map[string]string{
		"status":  "success",
		"message": "Setup completed successfully",
	})
}

// needsInitialSetup reports whether the daemon is still unclaimed: no
// administrator password has been stored or configured.
func (s *Server) needsInitialSetup() bool {
	return auth.IsDefaultPasswordHash(s.authManager.GetPasswordHash())
}
