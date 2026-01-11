// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package server provides the unified HTTP server for The Stem WebUI.
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
// Mode is selected via the API (/api/mode) and determines which features
// are active. Both modes share the same server instance and API surface.
//
// # API Endpoints (v1)
//
// All API endpoints are versioned under /api/v1/. Legacy /api/* requests are
// redirected to /api/v1/* for backward compatibility.
// API responses include the X-API-Version header.
//
// Health and Status:
//   - GET /api/v1/health       - Server health check
//   - GET /api/v1/version      - Version information
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
package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
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
	"github.com/krisarmstrong/stem/internal/modules/reflector"
	"github.com/krisarmstrong/stem/internal/netif"
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

//go:embed dist/*
var staticFiles embed.FS

// Server represents the web server.
type Server struct {
	port                 int
	mux                  *http.ServeMux
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
	csrfManager          *auth.CSRFManager          // CSRF token manager for protection against CSRF attacks
	setupTokenManager    *auth.SetupTokenManager    // Setup token manager for first-time setup security
	setupComplete        bool                       // Whether initial setup has been completed
	setupModeStartTime   time.Time                  // When setup mode was activated (for timeout)
	recoveryTokenManager *auth.RecoveryTokenManager // Recovery token manager for password recovery
	dataDir              string                     // Application data directory for recovery files
	acmeChallengeServer  *http.Server               // HTTP-01 challenge server for ACME
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

	// Determine if TLS is enabled via environment variable.
	tlsEnabled := os.Getenv("STEM_TLS_ENABLED") != "false" // Default: enabled

	//nolint:exhaustruct // httpServer is set in Run() after creating the server.
	s := &Server{
		port:    port,
		mux:     http.NewServeMux(),
		statsMu: sync.RWMutex{},
		stats: &Stats{
			PacketsReceived: 0,
			PacketsSent:     0,
			BytesReceived:   0,
			BytesSent:       0,
			CurrentPPS:      0,
			CurrentMbps:     0,
			Uptime:          0,
			TestStatus:      "",
			CurrentTest:     nil,
		},
		testStatus:    statusIdle,
		currentTest:   "",
		testResult:    nil,
		startTime:     time.Now(),
		selectedIface: defaultIface,
		mode:          modeTestMaster,
		reflectorConfig: ReflectorConfig{
			Profile:         DefaultProfile,
			SignatureFilter: nil,
			OUIFilter:       DefaultOUIFilter,
			PortFilter:      DefaultPortFilter,
		},
		reflectorExec:  nil,
		licenseManager: licMgr,
		authManager:    authMgr,
		currentModule:  "",
		authLimiter:    NewAuthRateLimiter(),
		apiLimiter:     NewAPIRateLimiter(),
		tlsConfig: TLSConfig{
			Enabled:  tlsEnabled,
			CertFile: os.Getenv("STEM_TLS_CERT"),
			KeyFile:  os.Getenv("STEM_TLS_KEY"),
			CertsDir: os.Getenv("STEM_TLS_CERTS_DIR"),
		},
		cookieConfig:         auth.DefaultCookieConfig(tlsEnabled),
		csrfManager:          auth.NewCSRFManager(logging.Get()),
		setupTokenManager:    auth.NewSetupTokenManager(),
		setupComplete:        false,
		setupModeStartTime:   time.Time{},
		recoveryTokenManager: auth.NewRecoveryTokenManager(getDataDir()),
		dataDir:              getDataDir(),
	}
	s.setupRoutes()
	return s, nil
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
// This allows browsers to access the server from its actual address (e.g., 10.0.0.210:8080).
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
	s.handleRateLimited("/api/v1/stats", s.handleStats, s.apiLimiter)

	// Kubernetes health probes (not versioned - infrastructure endpoints).
	s.handle("/health/live", s.handleHealthLive)
	s.handle("/health/ready", s.handleHealthReady)

	// API v1 routes - Interfaces (rate limited).
	s.handleRateLimited("/api/v1/interfaces", s.handleInterfaces, s.apiLimiter)

	// API v1 routes - Settings and Mode (rate limited).
	s.handleRateLimited("/api/v1/settings", s.handleSettings, s.apiLimiter)
	s.handleRateLimited("/api/v1/mode", s.handleMode, s.apiLimiter)

	// API v1 routes - Test Execution (rate limited + auth).
	s.handleAuthRateLimited("/api/v1/test/start", s.handleTestStart, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/test/stop", s.handleTestStop, s.apiLimiter)
	s.handleAuthRateLimited("/api/v1/test/result", s.handleTestResult, s.apiLimiter)

	// API v1 routes - Authentication (strict rate limiting: 5/min).
	s.handleRateLimited("/api/v1/auth/login", s.handleAuthLogin, s.authLimiter)
	s.handleRateLimited("/api/v1/auth/logout", s.handleAuthLogout, s.apiLimiter)
	s.handleRateLimited("/api/v1/auth/refresh", s.handleAuthRefresh, s.authLimiter)
	s.handleAuthRateLimited("/api/v1/auth/csrf", s.handleAuthCSRF, s.apiLimiter)

	// API v1 routes - Setup (rate limited, no auth - for first-time setup).
	s.handleRateLimited("/api/v1/setup/status", s.handleSetupStatus, s.apiLimiter)
	s.handleRateLimited("/api/v1/setup/complete", s.handleSetupComplete, s.authLimiter)

	// API v1 routes - Password Recovery (rate limited, no auth - for recovery).
	s.handleRateLimited("/api/v1/recovery/status", s.handleRecoveryStatus, s.apiLimiter)
	s.handleRateLimited("/api/v1/recovery/complete", s.handleRecoveryComplete, s.authLimiter)
	s.handleRateLimited("/api/v1/recovery/instructions", s.handleRecoveryInstructions, s.apiLimiter)

	// API v1 routes - Reflector (rate limited).
	s.mux.Handle("/api/v1/reflector/config", s.apiLimiter.Middleware(http.HandlerFunc(s.handleReflectorConfig)))
	s.mux.Handle("/api/v1/reflector/stats", s.apiLimiter.Middleware(http.HandlerFunc(s.handleReflectorStats)))

	// API v1 routes - License (rate limited).
	s.handleRateLimited("/api/v1/license", s.handleLicense, s.apiLimiter)
	s.handleRateLimited("/api/v1/license/activate", s.handleLicenseActivate, s.apiLimiter)
	s.handleRateLimited("/api/v1/license/trial", s.handleLicenseTrial, s.apiLimiter)

	// API v1 routes - Modules (rate limited).
	s.handleRateLimited("/api/v1/modules", s.handleModules, s.apiLimiter)
	s.handleRateLimited("/api/v1/modules/", s.handleModuleByName, s.apiLimiter)

	// Backward compatibility: redirect /api/* to /api/v1/*.
	s.mux.HandleFunc("/api/", s.handleAPIRedirect)

	// Static files (embedded UI).
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		logging.Warn("Could not load embedded UI", "error", err)
		// Serve a simple fallback page.
		s.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>The Stem</title></head>
<body>
<h1>The Stem</h1>
<p>WebUI not built. Run 'cd ui && npm install && npm run build' first.</p>
<p>API available at <a href="/api/v1/health">/api/health</a></p>
</body>
</html>`))
		})
	} else {
		fileServer := http.FileServer(http.FS(staticFS))
		// Static files don't need CORS headers - they're served from the same origin.
		// CORS is handled by corsMiddleware for API endpoints.
		s.mux.Handle("/", fileServer)
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

// handleAPIRedirect redirects legacy /api/* requests to /api/v1/*.
// This provides backward compatibility for clients using unversioned API paths.
func (s *Server) handleAPIRedirect(w http.ResponseWriter, r *http.Request) {
	// Extract path after /api/.
	oldPath := r.URL.Path
	if !strings.HasPrefix(oldPath, "/api/") {
		http.NotFound(w, r)
		return
	}

	// Don't redirect if already using versioned path.
	if strings.HasPrefix(oldPath, "/api/v1/") {
		http.NotFound(w, r)
		return
	}

	// Build new versioned path.
	suffix := strings.TrimPrefix(oldPath, "/api/")
	newPath := "/api/v1/" + suffix

	// Preserve query string.
	if r.URL.RawQuery != "" {
		newPath += "?" + r.URL.RawQuery
	}

	// Use 308 Permanent Redirect to preserve HTTP method.
	http.Redirect(w, r, newPath, http.StatusPermanentRedirect)
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
// Tries cookie first, then Bearer header for API client compatibility.
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

// corsMiddleware enforces CORS for API security.
// Allows localhost origins and same-origin requests (browser accessing server's own address).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow requests without Origin header (same-origin, curl, etc.).
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Allow localhost, same-origin, or RFC 1918 private network origins.
		// Same-origin: browser accessing server from server's actual IP (e.g., http://10.0.0.210:8080).
		// RFC 1918: allows private network addresses (192.168.x.x, 10.x.x.x, 172.16-31.x.x).
		if !isLocalhostOrigin(origin) && !isSameOrigin(origin, r.Host) && !isRFC1918Origin(origin) {
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
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)

	// Determine protocol for logging.
	protocol := "http"
	if s.tlsConfig.Enabled {
		protocol = "https"
	}
	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("%s://localhost%s", protocol, addr),
		"version", version.Version,
		"tls_enabled", s.tlsConfig.Enabled,
	)

	// Set up signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Wrap with middleware stack: SecurityHeaders -> CORS -> APIVersion -> RequestID -> Logging -> CSRF -> Handler.
	handler := securityHeadersMiddleware(
		corsMiddleware(
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

	// Start server in goroutine.
	errChan := make(chan error, 1)
	go func() {
		var listenErr error
		if s.tlsConfig.Enabled {
			listenErr = s.startTLS()
		} else {
			listenErr = s.httpServer.ListenAndServe()
		}
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

// startTLS starts the server with TLS encryption.
// Priority order: ACME → manual certificates → self-signed.
func (s *Server) startTLS() error {
	// Priority 1: ACME/Let's Encrypt automatic certificates
	if s.tlsConfig.ACME.Enabled {
		if s.tlsConfig.ACME.Domain == "" {
			return errors.New("ACME enabled but no domain specified")
		}
		return s.startTLSWithACME()
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

	listenErr := s.httpServer.ListenAndServeTLS(certFile, keyFile)
	if listenErr != nil {
		return fmt.Errorf("listen and serve TLS: %w", listenErr)
	}
	return nil
}

// startTLSWithACME starts the server with automatic Let's Encrypt certificates.
// Ported from Seed project for automatic certificate management.
func (s *Server) startTLSWithACME() error {
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
	//nolint:exhaustruct // Only set required http.Server fields
	s.acmeChallengeServer = &http.Server{
		Addr:              ":80",
		Handler:           manager.HTTPHandler(nil),
		ReadHeaderTimeout: acmeReadHeaderTimeoutSec * time.Second,
	}
	go func() {
		if listenErr := s.acmeChallengeServer.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logging.Error("ACME challenge server error", "error", listenErr)
		}
	}()

	// ListenAndServeTLS with empty cert/key paths uses GetCertificate from TLSConfig
	if listenErr := s.httpServer.ListenAndServeTLS("", ""); listenErr != nil {
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

// decodeJSONStrict decodes JSON from the request body with size limits and strict validation.
// Returns false if decoding fails (error response already written to w).
//
//nolint:unparam // maxSize allows future per-endpoint customization
func decodeJSONStrict(w http.ResponseWriter, r *http.Request, v any, maxSize int64) bool {
	// Limit request body size.
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(v)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			WriteAPIError(w, ErrAPIRequestTooLarge)
			return false
		}
		logging.Warn("JSON decode failed", "error", err)
		// Return sanitized error message - don't expose internal JSON parsing details.
		WriteInvalidRequest(w, "Invalid JSON in request body")
		return false
	}
	return true
}

// safeIntToUint16 safely converts an int to uint16.
// Returns the converted value and true if in range, or 0 and false if out of range.
func safeIntToUint16(v int) (uint16, bool) {
	if v < 0 || v > math.MaxUint16 {
		return 0, false
	}
	return uint16(v), true
}

// ServeHTTP implements the http.Handler interface for testing purposes.
// Applies the same middleware stack used in production (API versioning, CORS, CSRF).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply middleware stack for consistent behavior in tests.
	handler := corsMiddleware(apiVersionMiddleware(s.csrfManager.CSRFMiddleware(s.mux)))
	handler.ServeHTTP(w, r)
}
