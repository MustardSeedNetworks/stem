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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthLoginRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	accessToken, refreshToken, err := s.authManager.AuthenticateWithRefresh(r.Context(), req.Username, req.Password)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.authManager.SessionDuration()).Unix(),
	})
}

// handleAuthLogout revokes the current access token.
func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	writeJSON(w, map[string]any{
		"success": true,
		"message": "Logged out successfully",
	})
}

// handleAuthRefresh exchanges a refresh token for a new access token.
func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRefreshRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	accessToken, err := s.authManager.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	writeJSON(w, AuthLoginResponse{
		Token:        accessToken,
		RefreshToken: "",
		ExpiresAt:    time.Now().Add(s.authManager.SessionDuration()).Unix(),
	})
}

// handleTestResultsWebSocket upgrades a connection and streams test events.
func (s *Server) handleTestResultsWebSocket(w http.ResponseWriter, r *http.Request) {
	authErr := s.requireAuth(r)
	if authErr != nil {
		s.writeAuthError(w, authErr)
		return
	}

	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Error("websocket upgrade failed", "error", err)
		http.Error(w, "Failed to upgrade websocket", http.StatusInternalServerError)
		return
	}

	s.registerWSClient(conn)
	defer s.unregisterWSClient(conn)
	s.sendCurrentTestState(conn)

	for {
		_, _, nextErr := conn.NextReader()
		if nextErr != nil {
			return
		}
	}
}
