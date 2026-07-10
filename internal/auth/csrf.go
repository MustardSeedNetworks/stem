// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"
)

// CSRF token configuration.
const (
	// CSRFHeaderName is the HTTP header name for CSRF tokens.
	CSRFHeaderName = "X-Csrf-Token"
	// CSRFCookieName is the cookie name for CSRF tokens.
	CSRFCookieName = "csrf_token"
)

// CSRFManager backs Stem's CSRF protection with the fleet-shared foundation
// per-session token manager (github.com/MustardSeedNetworks/foundation/pkg/csrf).
// Token storage, generation, constant-time validation, expiry sweep and the
// sha256(session-key) model all live in foundation now; Stem keeps only its
// product-specific policy — the fail-closed exempt-list, the response format,
// and the JWT-derived session key — in CSRFMiddleware and GetSessionIDFromRequest
// below.
type CSRFManager struct {
	mgr    *csrf.Manager
	logger *slog.Logger
}

// NewCSRFManager creates a CSRF manager backed by foundation's per-session
// token manager, whose cleanup goroutine is stopped via Stop() on shutdown.
func NewCSRFManager(logger *slog.Logger) *CSRFManager {
	return &CSRFManager{mgr: csrf.NewManager(), logger: logger}
}

// GenerateToken mints a fresh CSRF token for the given session key, replacing
// any existing one.
func (m *CSRFManager) GenerateToken(sessionID string) (string, error) {
	return m.mgr.Generate(sessionID)
}

// GetOrCreateToken returns the session's existing unexpired token or mints a
// new one — used by the /auth/csrf-token endpoint so a polling UI gets a
// stable value within the session lifetime.
func (m *CSRFManager) GetOrCreateToken(sessionID string) (string, error) {
	return m.mgr.GetOrCreate(sessionID)
}

// ValidateToken checks the token against the one stored for sessionID. It
// returns one of csrf.ErrTokenMissing / ErrTokenInvalid / ErrTokenExpired so
// CSRFMiddleware can render a distinct response per cause.
func (m *CSRFManager) ValidateToken(sessionID, token string) error {
	return m.mgr.Validate(sessionID, token)
}

// RevokeToken drops a session's CSRF token, e.g. on logout or MFA session
// rotation (a new session ID is minted and the old token must not linger).
func (m *CSRFManager) RevokeToken(sessionID string) {
	m.mgr.Revoke(sessionID)
}

// Stop shuts down the manager's background cleanup goroutine.
func (m *CSRFManager) Stop() {
	m.mgr.Stop()
}

// isCSRFExemptPath reports whether path is on the explicit allow-list of
// endpoints that bypass CSRF validation. The default for everything else
// is "validate" — this is the fail-closed posture required by task #87.
// Add to this list only after thinking through whether the endpoint can
// be safely invoked cross-origin without a CSRF token.
//
// Allow-list contents (keep in sync with the body):
//
//   - /api/v1/auth/login: pre-session, no CSRF possible. The credential
//     itself (username + password) provides the proof-of-intent.
//   - /api/v1/setup/status: read-only GET — already covered by the safe-
//     method check above; not duplicated here. The safe-method short-
//     circuit returns first.
//   - /api/v1/harvest/logs/client: fire-and-forget client-side log
//     ingestion that runs before the user has authenticated and therefore
//     cannot present a CSRF token. The endpoint accepts only telemetry
//     payloads and writes no user-visible state.
//   - /api/v1/sso/*: SSO callback handlers (OAuth/SAML/OIDC). The IdP
//     POSTs to these endpoints with its own signed-assertion proof of
//     intent; a CSRF token would not exist at that point in the flow.
//
// REMOVED from the previous exempt list (#87 fail-closed):
//   - /api/v1/auth/refresh: state change on a session-scoped credential.
//     The browser holds the refresh-token cookie; CSRF is the standard
//     defense against an attacker initiating a silent refresh.
//   - /api/v1/auth/logout: without CSRF, an attacker can force-logout a
//     user (denial-of-service / session-fixation vector).
//   - /api/v1/setup/complete: completes initial admin setup; state-
//     changing. Setup token alone is not a substitute for CSRF — the
//     token can be lifted from the network or browser memory.
//
// Implemented as a function (not a package-level map/slice) to avoid
// gochecknoglobals while keeping the list reviewable in one place.
func isCSRFExemptPath(path string) bool {
	switch path {
	case "/api/v1/auth/login",
		"/api/v1/harvest/logs/client",
		// Wave 3 (#85): MFA login finishers are pre-session in the
		// same way /api/v1/auth/login is — the mfa_token (a server-
		// signed pending-MFA JWT) provides the proof of intent and
		// the request cannot carry a CSRF token because the caller
		// has not yet completed authentication. Rate-limiting still
		// caps brute force.
		"/api/v1/auth/login/totp",
		"/api/v1/auth/webauthn/login/begin",
		"/api/v1/auth/webauthn/login/finish":
		return true
	}
	// Prefix match for SSO callbacks whose suffix varies by provider
	// (e.g. /api/v1/sso/google/callback, /api/v1/sso/saml/acs).
	return strings.HasPrefix(path, "/api/v1/sso/")
}

