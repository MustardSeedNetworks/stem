// SPDX-License-Identifier: BUSL-1.1

// Package api provides the unified HTTP server for The Stem WebUI.
//
// # Architecture Overview
//
// This is the single web server for The Stem, serving both the embedded React
// frontend and the REST API. There are no separate web servers for reflector
// or testmaster modes - all functionality is consolidated here.
//
// The server supports two operating modes:
//   - "reflector" - Packet reflection mode (Tier 1 license)
//   - "test_master" - Test execution mode (Tier 2 license)
//
// Mode is selected via the API (/api/v1/mode) and determines which features
// are active. Both modes share the same server instance and API surface.
//
// # API Endpoints (v1)
//
// All API endpoints are versioned under /api/v1/.
// API responses include the X-API-Version header.
//
// Kubernetes Health Probes (not versioned):
//   - GET /health/live      - Liveness probe (returns 200 if server is running)
//   - GET /health/ready     - Readiness probe (returns 200 if ready to accept traffic)
//
// Mode Management:
//   - GET  /api/v1/mode        - Get current operating mode
//   - POST /api/v1/mode        - Set operating mode (reflector/test_master)
//
// Interface Management:
//   - GET  /api/v1/interfaces  - List available network interfaces
//   - GET  /api/v1/settings    - Get current settings (interface, mode)
//   - POST /api/v1/settings    - Update settings (validates interface exists)
//
// Reflector Mode:
//   - GET  /api/v1/reflector/config - Get reflector configuration
//   - POST /api/v1/reflector/config - Update reflector configuration
//   - GET  /api/v1/reflector/stats  - Get reflector statistics
//
// Test Execution:
//   - POST /api/v1/test/start  - Start a test (requires test_type parameter)
//   - POST /api/v1/test/stop   - Stop running test
//   - GET  /api/v1/test/status - Get test execution status
//
// Module Information:
//   - GET /api/v1/modules      - List all test modules
//   - GET /api/v1/modules/{n}  - Get specific module details
//
// License Management:
//   - GET  /api/v1/license     - Get license status
//   - POST /api/v1/license/activate - Activate a license key
//
// # Security
//
// CORS is restricted to localhost origins only (127.0.0.1, localhost, ::1).
// HTTP timeouts are configured to prevent slowloris and similar attacks.
// Interface names are validated before acceptance.
//
// # Static Files
//
// The React frontend is embedded via go:embed and served from the root path.
// If the embedded UI is not built, a simple HTML fallback is served.
package api

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/license"
	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/netif"
	"github.com/krisarmstrong/stem/internal/services/reflector"
	"github.com/krisarmstrong/stem/internal/version"
)

// HTTP server timeout constants.
const (
	HTTPReadHeaderTimeout = 10 * time.Second
	HTTPReadTimeout       = 30 * time.Second
	HTTPWriteTimeout      = 30 * time.Second
	HTTPIdleTimeout       = 120 * time.Second
)

// APIVersion is the current API version.
const APIVersion = "v1"

// APIVersionHeader is the header name for the API version.
const APIVersionHeader = "X-Api-Version"

const (
	defaultAuthSessionTimeout = 30 * time.Minute
	maxRequestBodySize        = 1024 * 1024 // 1 MB max request body
	shutdownTimeout           = 30 * time.Second
)

// RFC 1918 validation constants (ported from Seed for CORS).
const (
	// ipPartsClassC is the expected number of IP parts for Class C address validation.
	ipPartsClassC = 2

	// ipPartsClassAB is the expected number of IP parts for Class A/B address validation.
	ipPartsClassAB = 3

	// decimalParseBase is the base for decimal digit parsing.
	decimalParseBase = 10

	// maxIPOctetValue is the maximum valid value for an IP address octet (255).
	maxIPOctetValue = 255

	// classBMinOctet is the minimum second octet for 172.x.x.x private range.
	classBMinOctet = 16

	// classBMaxOctet is the maximum second octet for 172.x.x.x private range.
	classBMaxOctet = 31
)

//go:embed ui/*
var staticFiles embed.FS

