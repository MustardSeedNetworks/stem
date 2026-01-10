// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CSRF token configuration.
const (
	// CSRFTokenLength is the length of the CSRF token in bytes before encoding.
	CSRFTokenLength = 32
	// CSRFTokenExpiry is how long a CSRF token remains valid.
	CSRFTokenExpiry = 24 * time.Hour
	// CSRFHeaderName is the HTTP header name for CSRF tokens.
	CSRFHeaderName = "X-Csrf-Token"
	// CSRFCookieName is the cookie name for CSRF tokens.
	CSRFCookieName = "csrf_token"
)

// Internal constants for CSRF token management.
const (
	// csrfCleanupIntervalMinutes is how often expired tokens are cleaned up.
	csrfCleanupIntervalMinutes = 5
	// jwtTokenPartsCount is the minimum number of parts in a valid JWT token.
	jwtTokenPartsCount = 2
)

// CSRF errors.
var (
	// ErrCSRFTokenMissing is returned when no CSRF token is provided.
	ErrCSRFTokenMissing = errors.New("CSRF token missing")
	// ErrCSRFTokenInvalid is returned when the CSRF token is invalid.
	ErrCSRFTokenInvalid = errors.New("CSRF token invalid")
	// ErrCSRFTokenExpired is returned when the CSRF token has expired.
	ErrCSRFTokenExpired = errors.New("CSRF token expired")
)

// CSRFToken represents a CSRF token with its metadata.
type CSRFToken struct {
	Token     string    // The actual token string
	ExpiresAt time.Time // When the token expires
}

// CSRFManager manages CSRF token generation and validation.
type CSRFManager struct {
	mu     sync.RWMutex
	tokens map[string]*CSRFToken // Map of session ID to token metadata
	ctx    context.Context       // Context for shutdown coordination
	cancel context.CancelFunc    // Cancel function for shutdown
	logger *slog.Logger
}

// NewCSRFManager creates a new CSRF manager with context-based cleanup coordination.
func NewCSRFManager(logger *slog.Logger) *CSRFManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &CSRFManager{
		mu:     sync.RWMutex{},
		tokens: make(map[string]*CSRFToken),
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}

	// Start background cleanup goroutine with context cancellation
	go manager.cleanupExpiredTokens()

	return manager
}

// GenerateToken creates a new CSRF token for the given session/user.
// The sessionID should be derived from the user's JWT.
func (m *CSRFManager) GenerateToken(sessionID string) (string, error) {
	// Generate cryptographically secure random bytes
	tokenBytes := make([]byte, CSRFTokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Store the token with expiry
	m.tokens[sessionID] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(CSRFTokenExpiry),
	}

	return token, nil
}

// ValidateToken checks if the provided token is valid for the given session.
// Uses constant-time comparison to prevent timing attacks.
func (m *CSRFManager) ValidateToken(sessionID, token string) error {
	if token == "" {
		return ErrCSRFTokenMissing
	}

	m.mu.RLock()
	storedToken, exists := m.tokens[sessionID]
	m.mu.RUnlock()

	if !exists {
		return ErrCSRFTokenInvalid
	}

	// Check expiry
	now := time.Now()
	if now.After(storedToken.ExpiresAt) {
		// Clean up expired token - re-check under write lock to prevent TOCTOU race
		m.mu.Lock()
		currentToken, stillExists := m.tokens[sessionID]
		if stillExists && currentToken == storedToken {
			delete(m.tokens, sessionID)
		}
		m.mu.Unlock()
		return ErrCSRFTokenExpired
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(token), []byte(storedToken.Token)) != 1 {
		return ErrCSRFTokenInvalid
	}

	return nil
}

// RevokeToken removes a CSRF token, typically on logout.
func (m *CSRFManager) RevokeToken(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, sessionID)
}

// HasToken checks if a CSRF token exists for the given session.
func (m *CSRFManager) HasToken(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.tokens[sessionID]
	return exists
}

// cleanupExpiredTokens periodically removes expired tokens.
func (m *CSRFManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(csrfCleanupIntervalMinutes * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Debug("CSRF cleanup goroutine stopping")
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for sessionID, token := range m.tokens {
				if now.After(token.ExpiresAt) {
					delete(m.tokens, sessionID)
				}
			}
			m.mu.Unlock()
		}
	}
}

// Stop gracefully shuts down the CSRF manager by stopping the cleanup goroutine.
func (m *CSRFManager) Stop() {
	m.cancel()
}

// CSRFMiddleware returns HTTP middleware that validates CSRF tokens on state-changing requests.
// It exempts GET, HEAD, OPTIONS, and TRACE methods as they should be safe/idempotent.
func (m *CSRFManager) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods (RFC 7231)
		if r.Method == http.MethodGet ||
			r.Method == http.MethodHead ||
			r.Method == http.MethodOptions ||
			r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for non-API routes (static files, etc.)
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for auth endpoints that don't have a session yet
		// - login: no session yet
		// - refresh: may not have valid CSRF when access token expires
		// - logout: safe operation
		if r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/auth/refresh" ||
			r.URL.Path == "/api/v1/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract session ID from request (use username from JWT payload)
		sessionID := GetSessionIDFromRequest(r)

		// If no session ID, skip CSRF check - let auth middleware handle authentication.
		// CSRF protection is only relevant for authenticated sessions.
		if sessionID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if a CSRF token exists for this session.
		// If no token was ever generated (user never called /api/v1/auth/csrf),
		// skip CSRF check and let auth middleware validate the session.
		if !m.HasToken(sessionID) {
			next.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from request header
		token := r.Header.Get(CSRFHeaderName)

		// Validate the token
		err := m.ValidateToken(sessionID, token)
		if err != nil {
			m.logger.Warn("CSRF validation failed",
				"path", r.URL.Path,
				"method", r.Method,
				"error", err)

			switch {
			case errors.Is(err, ErrCSRFTokenMissing):
				http.Error(w, "CSRF token required", http.StatusForbidden)
			case errors.Is(err, ErrCSRFTokenExpired):
				http.Error(w, "CSRF token expired", http.StatusForbidden)
			default:
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetSessionIDFromRequest extracts a session identifier from the request.
// Uses the JWT payload part as a session identifier.
func GetSessionIDFromRequest(r *http.Request) string {
	// Extract from JWT token in cookie
	token, _ := GetTokenFromRequest(r)
	if token != "" {
		// Use the payload part of the token as a session identifier
		parts := strings.Split(token, ".")
		if len(parts) >= jwtTokenPartsCount {
			return parts[1] // Use payload part as identifier
		}
	}

	return ""
}

// TokenCount returns the number of active tokens (for testing).
func (m *CSRFManager) TokenCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tokens)
}

// CtxDone returns the context done channel (for testing shutdown).
func (m *CSRFManager) CtxDone() <-chan struct{} {
	return m.ctx.Done()
}
