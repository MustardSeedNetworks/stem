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
// # API Endpoints
//
// Health and Status:
//   - GET /api/health       - Server health check
//   - GET /api/version      - Version information
//
// Mode Management:
//   - GET  /api/mode        - Get current operating mode
//   - POST /api/mode        - Set operating mode (reflector/test_master)
//
// Interface Management:
//   - GET  /api/interfaces  - List available network interfaces
//   - GET  /api/settings    - Get current settings (interface, mode)
//   - POST /api/settings    - Update settings (validates interface exists)
//
// Reflector Mode:
//   - GET  /api/reflector/config - Get reflector configuration
//   - POST /api/reflector/config - Update reflector configuration
//   - GET  /api/reflector/stats  - Get reflector statistics
//
// Test Execution:
//   - POST /api/test/start  - Start a test (requires test_type parameter)
//   - POST /api/test/stop   - Stop running test
//   - GET  /api/test/status - Get test execution status
//
// Module Information:
//   - GET /api/modules      - List all test modules
//   - GET /api/modules/{n}  - Get specific module details
//
// License Management:
//   - GET  /api/license     - Get license status
//   - POST /api/license/activate - Activate a license key
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

	"github.com/gorilla/websocket"

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

const (
	defaultAuthSessionTimeout = 30 * time.Minute
	wsWriteTimeout            = 5 * time.Second
	wsPingInterval            = 30 * time.Second
	wsPongTimeout             = 10 * time.Second
	defaultWSBufferSize       = 1024
	maxRequestBodySize        = 1024 * 1024 // 1 MB max request body
	shutdownTimeout           = 30 * time.Second
)

//go:embed dist/*
var staticFiles embed.FS

// Server represents the web server.
type Server struct {
	port            int
	mux             *http.ServeMux
	httpServer      *http.Server
	stats           *Stats
	statsMu         sync.RWMutex
	testStatus      string
	currentTest     string
	testResult      *TestResultResponse
	startTime       time.Time
	selectedIface   string
	mode            string // "reflector" or "test_master"
	reflectorConfig ReflectorConfig
	reflectorExec   *reflector.Executor // Active reflector executor (nil when not in reflector mode)
	licenseManager  *license.Manager
	authManager     *auth.Manager
	currentModule   string
	wsClients       sync.Map // map[*websocket.Conn]struct{} - concurrent-safe
	wsUpgrader      websocket.Upgrader
}

var (
	errMissingAuthToken  = errors.New("missing authorization token")
	errInvalidAuthHeader = errors.New("invalid authorization header")
)

// NewServer creates a new web server.
// Returns an error if required credentials are not configured via environment variables.
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
		wsClients:      sync.Map{},
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  defaultWSBufferSize,
			WriteBufferSize: defaultWSBufferSize,
			WriteBufferPool: nil,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return isLocalhostOrigin(origin)
			},
			HandshakeTimeout:  0,
			Subprotocols:      nil,
			Error:             nil,
			EnableCompression: false,
		},
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