// CSRFMiddleware returns HTTP middleware that validates CSRF tokens on
// state-changing requests. It exempts GET, HEAD, OPTIONS, and TRACE
// methods as they should be safe/idempotent, plus an explicit allow-list
// (isCSRFExemptPath) of endpoints that cannot reasonably carry a CSRF
// token (pre-session login, IdP callbacks, etc.).
//
// Everything else fails closed: an authenticated state-changing request
// without a valid CSRF token returns 403 Forbidden.
func (m *CSRFManager) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods (RFC 7231).
		if r.Method == http.MethodGet ||
			r.Method == http.MethodHead ||
			r.Method == http.MethodOptions ||
			r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for non-API routes (static files, etc.).
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Explicit exempt-list — see isCSRFExemptPath for the rationale.
		if isCSRFExemptPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Derive the per-session key from the request's authenticated JWT.
		sessionID := GetSessionIDFromRequest(r)

		// No session = no CSRF protection needed yet. Auth middleware
		// will handle the authentication check and return 401 for
		// state-changing requests that need a session. This preserves
		// the "401 if unauthenticated, 403 if CSRF-missing" semantics
		// the UI relies on to differentiate the two failure modes.
		if sessionID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from request header.
		token := r.Header.Get(CSRFHeaderName)

		// Validate the token. Fail closed: if the session never minted a
		// CSRF token, validation returns csrf.ErrTokenInvalid and we 403.
		// Clients must call GET /api/v1/auth/csrf-token after login to
		// obtain the token.
		err := m.ValidateToken(sessionID, token)
		if err != nil {
			m.logger.WarnContext(r.Context(), "CSRF validation failed",
				"path", r.URL.Path,
				"method", r.Method,
				"error", err)

			switch {
			case errors.Is(err, csrf.ErrTokenMissing):
				http.Error(w, "CSRF token required", http.StatusForbidden)
			case errors.Is(err, csrf.ErrTokenExpired):
				http.Error(w, "CSRF token expired", http.StatusForbidden)
			default:
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetSessionIDFromRequest derives the CSRF session key from the request's
// authenticated JWT. The JWT is extracted the same way the auth middleware
// reads it (GetTokenFromRequest), then hashed via foundation's SessionKey so
// the bearer plaintext is never stored in the manager. Returns "" when the
// request carries no token. Exported for the CSRF token endpoint handler,
// which must derive the same key it later validates against.
func GetSessionIDFromRequest(r *http.Request) string {
	token, _ := GetTokenFromRequest(r)
	// Only a JWT-shaped bearer (a browser session token) is CSRF-relevant.
	// A malformed value or a non-JWT bearer (e.g. an API token, which a
	// cross-site attacker cannot set) gets no session key, so the request
	// passes through to the auth layer — which returns 401 for an invalid
	// token — instead of being 403'd here for a missing CSRF token. This
	// mirrors the pre-migration gate (payload-segment extraction returned ""
	// for a non-JWT), changing only the key derivation to sha256(bearer).
	const jwtMinParts = 2
	if len(strings.Split(token, ".")) < jwtMinParts {
		return ""
	}
	return csrf.SessionKey(token)
}