// Server represents the web server.
type Server struct {
	port                 int
	mux                  *http.ServeMux
	sseBroadcaster       *SSEBroadcaster // Fan-out for /api/v1/events subscribers; nil-safe via getSSE()
	httpServer           *http.Server
	stats                *Stats
	statsMu              sync.RWMutex
	testStatus           string
	currentTest          string
	testResult           *TestResultResponse
	startTime            time.Time
	selectedIface        string
	mode                 string // "reflector" or "test_master"
	reflectorConfig      ReflectorConfig
	reflectorExec        *reflector.Executor // Active reflector executor (nil when not in reflector mode)
	licenseManager       *license.Manager
	authManager          *auth.Manager
	currentModule        string
	authLimiter          *RateLimiter               // Rate limiter for auth endpoints (5/min)
	apiLimiter           *RateLimiter               // Rate limiter for standard API endpoints (100/min)
	tlsConfig            TLSConfig                  // TLS configuration for HTTPS
	cookieConfig         auth.CookieConfig          // Cookie configuration for secure auth
	corsAllowPrivate     bool                       // STEM_CORS_ALLOW_PRIVATE: reflect RFC1918 cross-origins (default off)
	csrfManager          *auth.CSRFManager          // CSRF token manager for protection against CSRF attacks
	setupTokenManager    *auth.SetupTokenManager    // Setup token manager for first-time setup security
	setupComplete        bool                       // Whether initial setup has been completed
	setupModeStartTime   time.Time                  // When setup mode was activated (for timeout)
	recoveryTokenManager *auth.RecoveryTokenManager // Recovery token manager for password recovery
	dataDir              string                     // Application data directory for recovery files
	acmeChallengeServer  *http.Server               // HTTP-01 challenge server for ACME
	tlsFingerprint       tlsFingerprintCache        // Cached SHA-256 fingerprint of the active TLS cert (exposed via /__version)

	// executorResolver maps a module name to a factory producing a
	// testExecutor. If nil, defaultExecutorFactory is used. Tests inject
	// an override here to swap in a mock executor without touching the
	// real cgo dataplane.
	executorResolver func(moduleName string) (executorFactory, bool)

	// reflectorAvailability is the platform-capability probe used by
	// the POST /api/v1/mode handler to reject role switches the binary
	// cannot support (e.g. reflector on macOS / Windows pure-Go
	// builds). If nil, [defaultReflectorAvailability] is used.
	// Tests override via [Server.UseReflectorAvailabilityForTest] so
	// they can exercise the 403 path without rebuilding with
	// different cgo tags.
	reflectorAvailability reflectorAvailabilityFn
}

var (
	errMissingAuthToken  = errors.New("missing authorization token")
	errInvalidAuthHeader = errors.New("invalid authorization header")
)

