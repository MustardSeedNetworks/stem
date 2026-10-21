// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/password"

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
	// SetupTokenFile is where the operator reads the token setup/complete
	// requires. The token itself is never returned here.
	SetupTokenFile string `json:"setupTokenFile,omitempty"`
}

// SetupCompleteRequest is the request body for completing setup.
type SetupCompleteRequest struct {
	Password   string `json:"password"`
	SetupToken string `json:"setupToken"`
}

// handleSetupStatus reports whether initial setup is required and, while it
// is, where the operator finds the setup token.
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	resp := SetupStatusResponse{
		NeedsSetup:        s.needsInitialSetup(),
		Username:          s.authManager.GetUsername(),
		SuggestedPassword: "",
		SetupTokenFile:    "",
	}
	if resp.NeedsSetup {
		if suggestedPassword, err := auth.GenerateSecurePassword(suggestedPasswordLength); err == nil {
			resp.SuggestedPassword = suggestedPassword
		}
		resp.SetupTokenFile = s.setupTokenManager.FilePath()
	}

	writeJSON(w, resp)
}

// issueSetupToken writes the token for this setup window and logs where the
// operator finds it. Without the file setup cannot complete, so a failure is
// logged rather than ending startup: the rest of the daemon runs.
func (s *Server) issueSetupToken() {
	if err := s.setupTokenManager.Issue(); err != nil {
		logging.Error("Failed to write the setup token; setup cannot complete",
			"event", "auth.setup.token_failed", "path", s.setupTokenManager.FilePath(), "error", err)
		return
	}
	logging.Info("Setup pending: enter the setup token from this file in the setup wizard",
		"event", "auth.setup.token_issued", "path", s.setupTokenManager.FilePath())
}

// handleSetupComplete completes initial setup by setting admin password.
// The setup token proves the caller can read the host's data directory, so
// neither a network peer nor a cross-site form can claim the daemon.
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
	if !s.setupTokenManager.Validate(strings.TrimSpace(req.SetupToken)) {
		logging.Warn("Invalid setup token provided", "event", "auth.setup.invalid_token")
		http.Error(w, "Invalid setup token. Copy it from "+s.setupTokenManager.FilePath()+" and try again.",
			http.StatusForbidden)
		return
	}

	username := s.authManager.GetUsername()
	prevAlgorithm := detectHashAlgorithm(s.authManager.GetPasswordHash())

	// Validate password strength.
	if !validatePasswordOrReject(w, r, req.Password, username, prevAlgorithm) {
		return
	}

	// Hash the new password.
	hash, err := password.Hash(req.Password)
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

	if consumeErr := s.setupTokenManager.Consume(); consumeErr != nil {
		logging.Warn("Failed to remove the setup token file", "error", consumeErr)
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
	return s.authManager.GetPasswordHash() == ""
}
