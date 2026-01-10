// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"net/http"
	"time"

	"github.com/krisarmstrong/stem/internal/logging"
)

// handleAuthLogin issues JWT tokens for valid credentials.
// Returns both access and refresh tokens.
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

	// Audit log the successful login.
	logging.AuditLoginSuccess(r.Context(), r, req.Username, req.Username)

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.authManager.SessionDuration()).Unix(),
	})
}

// handleAuthLogout revokes the current access token.
func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	// Get the token from the Authorization header.
	claims, err := s.extractClaims(r)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	// Revoke the token.
	s.authManager.RevokeToken(claims)

	// Audit log the logout.
	logging.AuditLogout(r.Context(), r, claims.Username)

	writeJSON(w, map[string]any{
		"success": true,
		"message": "Logged out successfully",
	})
}

// handleAuthRefresh exchanges a refresh token for a new access token.
func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req AuthRefreshRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	accessToken, err := s.authManager.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Audit log the token refresh failure.
		logging.AuditTokenInvalid(r.Context(), r, "refresh token: "+err.Error())
		s.writeAuthError(w, err)
		return
	}

	// Audit log the successful token refresh (extract username from the new token if possible).
	logging.AuditTokenRefresh(r.Context(), r, "")

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: "",
		ExpiresAt:    time.Now().Add(s.authManager.SessionDuration()).Unix(),
	})
}
