// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// handleAuthLogin issues JWT tokens for valid credentials.
// Sets httpOnly cookies for browser auth and returns tokens for API clients.
//
// Rotates the CSRF token on successful authentication (#87): any token
// previously bound to this session ID is revoked and a fresh one will be
// minted on the next call to GET /api/v1/auth/csrf-token. Rotating on
// every login defends against session-fixation flavored CSRF attacks
// where an attacker primes a token before tricking the user into
// authenticating.
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req AuthLoginRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
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

	// Rotate the CSRF token: revoke any token bound to the new session
	// ID (the JWT payload). The next /api/v1/auth/csrf-token call mints
	// a fresh one. Done before audit logging so an error here does not
	// disturb the audit trail.
	if newSessionID := sessionIDFromJWT(accessToken); newSessionID != "" {
		s.csrfManager.RevokeToken(newSessionID)
	}

	// Audit log the successful login.
	logging.AuditLoginSuccess(r.Context(), r, req.Username, req.Username)

	// Also return tokens in response body for API clients.
	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionDuration).Unix(),
	})
}

// sessionIDFromJWT extracts the JWT payload segment, which is the same
// identifier auth.GetSessionIDFromRequest uses to key CSRF tokens.
// Returns "" if the token is malformed.
func sessionIDFromJWT(token string) string {
	const jwtParts = 2
	parts := splitJWT(token)
	if len(parts) < jwtParts {
		return ""
	}
	return parts[1]
}

// splitJWT splits a JWT on '.' without allocating for the empty-string
// edge case. Defined here rather than in auth/ because
// auth.GetSessionIDFromRequest expects an [http.Request]; this is the
// raw-token equivalent.
func splitJWT(token string) []string {
	if token == "" {
		return nil
	}
	const jwtSegments = 3
	parts := make([]string, 0, jwtSegments)
	start := 0
	for i := range len(token) {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
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

// handleAuthCSRF returns the current session's CSRF token, minting one
// on the first call after login. Subsequent calls for the same session
// return the same token until the session is rotated (next login) or the
// token expires. The token is required in the X-Csrf-Token header on
// every state-changing request — see csrfManager.CSRFMiddleware.
//
// Authentication is required; the handler is registered behind
// s.authMiddleware (see setupRoutes). CSRF is not required to fetch a
// CSRF token (that would be a chicken-and-egg deadlock): a GET request
// short-circuits the CSRF middleware via the safe-method check.
//
// Route: GET /api/v1/auth/csrf-token  (also reachable at the legacy
//
//	/api/v1/auth/csrf path for backwards compatibility).
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

	token, err := s.csrfManager.GetOrCreateToken(sessionID)
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
		if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
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
