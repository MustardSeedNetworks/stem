// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/logging"
)

// RecoveryStatusResponse represents the recovery mode status.
type RecoveryStatusResponse struct {
	Active        bool   `json:"active"`
	RemainingTime int    `json:"remainingTime,omitempty"` // Seconds remaining until token expires
	Instructions  string `json:"instructions,omitempty"`
}

// RecoveryCompleteRequest represents a password recovery request.
type RecoveryCompleteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// RecoveryCompleteResponse represents a recovery completion response.
type RecoveryCompleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RecoveryInstructionsResponse provides instructions for starting recovery.
type RecoveryInstructionsResponse struct {
	TriggerFile string   `json:"triggerFile"`
	TokenFile   string   `json:"tokenFile"`
	ExpiryTime  string   `json:"expiryTime"`
	Steps       []string `json:"steps"`
}

// handleRecoveryStatus checks if password recovery mode is active.
// This endpoint is public (no auth required) so the login page can check status.
func (s *Server) handleRecoveryStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	// Check if recovery manager is configured.
	if s.recoveryTokenManager == nil {
		writeJSON(w, RecoveryStatusResponse{
			Active:        false,
			RemainingTime: 0,
			Instructions:  "",
		})
		return
	}

	// Check if recovery mode is active (this also generates token if trigger file exists).
	active := s.recoveryTokenManager.CheckRecoveryMode()

	resp := RecoveryStatusResponse{
		Active:        active,
		RemainingTime: 0,
		Instructions:  "",
	}

	if active {
		remaining := s.recoveryTokenManager.RemainingTime()
		resp.RemainingTime = int(remaining.Seconds())
		resp.Instructions = "Enter the recovery token from " + s.recoveryTokenManager.TokenFilePath()
	}

	writeJSON(w, resp)
}

// handleRecoveryComplete processes password recovery with a valid token.
// Requires a valid recovery token that was written to the filesystem.
func (s *Server) handleRecoveryComplete(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Check if recovery manager is configured.
	if s.recoveryTokenManager == nil {
		logging.Warn("Recovery attempt but recovery manager not configured",
			"client_ip", clientIP,
			"event", "auth.recovery.not_configured")
		http.Error(w, "Password recovery is not available", http.StatusServiceUnavailable)
		return
	}

	// Decode request.
	var req RecoveryCompleteRequest
	if !decodeJSONStrict(w, r, &req) {
		return
	}

	// Validate the recovery token.
	if !s.recoveryTokenManager.ValidateAndConsume(req.Token) {
		logging.Warn("Recovery failed - invalid or expired token",
			"client_ip", clientIP,
			"event", "auth.recovery.invalid_token")
		http.Error(w, "Invalid or expired recovery token", http.StatusUnauthorized)
		return
	}

	// Token is valid - proceed with password reset.

	// Validate password strength.
	passwordErr := auth.ValidatePasswordStrength(req.Password)
	if passwordErr != nil {
		logging.Warn("Recovery failed - weak password",
			"client_ip", clientIP,
			"event", "auth.recovery.weak_password")
		http.Error(w, passwordErr.Error(), http.StatusBadRequest)
		return
	}

	// Hash the new password.
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		logging.Error("Failed to hash password during recovery", "error", err)
		WriteError(w, ErrInternalError)
		return
	}

	// Update the auth manager with the new password hash.
	s.authManager.UpdatePasswordHash(r.Context(), hash)

	// Cleanup recovery files.
	s.recoveryTokenManager.Cleanup()

	// Security audit log.
	logging.Info("Password recovery completed successfully",
		"client_ip", clientIP,
		"event", "auth.recovery.success")

	writeJSON(w, RecoveryCompleteResponse{
		Success: true,
		Message: "Password has been reset. All existing sessions have been invalidated.",
	})
}

// handleRecoveryInstructions returns instructions for password recovery.
// This is a public endpoint that provides information without exposing secrets.
func (s *Server) handleRecoveryInstructions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	if s.recoveryTokenManager == nil {
		http.Error(w, "Password recovery is not configured", http.StatusServiceUnavailable)
		return
	}

	expiryDuration := s.recoveryTokenManager.TokenExpiryDuration()
	resp := RecoveryInstructionsResponse{
		TriggerFile: s.recoveryTokenManager.TriggerFilePath(),
		TokenFile:   s.recoveryTokenManager.TokenFilePath(),
		ExpiryTime:  expiryDuration.String(),
		Steps: []string{
			"1. SSH into the server",
			"2. Create the trigger file: touch " + s.recoveryTokenManager.TriggerFilePath(),
			"3. Wait a moment, then read the token: cat " + s.recoveryTokenManager.TokenFilePath(),
			"4. Return to this page and enter the token with your new password",
			"5. The token expires after " + expiryDuration.String(),
		},
	}

	writeJSON(w, resp)
}
