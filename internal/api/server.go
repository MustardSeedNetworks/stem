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
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"
	"github.com/MustardSeedNetworks/foundation/pkg/httpserver"
	"github.com/MustardSeedNetworks/foundation/pkg/instance"

	"github.com/MustardSeedNetworks/stem/internal/api/cors"
	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
	"github.com/MustardSeedNetworks/stem/internal/api/sse"
	"github.com/MustardSeedNetworks/stem/internal/api/tlsutil"
	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	"github.com/MustardSeedNetworks/stem/internal/services/reflector"
	"github.com/MustardSeedNetworks/stem/internal/version"
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

//go:embed ui/*
var staticFiles embed.FS

// Server represents the web server.
type Server struct {
	port                 int
	handler              http.Handler     // every route through the Registrar, inside the global headers (setupRoutes)
	sseBroadcaster       *sse.Broadcaster // Fan-out for /api/v1/events subscribers; nil when SSE not available.
	httpServer           *http.Server
	stats                *Stats
	statsMu              sync.RWMutex
	testRunMu            sync.Mutex
	reflectorMu          sync.Mutex
	testStatus           string
	testRunID            uint64
	activeTestExec       testExecutor
	currentTest          string
	currentRunID         string    // ID naming the run in flight, for runs with no plan (#1193)
	runStartedAt         time.Time // When the run in flight began; the daemon's own clock (#1333)
	testResult           *TestResultResponse
	testError            string // classified cause of the failed run, reported on /api/v1/stats
	runPlan              *runPlan
	startTime            time.Time
	selectedIface        string
	mode                 string // "reflector" or "test_master"
	reflectorConfig      ReflectorConfig
	reflectorExec        *reflector.Executor // Active reflector executor (nil when not in reflector mode)
	licenseManager       *license.Manager
	authManager          *auth.Manager
	currentModule        string
	authLimiter          *ratelimit.RateLimiter     // Rate limiter for auth endpoints (5/min)
	auditor              *logging.Auditor           // Owns the failed-login tracker and its cleanup loop
	apiLimiter           *ratelimit.RateLimiter     // Rate limiter for standard API endpoints (100/min)
	listenConfig         httpserver.Config          // the one HTTPS listener: port fallback, TLS, same-port redirect
	cookieConfig         auth.CookieConfig          // Cookie configuration for secure auth
	corsAllowPrivate     bool                       // STEM_CORS_ALLOW_PRIVATE: reflect RFC1918 cross-origins (default off)
	trustedProxies       []netip.Prefix             // STEM_TRUSTED_PROXIES: hops whose X-Forwarded-For may key security counters (default none)
	csrfManager          *csrf.Manager              // per-session CSRF tokens, enforced per route by the Registrar
	setupTokenManager    *auth.SetupTokenManager    // Setup token manager for first-time setup security
	setupMu              sync.Mutex                 // Serializes first-run setup so only one claim wins
	recoveryTokenManager *auth.RecoveryTokenManager // Recovery token manager for password recovery
	dataDir              string                     // Application data directory for recovery files
	instanceLock         *instance.Lock             // Single-instance lock on dataDir, held for the lifetime of Run (#1336)
	publishedURL         atomic.Pointer[string]     // base URL this daemon published for the local CLI (#1166), nil when none
	runInstanceID        string                     // per-daemon segment of every run ID (#1166)
	tlsFingerprint       tlsutil.FingerprintCache   // Cached SHA-256 fingerprint of the active TLS cert (exposed via /__version)
	background           *BackgroundComponents      // Run-scoped long-lived goroutines (reflector-stats SSE publisher); ordered Start/Stop (background.go)

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

// NewServer builds a Server bound to port: it loads the license manager,
// auto-selects a network interface, and resolves the administrator
// credential (see newAuthManager), returning an error when the credential
// store is damaged or the environment names only half a credential. The
// returned Server has not started listening yet; call Run to bind the TLS
// listener and serve.
func NewServer(port int) (*Server, error) {
	// Initialize license manager. A state Stem cannot use is reported once
	// here rather than on every gated request; the entitlement consequence is
	// hasFeature's, which grants only the Free features without a manager.
	licMgr, err := license.Load()
	if err != nil {
		logging.Error("license manager unavailable; entitlements limited to Free",
			"event", "license.unusable", "error", err)
	} else if licStatus := licMgr.LoadStatus(); !licStatus.Usable() {
		logging.Error("license file unusable; entitlements limited to Free",
			"event", "license.unusable", "status", licStatus.String(),
			"path", license.DefaultLicensePath(), "error", licMgr.LoadError())
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

	authMgr, err := newAuthManager(getDataDir())
	if err != nil {
		return nil, fmt.Errorf("authentication setup failed: %w", err)
	}

	trustedProxies, err := trustedProxiesFromEnv()
	if err != nil {
		return nil, err
	}

	// HTTPS is required, unconditionally. Auth cookies hardcode Secure=true
	// and browsers refuse them over plain HTTP. The one listener answers a
	// plaintext request with a 308 to https on the same port and serves it
	// nothing else (httpserver.Listen).

	s := &Server{}
	s.port = port
	s.sseBroadcaster = sse.New()
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
	s.testError = ""
	s.currentTest = ""
	s.currentRunID = ""
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
	s.trustedProxies = trustedProxies
	s.authLimiter = ratelimit.NewAuthRateLimiter(trustedProxies)
	s.auditor = logging.NewAuditor(trustedProxies)
	s.apiLimiter = ratelimit.NewAPIRateLimiter(trustedProxies)
	s.listenConfig = httpserver.Config{
		Addr:     fmt.Sprintf(":%d", port),
		CertFile: os.Getenv("STEM_TLS_CERT"),
		KeyFile:  os.Getenv("STEM_TLS_KEY"),
		CertDir:  os.Getenv("STEM_TLS_CERTS_DIR"),
		Cert: httpserver.CertOptions{
			CommonName: "The Stem Self-Signed",
			DNSNames:   []string{"localhost", "stem.local"},
		},
		Logger: logging.Get(),
	}
	s.cookieConfig = auth.DefaultCookieConfig()
	s.corsAllowPrivate = corsAllowPrivateEnabled()
	s.csrfManager = csrf.NewManager()
	s.setupTokenManager = auth.NewSetupTokenManager()
	s.recoveryTokenManager = auth.NewRecoveryTokenManager(getDataDir())
	s.dataDir = getDataDir()

	if stateErr := s.initStateFromDataDir(); stateErr != nil {
		return nil, stateErr
	}
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

// setupRoutes registers every route on one Registrar and wraps it in the
// headers every response carries.
func (s *Server) setupRoutes() {
	reg := s.newRegistrar()
	reg.RegisterAll(s.routes(reg))
	s.handler = securityHeadersMiddleware(s.corsMiddleware(apiVersionMiddleware(reg.Handler())))
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

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authErr := s.requireAuth(r)
		if authErr != nil {
			// Audit log the authentication failure.
			s.auditAuthFailure(r, authErr)
			s.writeAuthError(w, authErr)
			return
		}
		next.ServeHTTP(w, r)
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
		allowed := cors.IsLocalhostOrigin(origin) || cors.IsSameOrigin(origin, r.Host) ||
			(s.corsAllowPrivate && cors.IsRFC1918Origin(origin))
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
// The listener is foundation's: a busy canonical port (8444) falls back to
// port+1..+9 instead of refusing to start (see #69), and a plaintext request
// on it is answered with a 308 to https.
func (s *Server) Run() error {
	// One daemon per data directory (#1336). The port is not the guard: the
	// +1..+9 fallback below means a second `stem web` does not collide, it
	// binds a neighbour and then shares this directory's state and
	// descriptor with the first. The lock is taken before the bind because
	// a refused start must not have touched anything.
	lock, lockErr := instance.Acquire(s.dataDir)
	if lockErr != nil {
		return lockErr
	}
	s.instanceLock = lock

	// Bind first so the actual bound port is known before we announce it.
	ln, listenErr := httpserver.Listen(context.Background(), s.listenConfig)
	if listenErr != nil {
		return fmt.Errorf("start HTTPS listener: %w", listenErr)
	}
	tcpAddr, isTCP := ln.Addr().(*net.TCPAddr)
	if !isTCP {
		_ = ln.Close()
		return fmt.Errorf("HTTPS listener bound a non-TCP address %s", ln.Addr())
	}
	actualPort := tcpAddr.Port
	addr := fmt.Sprintf(":%d", actualPort)

	// Record the port in the lock so the next `stem web` can name where the
	// holder actually ended up rather than where it was asked to go.
	if portErr := lock.SetPort(actualPort); portErr != nil {
		logging.Warn("could not record the bound port in the instance lock",
			"error", portErr.Error())
	}

	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("https://localhost%s", addr),
		"version", version.GetVersion(),
	)

	// Set up signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Background goroutines (#296). Started after ctx is created so they die
	// with the server on SIGINT/SIGTERM; Shutdown also stops them explicitly
	// and waits for them to exit. The reflector-stats publisher is always-on
	// but cheap when nobody's subscribed.
	// The local CLI is a client of this daemon (#1166); publish how to
	// reach it before the listener accepts, and withdraw it in Shutdown.
	// This is after the bind so the URL names the port actually bound
	// rather than the one that was asked for.
	if connErr := s.publishConnection(fmt.Sprintf("https://localhost:%d", actualPort)); connErr != nil {
		return fmt.Errorf("publish daemon descriptor: %w", connErr)
	}

	s.background = newBackgroundComponents(s)
	s.background.Start(ctx)

	s.autostartReflector()

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
		ReadTimeout:       HTTPReadTimeout,
		WriteTimeout:      HTTPWriteTimeout,
		IdleTimeout:       HTTPIdleTimeout,
	}

	logging.Info("Starting HTTPS server", "addr", addr, "cert_file", s.activeCertPath())

	// ln already speaks TLS, so Serve rather than ServeTLS.
	errChan := make(chan error, 1)
	go func() {
		serveErr := s.httpServer.Serve(ln)
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errChan <- fmt.Errorf("server failed: %w", serveErr)
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

// Shutdown gracefully shuts down the server.
// Stops running tests and drains HTTP connections.
func (s *Server) Shutdown() error {
	// Create shutdown context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop Run-scoped background goroutines first and wait for them to exit,
	// so the reflector-stats publisher stops reading the reflector executor
	// before we tear it down below. No-op when Run was never called.
	if s.background != nil {
		s.background.Stop()
	}

	// Withdraw the local CLI credential before anything else: a descriptor
	// that outlives the daemon points a CLI at whatever binds the port next.
	s.withdrawConnection()

	// Hand the data directory back after the descriptor is gone, so the next
	// daemon never finds the lock free and a stale credential still present.
	if s.instanceLock != nil {
		if err := s.instanceLock.Release(); err != nil {
			logging.Warn("failed to release the instance lock", "error", err.Error())
		}
		s.instanceLock = nil
	}

	// Stop rate limiter cleanup goroutines.
	if s.authLimiter != nil {
		s.authLimiter.Stop()
	}

	// Stop the failed-login tracker cleanup goroutine.
	if s.auditor != nil {
		s.auditor.Stop()
	}
	if s.apiLimiter != nil {
		s.apiLimiter.Stop()
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
		s.markStoppedLocked()
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

// ServeHTTP implements [http.Handler] with the exact handler Run serves, so
// tests exercise the production composition.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