// setupRoutes configures the HTTP routes.
func (s *Server) setupRoutes() {
	// API routes - Health and Status.
	s.handle("/api/health", s.handleHealth)
	s.handle("/api/stats", s.handleStats)

	// API routes - Interfaces.
	s.handle("/api/interfaces", s.handleInterfaces)

	// API routes - Settings and Mode.
	s.handle("/api/settings", s.handleSettings)
	s.handle("/api/mode", s.handleMode)

	// API routes - Test Execution.
	s.handleAuth("/api/test/start", s.handleTestStart)
	s.handleAuth("/api/test/stop", s.handleTestStop)
	s.handleAuth("/api/test/result", s.handleTestResult)

	// API routes - Authentication.
	s.handle("/api/auth/login", s.handleAuthLogin)
	s.handle("/api/auth/logout", s.handleAuthLogout)
	s.handle("/api/auth/refresh", s.handleAuthRefresh)
	s.mux.HandleFunc("/api/ws/test-results", s.handleTestResultsWebSocket)

	// API routes - Reflector.
	s.mux.HandleFunc("/api/reflector/config", s.handleReflectorConfig)
	s.mux.HandleFunc("/api/reflector/stats", s.handleReflectorStats)

	// API routes - License.
	s.handle("/api/license", s.handleLicense)
	s.handle("/api/license/activate", s.handleLicenseActivate)
	s.handle("/api/license/trial", s.handleLicenseTrial)

	// API routes - Modules.
	s.handle("/api/modules", s.handleModules)
	s.handle("/api/modules/", s.handleModuleByName)

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
<p>API available at <a href="/api/health">/api/health</a></p>
</body>
</html>`))
		})
	} else {
		fileServer := http.FileServer(http.FS(staticFS))
		s.mux.Handle("/", fileServer)
	}
}

func (s *Server) handle(path string, handler http.HandlerFunc) {
	s.mux.HandleFunc(path, handler)
}

func (s *Server) handleAuth(path string, handler http.HandlerFunc) {
	s.mux.Handle(path, s.authMiddleware(handler))
}

func (s *Server) authMiddleware(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authErr := s.requireAuth(r)
		if authErr != nil {
			s.writeAuthError(w, authErr)
			return
		}
		handler(w, r)
	})
}

func (s *Server) requireAuth(r *http.Request) error {
	token, err := extractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return err
	}
	_, validateErr := s.authManager.ValidateToken(r.Context(), token)
	if validateErr != nil {
		return fmt.Errorf("validate token: %w", validateErr)
	}

	return nil
}

// extractClaims parses and validates the JWT from the request, returning claims.
func (s *Server) extractClaims(r *http.Request) (*auth.Claims, error) {
	token, err := extractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return nil, err
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
	status := http.StatusUnauthorized
	message := "Unauthorized"
	isKnownErr := errors.Is(err, auth.ErrInvalidToken) ||
		errors.Is(err, auth.ErrTokenExpired) ||
		errors.Is(err, auth.ErrTokenRevoked)
	if isKnownErr {
		message = err.Error()
	}
	http.Error(w, message, status)
}

// corsMiddleware enforces localhost-only CORS for API security.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow requests without Origin header (same-origin, curl, etc.).
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Only allow localhost origins.
		if !isLocalhostOrigin(origin) {
			http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
			return
		}

		// Set CORS headers for allowed origins.
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Run starts the web server with graceful shutdown support.
// Listens for SIGTERM and SIGINT signals to initiate shutdown.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("http://localhost%s", addr),
		"version", version.Version,
	)

	// Set up signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Wrap with middleware stack: CORS -> RequestID -> Logging -> Handler.
	handler := corsMiddleware(logging.RequestIDMiddleware(logging.Middleware(s.mux)))
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
		listenErr := s.httpServer.ListenAndServe()
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

// Shutdown gracefully shuts down the server.
// Closes WebSocket connections, stops running tests, and drains HTTP connections.
func (s *Server) Shutdown() error {
	// Create shutdown context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Close all WebSocket connections.
	s.closeAllWebSockets()

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

// closeAllWebSockets closes all active WebSocket connections.
func (s *Server) closeAllWebSockets() {
	var count int
	s.wsClients.Range(func(_, _ any) bool {
		count++
		return true
	})

	if count > 0 {
		logging.Info("Closing WebSocket connections", "count", count)
	}

	s.wsClients.Range(func(key, _ any) bool {
		if conn, ok := key.(*websocket.Conn); ok {
			_ = conn.Close()
			s.wsClients.Delete(conn)
		}
		return true
	})
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
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return false
		}
		logging.Warn("JSON decode failed", "error", err)
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func (s *Server) publishTestState(status, module, testType string, resp *TestResultResponse) {
	if resp != nil {
		s.broadcastTestEvent(resp)
		return
	}
	s.broadcastTestEvent(&TestResultResponse{
		Status:   status,
		Module:   module,
		TestType: testType,
		Success:  false,
		Error:    "",
		Message:  "",
		Data:     nil,
	})
}

func (s *Server) broadcastTestEvent(resp *TestResultResponse) {
	if resp == nil {
		return
	}

	s.wsClients.Range(func(key, _ any) bool {
		if conn, ok := key.(*websocket.Conn); ok {
			event := copyTestResultResponse(resp)
			go s.writeTestEvent(conn, event)
		}
		return true
	})
}

func copyTestResultResponse(resp *TestResultResponse) *TestResultResponse {
	if resp == nil {
		return nil
	}
	respCopy := *resp
	return &respCopy
}

func (s *Server) writeTestEvent(conn *websocket.Conn, resp *TestResultResponse) {
	if resp == nil {
		return
	}
	deadlineErr := conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	if deadlineErr != nil {
		logging.Warn("failed to set websocket write deadline", "error", deadlineErr)
	}
	writeErr := conn.WriteJSON(resp)
	if writeErr != nil {
		logging.Warn("websocket client write failed", "error", writeErr)
		s.unregisterWSClient(conn)
	}
}

func (s *Server) registerWSClient(conn *websocket.Conn) {
	s.wsClients.Store(conn, struct{}{})
}

func (s *Server) unregisterWSClient(conn *websocket.Conn) {
	s.wsClients.Delete(conn)
	_ = conn.Close()
}

func (s *Server) sendCurrentTestState(conn *websocket.Conn) {
	resp := s.snapshotCurrentTest()
	s.writeTestEvent(conn, resp)
}

func (s *Server) snapshotCurrentTest() *TestResultResponse {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	if s.testResult != nil {
		resultCopy := *s.testResult
		return &resultCopy
	}
	return &TestResultResponse{
		Status:   s.testStatus,
		TestType: s.currentTest,
		Module:   s.currentModule,
		Success:  false,
		Error:    "",
		Message:  "",
		Data:     nil,
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

// ServeHTTP implements the http.Handler interface for testing purposes.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