// NewServer creates a new web server.
// Returns an error if required credentials are not configured via environment variables.
// getDataDir returns the application data directory.
// Uses STEM_DATA_DIR environment variable, defaults to current directory.
func getDataDir() string {
	dataDir := os.Getenv("STEM_DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	return dataDir
}

// serveFallbackUIPage handles "/" when the embedded UI sub-FS failed to
// load. Hoisted out of the registration site to keep that block flat.
func serveFallbackUIPage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>The Stem</title></head>
<body>
<h1>The Stem</h1>
<p>WebUI not built. Run 'cd ui && npm install && npm run build' first.</p>
<p>API available under <code>/api/v1/</code></p>
</body>
</html>`))
}

func NewServer(port int) (*Server, error) {
	// Initialize license manager.
	licMgr, err := license.NewManager()
	if err != nil {
		logging.Warn("Failed to initialize license manager", "error", err)
	}

	// Auto-select best interface if available.
	var defaultIface string
	best, ifaceErr := netif.GetBestInterface()
	if ifaceErr == nil {
		defaultIface = best.Name
		logging.Info("Auto-selected network interface", "interface", best.Name, "score", best.Score)
	} else {
		logging.Warn("No suitable interface found for auto-selection", "error", ifaceErr)
	}

	// Hook the auth package's HIBP soft-failure logger into our slog
	// instance. Done before NewManager so any early breach checks (e.g.
	// a credential rotation during boot) have a logger configured.
	auth.SetHIBPLogger(func(msg string, hibpErr error) {
		logging.Warn(msg, "error", hibpErr, "event", "auth.hibp.soft_failure")
	})

	// Create auth manager - credentials are required via env vars.
	authMgr, err := auth.NewManager(
		os.Getenv("STEM_JWT_SECRET"),
		defaultAuthSessionTimeout,
		os.Getenv("STEM_AUTH_USERNAME"),
		os.Getenv("STEM_AUTH_PASSWORD"),
	)
	if err != nil {
		return nil, fmt.Errorf("authentication setup failed: %w", err)
	}

	// HTTPS is required, unconditionally. Auth cookies hardcode Secure=true
	// and browsers refuse them over plain HTTP. There is no HTTP listener
	// at all — operators must use https://; typing the host without a
	// scheme will get connection refused, by design.

	s := &Server{}
	s.port = port
	s.mux = http.NewServeMux()
	s.sseBroadcaster = NewSSEBroadcaster()
	s.statsMu = sync.RWMutex{}
	s.stats = &Stats{
		PacketsReceived: 0,
		PacketsSent:     0,
		BytesReceived:   0,
		BytesSent:       0,
		CurrentPPS:      0,
		CurrentMbps:     0,
		Uptime:          0,
		TestStatus:      "",
		CurrentTest:     nil,
	}
	s.testStatus = statusIdle
	s.currentTest = ""
	s.testResult = nil
	s.startTime = time.Now()
	s.selectedIface = defaultIface
	s.mode = modeTestMaster
	s.reflectorConfig = ReflectorConfig{
		Profile:         DefaultProfile,
		SignatureFilter: nil,
		OUIFilter:       DefaultOUIFilter,
		PortFilter:      DefaultPortFilter,
	}
	s.licenseManager = licMgr
	s.authManager = authMgr
	s.currentModule = ""
	s.authLimiter = NewAuthRateLimiter()
	s.apiLimiter = NewAPIRateLimiter()
	s.tlsConfig = TLSConfig{
		Enabled:  true,
		CertFile: os.Getenv("STEM_TLS_CERT"),
		KeyFile:  os.Getenv("STEM_TLS_KEY"),
		CertsDir: os.Getenv("STEM_TLS_CERTS_DIR"),
	}
	s.cookieConfig = auth.DefaultCookieConfig()
	s.corsAllowPrivate = corsAllowPrivateEnabled()
	s.csrfManager = auth.NewCSRFManager(logging.Get())
	s.setupTokenManager = auth.NewSetupTokenManager()
	s.setupComplete = false
	s.setupModeStartTime = time.Time{}
	s.recoveryTokenManager = auth.NewRecoveryTokenManager(getDataDir())
	s.dataDir = getDataDir()
	s.setupRoutes()
	return s, nil
}

// corsAllowPrivateEnabled reports whether the operator opted into reflecting
// RFC1918 private-network origins via STEM_CORS_ALLOW_PRIVATE. Off by default:
// pairing Allow-Credentials with an arbitrary reflected LAN origin is a
// cross-origin CSRF-bypass vector, so cross-origin LAN access is opt-in.
func corsAllowPrivateEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("STEM_CORS_ALLOW_PRIVATE"))) {
	case "1", "true", "yes", "on":
		logging.Warn("CORS: STEM_CORS_ALLOW_PRIVATE enabled — reflecting RFC1918 " +
			"private-network origins with credentials; any LAN origin may make " +
			"credentialed cross-origin requests")
		return true
	default:
		return false
	}
}

// isLocalhostOrigin validates that the origin is actually localhost.
// Prevents CORS bypass via origins like "localhost.evil.com".
func isLocalhostOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// isSameOrigin checks if the Origin header matches the request's Host.
// This allows browsers to access the server from its actual address (e.g., 10.0.0.210:8444).
func isSameOrigin(origin string, requestHost string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// Compare origin host:port with request host.
	originHost := u.Host // Includes port if present.
	return originHost == requestHost
}

// isRFC1918Origin checks if the origin is an RFC 1918 private network address.
// Ported from Seed for CORS validation - allows connections from private networks.
//
// Allowed addresses:
//   - Class A private: 10.0.0.0/8 (10.x.x.x)
//   - Class B private: 172.16.0.0/12 (172.16.x.x through 172.31.x.x)
//   - Class C private: 192.168.0.0/16 (192.168.x.x)
//
// Uses proper IP validation to prevent subdomain bypass attacks.
// Rejects malicious origins like "http://192.168.1.1.evil.com".
func isRFC1918Origin(origin string) bool {
	// Reject null origin
	if origin == "null" {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := u.Hostname()
	if host == "" {
		return false
	}

	// Check for RFC 1918 private network ranges
	return isPrivateNetworkAddress(host)
}

// isPrivateNetworkAddress checks if the host is an RFC 1918 private network address.
// This prevents subdomain attacks like "192.168.1.1.evil.com" by validating
// the complete IP address structure.
func isPrivateNetworkAddress(host string) bool {
	// Class C: 192.168.0.0/16
	if strings.HasPrefix(host, "192.168.") {
		return isValidClassCAddress(host)
	}

	// Class A: 10.0.0.0/8
	if strings.HasPrefix(host, "10.") {
		return isValidClassAAddress(host)
	}

	// Class B: 172.16.0.0/12 (172.16.0.0 - 172.31.255.255)
	if strings.HasPrefix(host, "172.") {
		return isValidClassBAddress(host)
	}

	return false
}

// isValidClassCAddress validates a 192.168.x.x address.
// Returns true if the host is a valid Class C private address.
func isValidClassCAddress(host string) bool {
	remainder := host[8:] // After "192.168."
	// Should be X.Y where X and Y are 0-255
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassC {
		return false
	}
	return isValidIPOctet(parts[0]) && isValidIPOctet(parts[1])
}

// isValidClassAAddress validates a 10.x.x.x address.
// Returns true if the host is a valid Class A private address.
func isValidClassAAddress(host string) bool {
	remainder := host[3:] // After "10."
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassAB {
		return false
	}
	return isValidIPOctet(parts[0]) && isValidIPOctet(parts[1]) && isValidIPOctet(parts[2])
}

// isValidClassBAddress validates a 172.16-31.x.x address.
// Returns true if the host is a valid Class B private address (172.16.0.0/12).
func isValidClassBAddress(host string) bool {
	remainder := host[4:] // After "172."
	parts := strings.Split(remainder, ".")
	if len(parts) != ipPartsClassAB {
		return false
	}

	// Validate and parse second octet to verify range 16-31
	secondOctet, ok := parseOctetInRange(parts[0], classBMinOctet, classBMaxOctet)
	if !ok || secondOctet < classBMinOctet || secondOctet > classBMaxOctet {
		return false
	}

	return isValidIPOctet(parts[1]) && isValidIPOctet(parts[2])
}

// parseOctetInRange parses an octet string and checks if it's within the given range.
// Returns the parsed value and true if valid, 0 and false otherwise.
func parseOctetInRange(s string, minVal, maxVal int) (int, bool) {
	if s == "" || len(s) > 3 {
		return 0, false
	}

	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		val = val*decimalParseBase + int(c-'0')
		if val > maxIPOctetValue {
			return 0, false
		}
	}

	if val < minVal || val > maxVal {
		return val, false
	}

	return val, true
}

// isValidIPOctet checks if a string is a valid IP octet (0-255).
// Helper function for proper IP validation.
func isValidIPOctet(s string) bool {
	if s == "" || len(s) > 3 {
		return false
	}

	val := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		val = val*decimalParseBase + int(c-'0')
		if val > maxIPOctetValue {
			return false
		}
	}

	return true
}

// setupRoutes configures the HTTP routes.
func (s *Server) setupRoutes() {
	// API v1 routes - Health and Status (no rate limiting for health checks).
	s.handle("/api/v1/health", s.handleHealth)
	// /stats exposes operational telemetry; require auth (#340).
	s.handleAuthRateLimited("/api/v1/stats", s.handleStats, s.apiLimiter)

	// Kubernetes health probes (not versioned - infrastructure endpoints).
	s.handle("/health/live", s.handleHealthLive)
	s.handle("/health/ready", s.handleHealthReady)

	// Build metadata (universal contract — seed/stem/niac all expose this
	// unauthenticated for deployment validation). Returns lowercase JSON
	// keys: version, commit, buildTime, uiBuildHash.
	s.handle("/__version", s.handleBuildVersion)

	// Platform capabilities (unauthenticated by design — the UI calls
	// this before login to gate features like the Reflector page on
	// CGO-less builds). See handlers_capabilities.go.
	s.handle("/api/v1/capabilities", s.handleCapabilities)

	// API v1 routes - Interfaces (auth + rate limited #340 — discloses
	// the host's network interface inventory).
	s.handleAuthRateLimited("/api/v1/interfaces", s.handleInterfaces, s.apiLimiter)

	// API v1 routes - Settings and Mode (auth + rate limited #340).
	// POST /settings writes config; POST /mode flips the operating role
	// between Reflector and Test Master — both must require auth.
	s.handleAuthRateLimited("/api/v1/settings", s.handleSettings, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/mode", s.handleMode, s.apiLimiter)

	// Server-Sent Events stream for live UI updates (#296). Not rate-
	// limited — connections are long-lived, intentionally one per
	// browser-tab. Auth not required because mode-changed / reflector-
	// stats frames are observational, not credential-bearing. If we
	// ever publish privileged frames here, gate the subscribe path.
	s.mux.HandleFunc("/api/v1/events", s.handleSSEEvents)

	// API v1 routes - Test Execution (rate limited + auth).
	s.handleAuthRateLimited("/api/v1/test/start", s.handleTestStart, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/test/stop", s.handleTestStop, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/test/result", s.handleTestResult, s.apiLimiter)

	// API v1 routes - Authentication (strict rate limiting: 5/min).
	// Wave 3 (#85): the login handler now consults MFA state and may
	// return an mfa_required short-circuit instead of issuing tokens.
	s.handleRateLimited("/api/v1/auth/login", s.loginWithMFAGate, s.authLimiter)
	s.handleRateLimited("/api/v1/auth/logout", s.handleAuthLogout, s.apiLimiter)
	s.handleRateLimited("/api/v1/auth/refresh", s.handleAuthRefresh, s.authLimiter)
	// /api/v1/auth/csrf-token is the canonical path (Wave 1 #87 task);
	// /api/v1/auth/csrf is kept as an alias for older clients.
	s.handleAuthRateLimited("/api/v1/auth/csrf-token", s.handleAuthCSRF, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/auth/csrf", s.handleAuthCSRF, s.apiLimiter)

	// API v1 routes - Multi-Factor Authentication (Wave 3 #85).
	// TOTP enrolment + management endpoints require auth (the user is
	// modifying their own MFA settings). The login-finisher
	// /api/v1/auth/login/totp does NOT require auth — it presents an
	// mfa_token from the password stage as its proof of intent (and
	// is CSRF-exempt for the same reason as /api/v1/auth/login).
	s.handleAuthRateLimited("/api/v1/auth/totp/setup", s.handleTOTPSetup, s.authLimiter)
	s.handleAuthRateLimited("/api/v1/auth/totp/verify", s.handleTOTPVerify, s.authLimiter)
	s.handleAuthRateLimited("/api/v1/auth/totp/disable", s.handleTOTPDisable, s.authLimiter)
	s.handleRateLimited("/api/v1/auth/login/totp", s.handleLoginTOTP, s.authLimiter)
	s.handleAuthRateLimited("/api/v1/auth/mfa/status", s.handleMFAStatus, s.apiLimiter)
	// WebAuthn (passkey) ceremonies. Register endpoints require auth.
	// Login endpoints do not — the assertion itself proves identity.
	s.handleAuthRateLimited("/api/v1/auth/webauthn/register/begin",
		s.handleWebAuthnRegisterBegin, s.authLimiter)
	s.handleAuthRateLimited("/api/v1/auth/webauthn/register/finish",
		s.handleWebAuthnRegisterFinish, s.authLimiter)
	s.handleRateLimited("/api/v1/auth/webauthn/login/begin",
		s.handleWebAuthnLoginBegin, s.authLimiter)
	s.handleRateLimited("/api/v1/auth/webauthn/login/finish",
		s.handleWebAuthnLoginFinish, s.authLimiter)

	// API v1 routes - Setup (rate limited, no auth - for first-time setup).
	s.handleRateLimited("/api/v1/setup/status", s.handleSetupStatus, s.apiLimiter)
	s.handleRateLimited("/api/v1/setup/complete", s.handleSetupComplete, s.authLimiter)

	// API v1 routes - Password Recovery (rate limited, no auth - for recovery).
	s.handleRateLimited("/api/v1/recovery/status", s.handleRecoveryStatus, s.apiLimiter)
	s.handleRateLimited("/api/v1/recovery/complete", s.handleRecoveryComplete, s.authLimiter)
	s.handleRateLimited("/api/v1/recovery/instructions", s.handleRecoveryInstructions, s.apiLimiter)

	// API v1 routes - Reflector (authenticated + rate limited).
	// These reconfigure/inspect the dataplane, so they require a valid token —
	// previously registered with only the rate limiter (unauthenticated).
	s.handleAuthRateLimited("/api/v1/reflector/config", s.handleReflectorConfig, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/reflector/stats", s.handleReflectorStats, s.apiLimiter)

	// API v1 routes - License (rate limited).
	s.handleRateLimited("/api/v1/license", s.handleLicense, s.apiLimiter)
	s.handleRateLimited("/api/v1/license/activate", s.handleLicenseActivate, s.apiLimiter)
	s.handleRateLimited("/api/v1/license/trial", s.handleLicenseTrial, s.apiLimiter)

	// API v1 routes - Modules (rate limited).
	s.handleRateLimited("/api/v1/modules", s.handleModules, s.apiLimiter)
	s.handleRateLimited("/api/v1/modules/", s.handleModuleByName, s.apiLimiter)

	// Static files (embedded UI).
	staticFS, err := fs.Sub(staticFiles, "ui")
	if err != nil {
		logging.Warn("Could not load embedded UI", "error", err)
		s.mux.HandleFunc("/", serveFallbackUIPage)
		return
	}
	s.mux.HandleFunc("/", spaFallbackHandler(staticFS))
}

// spaFallbackHandler returns an HTTP handler that serves static files from
// the embedded FS, falling back to index.html for unknown paths so client-
// side routes survive a refresh.
func spaFallbackHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))
	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath != "" {
			if _, statErr := fs.Stat(staticFS, cleanPath); statErr == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		indexBytes, readErr := fs.ReadFile(staticFS, "index.html")
		if readErr != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(indexBytes)
	}
}

func (s *Server) handle(path string, handler http.HandlerFunc) {
	s.mux.HandleFunc(path, handler)
}

func (s *Server) handleRateLimited(path string, handler http.HandlerFunc, rl *RateLimiter) {
	s.mux.Handle(path, rl.Middleware(handler))
}

func (s *Server) handleAuthRateLimited(path string, handler http.HandlerFunc, rl *RateLimiter) {
	s.mux.Handle(path, rl.Middleware(s.authMiddleware(handler)))
}

func (s *Server) authMiddleware(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authErr := s.requireAuth(r)
		if authErr != nil {
			// Audit log the authentication failure.
			s.auditAuthFailure(r, authErr)
			s.writeAuthError(w, authErr)
			return
		}
		handler(w, r)
	})
}

func (s *Server) requireAuth(r *http.Request) error {
	// Try cookie first (most secure), then Bearer header (API client fallback).
	token, source := auth.GetTokenFromRequest(r)
	if token == "" {
		return errMissingAuthToken
	}

	_, validateErr := s.authManager.ValidateToken(r.Context(), token)
	if validateErr != nil {
		return fmt.Errorf("validate token: %w", validateErr)
	}

	// Log token source for security monitoring.
	if source == "header" {
		logging.Debug("Auth via Bearer header (API client)", "path", r.URL.Path)
	}

	return nil
}

// extractClaims parses and validates the JWT from the request, returning claims.
// Tries cookie first, then Bearer header to support browser and API clients.
func (s *Server) extractClaims(r *http.Request) (*auth.Claims, error) {
	token, _ := auth.GetTokenFromRequest(r)
	if token == "" {
		return nil, errMissingAuthToken
	}

	claims, validateErr := s.authManager.ValidateToken(r.Context(), token)
	if validateErr != nil {
		return nil, fmt.Errorf("validate token: %w", validateErr)
	}
	return claims, nil
}

func extractBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", errMissingAuthToken
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errInvalidAuthHeader
	}
	return parts[1], nil
}

func (s *Server) writeAuthError(w http.ResponseWriter, err error) {
	WriteAuthError(w, err)
}

// auditAuthFailure logs authentication failures with appropriate event types.
func (s *Server) auditAuthFailure(r *http.Request, err error) {
	switch {
	case errors.Is(err, auth.ErrTokenExpired):
		logging.AuditTokenExpired(r.Context(), r, "")
	case errors.Is(err, auth.ErrTokenRevoked):
		logging.AuditTokenRevoked(r.Context(), r, "")
	case errors.Is(err, auth.ErrInvalidToken):
		logging.AuditTokenInvalid(r.Context(), r, err.Error())
	case errors.Is(err, errMissingAuthToken):
		logging.AuditTokenInvalid(r.Context(), r, "missing authorization token")
	case errors.Is(err, errInvalidAuthHeader):
		logging.AuditTokenInvalid(r.Context(), r, "invalid authorization header format")
	default:
		logging.AuditTokenInvalid(r.Context(), r, err.Error())
	}
}

// corsMiddleware enforces CORS for API security. By default it allows only
// localhost and same-origin requests (normal UI access is same-origin).
// RFC1918 private-network origins are reflected ONLY when the operator opts in
// via STEM_CORS_ALLOW_PRIVATE — otherwise pairing Allow-Credentials with a
// reflected arbitrary LAN origin lets a hostile LAN page drive credentialed
// cross-origin requests (a CSRF-bypass vector).
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow requests without Origin header (same-origin, curl, etc.).
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Allow localhost and same-origin always; RFC1918 private-network
		// origins only when the operator opts in (STEM_CORS_ALLOW_PRIVATE).
		// Same-origin (browser accessing the server's own address, e.g.
		// https://10.0.0.210:8444) is normal UI usage and always allowed.
		allowed := isLocalhostOrigin(origin) || isSameOrigin(origin, r.Host) ||
			(s.corsAllowPrivate && isRFC1918Origin(origin))
		if !allowed {
			http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
			return
		}

		// Set CORS headers for allowed origins.
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+auth.CSRFHeaderName)
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow cookies in CORS requests
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Expose-Headers", APIVersionHeader+", "+auth.CSRFHeaderName)

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// apiVersionMiddleware adds the API version header to all API responses.
func apiVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add API version header to all API responses.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set(APIVersionHeader, APIVersion)
		}
		next.ServeHTTP(w, r)
	})
}

// Run starts the web server with graceful shutdown support.
// Listens for SIGTERM and SIGINT signals to initiate shutdown.
//
// Binding goes through bindWithFallback so a busy canonical port (8444)
// falls back to port+1..+9 instead of refusing to start (see #69).
func (s *Server) Run() error {
	// Bind first so the actual bound port is known before we announce it.
	ln, actualPort, bindErr := bindWithFallback(context.Background(), "", s.port)
	if bindErr != nil {
		return fmt.Errorf("bind web server: %w", bindErr)
	}
	addr := fmt.Sprintf(":%d", actualPort)

	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("https://localhost%s", addr),
		"version", version.GetVersion(),
	)

	// Set up signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// SSE publishers (#296). Started after ctx is created so they die
	// with the server on SIGINT/SIGTERM. The reflector-stats publisher
	// is always-on but cheap when nobody's subscribed.
	s.startReflectorStatsPublisher(ctx)

	// Wrap with middleware stack: SecurityHeaders -> CORS -> APIVersion -> RequestID -> Logging -> CSRF -> Handler.
	handler := securityHeadersMiddleware(
		s.corsMiddleware(
			apiVersionMiddleware(
				logging.RequestIDMiddleware(
					logging.Middleware(
						s.csrfManager.CSRFMiddleware(s.mux))))))
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
		ReadTimeout:       HTTPReadTimeout,
		WriteTimeout:      HTTPWriteTimeout,
		IdleTimeout:       HTTPIdleTimeout,
	}

	// Start server in goroutine. HTTPS is required, no plaintext branch.
	errChan := make(chan error, 1)
	go func() {
		listenErr := s.startTLS(ln)
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			errChan <- fmt.Errorf("server failed: %w", listenErr)
		}
		close(errChan)
	}()

	// Wait for shutdown signal or server error.
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logging.Info("Shutdown signal received, initiating graceful shutdown...")
		return s.Shutdown()
	}
}

// startTLS starts the server with TLS encryption on the already-bound
// listener. Priority order: ACME → manual certificates → self-signed.
func (s *Server) startTLS(ln net.Listener) error {
	// Priority 1: ACME/Let's Encrypt automatic certificates
	if s.tlsConfig.ACME.Enabled {
		if s.tlsConfig.ACME.Domain == "" {
			return errors.New("ACME enabled but no domain specified")
		}
		return s.startTLSWithACME(ln)
	}

	// Priority 2: Manual certificates from config
	certFile := s.tlsConfig.CertFile
	keyFile := s.tlsConfig.KeyFile

	// Priority 3: Self-signed certificate (fallback)
	if certFile == "" || keyFile == "" {
		var err error
		certFile, keyFile, err = ensureSelfSignedCert(s.tlsConfig.CertsDir)
		if err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
	}

	// Configure TLS 1.3 minimum.
	s.httpServer.TLSConfig = createTLSConfig()

	logging.Info("Starting HTTPS server",
		"addr", s.httpServer.Addr,
		"tls_version", "1.3",
		"cert_file", certFile,
	)

	listenErr := s.httpServer.ServeTLS(ln, certFile, keyFile)
	if listenErr != nil {
		return fmt.Errorf("serve TLS: %w", listenErr)
	}
	return nil
}

// startTLSWithACME starts the server with automatic Let's Encrypt certificates
// on the already-bound listener. Ported from Seed project for automatic
// certificate management.
func (s *Server) startTLSWithACME(ln net.Listener) error {
	manager, err := createACMEManager(s.tlsConfig.ACME)
	if err != nil {
		return fmt.Errorf("create ACME manager: %w", err)
	}

	// Configure TLS with ACME
	s.httpServer.TLSConfig = createACMETLSConfig(manager)

	logging.Info("Starting HTTPS server with ACME",
		"addr", s.httpServer.Addr,
		"domain", s.tlsConfig.ACME.Domain)

	// Start HTTP-01 challenge handler on port 80
	// This is required for Let's Encrypt domain validation
	challengeServer := &http.Server{
		Addr:              ":80",
		Handler:           manager.HTTPHandler(nil),
		ReadHeaderTimeout: acmeReadHeaderTimeoutSec * time.Second,
	}
	s.acmeChallengeServer = challengeServer
	go func() {
		if listenErr := s.acmeChallengeServer.ListenAndServe(); listenErr != nil &&
			!errors.Is(listenErr, http.ErrServerClosed) {
			logging.Error("ACME challenge server error", "error", listenErr)
		}
	}()

	// ServeTLS with empty cert/key paths uses GetCertificate from TLSConfig.
	if listenErr := s.httpServer.ServeTLS(ln, "", ""); listenErr != nil {
		return fmt.Errorf("https server with ACME: %w", listenErr)
	}
	return nil
}

// Shutdown gracefully shuts down the server.
// Stops running tests and drains HTTP connections.
func (s *Server) Shutdown() error {
	// Create shutdown context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop rate limiter cleanup goroutines.
	if s.authLimiter != nil {
		s.authLimiter.Stop()
	}
	if s.apiLimiter != nil {
		s.apiLimiter.Stop()
	}

	// Shutdown ACME HTTP-01 challenge server if running.
	if s.acmeChallengeServer != nil {
		logging.Info("Shutting down ACME challenge server...")
		if err := s.acmeChallengeServer.Shutdown(ctx); err != nil {
			logging.Error("Error shutting down ACME challenge server", "error", err)
		}
	}

	// Stop CSRF manager cleanup goroutine.
	if s.csrfManager != nil {
		s.csrfManager.Stop()
	}

	// Stop auth manager cleanup goroutine.
	if s.authManager != nil {
		s.authManager.Stop()
	}

	// Stop any running reflector.
	if s.reflectorExec != nil {
		logging.Info("Stopping reflector...")
		s.reflectorExec.Stop()
	}

	// Stop any running test by updating status.
	s.statsMu.Lock()
	if s.testStatus == statusRunning {
		s.testStatus = statusStopped
		logging.Info("Stopped running test due to shutdown")
	}
	s.statsMu.Unlock()

	// Shutdown HTTP server with timeout for draining connections.
	if s.httpServer != nil {
		logging.Info("Shutting down HTTP server...")
		shutdownErr := s.httpServer.Shutdown(ctx)
		if shutdownErr != nil {
			logging.Error("HTTP server shutdown error", "error", shutdownErr)
			return fmt.Errorf("shutdown failed: %w", shutdownErr)
		}
	}

	logging.Info("Server shutdown complete")
	return nil
}

// UpdateStats updates the runtime statistics (called by test runner).
func (s *Server) UpdateStats(packetsRx, packetsTx, bytesRx, bytesTx uint64, pps, mbps float64) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.stats.PacketsReceived = packetsRx
	s.stats.PacketsSent = packetsTx
	s.stats.BytesReceived = bytesRx
	s.stats.BytesSent = bytesTx
	s.stats.CurrentPPS = pps
	s.stats.CurrentMbps = mbps
}

// writeJSON encodes v as JSON and writes it to w.
// If encoding fails, it logs the error and sends a 500 response.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		logging.Error("failed to encode JSON response", "error", err)
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
	_, writeErr := w.Write(buf.Bytes())
	if writeErr != nil {
		logging.Error("failed to write JSON response", "error", writeErr)
	}
}

// safeIntToUint16 safely converts an int to uint16.
// Returns the converted value and true if in range, or 0 and false if out of range.
func safeIntToUint16(v int) (uint16, bool) {
	if v < 0 || v > math.MaxUint16 {
		return 0, false
	}
	return uint16(v), true
}

// ServeHTTP implements the [http.Handler] interface for testing purposes.
// Applies the same middleware stack used in production (API versioning, CORS, CSRF).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply middleware stack for consistent behavior in tests.
	handler := s.corsMiddleware(apiVersionMiddleware(s.csrfManager.CSRFMiddleware(s.mux)))
	handler.ServeHTTP(w, r)
}
