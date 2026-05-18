// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/logging"
)

// handleAuthLogin issues JWT tokens for valid credentials.
// Sets httpOnly cookies for browser auth and returns tokens for API clients.
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req AuthLoginRequest
	if !decodeJSONStrict(w, r, &req) {
		return
	}

	accessToken, refreshToken, err := s.authManager.AuthenticateWithRefresh(r.Context(), req.Username, req.Password)
	if err != nil {
		// Audit log the failed login attempt.
		logging.AuditLoginFailure(r.Context(), r, req.Username, err.Error())
		s.writeAuthError(w, err)
		return
	}

	// Set httpOnly cookies for secure browser-based auth.
	sessionDuration := s.authManager.SessionDuration()
	auth.SetAccessTokenCookie(w, accessToken, sessionDuration, s.cookieConfig)
	auth.SetRefreshTokenCookie(w, refreshToken, sessionDuration*refreshMultiplier, s.cookieConfig)

	// Audit log the successful login.
	logging.AuditLoginSuccess(r.Context(), r, req.Username, req.Username)

	// Also return tokens in response body for API clients.
	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}

// handleAuthLogout revokes the current access token and clears auth cookies.
func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Get the token from cookies or Authorization header.
	claims, err := s.extractClaims(r)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	// Revoke the token.
	s.authManager.RevokeToken(claims)

	// Clear auth cookies.
	auth.ClearAuthCookies(w, s.cookieConfig)

	// Audit log the logout.
	logging.AuditLogout(r.Context(), r, claims.Username)

	writeJSON(w, map[string]any{
		"success": true,
		"message": "Logged out successfully",
	})
}

// handleAuthCSRF generates or returns the current CSRF token for the session.
// The token is required in the X-Csrf-Token header for all state-changing requests.
func (s *Server) handleAuthCSRF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	// Get session ID from the access token.
	sessionID := auth.GetSessionIDFromRequest(r)
	if sessionID == "" {
		WriteAuthError(w, errMissingAuthToken)
		return
	}

	// Generate a new CSRF token for this session.
	token, err := s.csrfManager.GenerateToken(sessionID)
	if err != nil {
		logging.Error("Failed to generate CSRF token", "error", err)
		WriteError(w, ErrInternalError)
		return
	}

	writeJSON(w, map[string]string{
		"token": token,
	})
}

// handleAuthRefresh exchanges a refresh token for a new access token.
// Supports both cookie-based and request body refresh tokens.
func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Try to get refresh token from cookie first (most secure), then request body.
	var refreshToken string

	// Try cookie first.
	token, cookieErr := auth.GetRefreshTokenFromCookie(r)
	if cookieErr == nil && token != "" {
		refreshToken = token
	} else {
		// Fallback to request body for API clients.
		var req AuthRefreshRequest
		if !decodeJSONStrict(w, r, &req) {
			return // Error already written.
		}
		if req.RefreshToken == "" {
			WriteInvalidRequest(w, "Missing refresh token")
			return
		}
		refreshToken = req.RefreshToken
	}

	accessToken, err := s.authManager.RefreshAccessToken(r.Context(), refreshToken)
	if err != nil {
		// Audit log the token refresh failure.
		logging.AuditTokenInvalid(r.Context(), r, "refresh token: "+err.Error())
		s.writeAuthError(w, err)
		return
	}

	// Set new access token cookie.
	sessionDuration := s.authManager.SessionDuration()
	auth.SetAccessTokenCookie(w, accessToken, sessionDuration, s.cookieConfig)

	// Audit log the successful token refresh.
	logging.AuditTokenRefresh(r.Context(), r, "")

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: "",
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}
