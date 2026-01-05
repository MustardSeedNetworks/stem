// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/krisarmstrong/stem/internal/logging"
)

// handleAuthLogin issues JWT tokens for valid credentials.
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	token, err := s.authManager.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	writeJSON(w, AuthLoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.authManager.SessionDuration()).Unix(),
	})
}

// handleTestResultsWebSocket upgrades a connection and streams test events.
func (s *Server) handleTestResultsWebSocket(w http.ResponseWriter, r *http.Request) {
	if err := s.requireAuth(r); err != nil {
		s.writeAuthError(w, err)
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
		if _, _, err := conn.NextReader(); err != nil {
			return
		}
	}
}
