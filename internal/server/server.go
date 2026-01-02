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
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"sync"
	"time"

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

//go:embed dist/*
var staticFiles embed.FS

// Server represents the web server.
type Server struct {
	port            int
	mux             *http.ServeMux
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
}

// NewServer creates a new web server.
func NewServer(port int) *Server {
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
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes.
func (s *Server) setupRoutes() {
	// API routes - Health and Status.
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/stats", s.handleStats)

	// API routes - Interfaces.
	s.mux.HandleFunc("/api/interfaces", s.handleInterfaces)

	// API routes - Settings and Mode.
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/mode", s.handleMode)

	// API routes - Test Execution.
	s.mux.HandleFunc("/api/test/start", s.handleTestStart)
	s.mux.HandleFunc("/api/test/stop", s.handleTestStop)
	s.mux.HandleFunc("/api/test/result", s.handleTestResult)

	// API routes - Reflector.
	s.mux.HandleFunc("/api/reflector/config", s.handleReflectorConfig)
	s.mux.HandleFunc("/api/reflector/stats", s.handleReflectorStats)

	// API routes - License.
	s.mux.HandleFunc("/api/license", s.handleLicense)
	s.mux.HandleFunc("/api/license/activate", s.handleLicenseActivate)
	s.mux.HandleFunc("/api/license/trial", s.handleLicenseTrial)

	// API routes - Modules.
	s.mux.HandleFunc("/api/modules", s.handleModules)
	s.mux.HandleFunc("/api/modules/", s.handleModuleByName)

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

// Run starts the web server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("http://localhost%s", addr),
		"version", version.Version,
	)

	// Wrap with logging middleware.
	handler := logging.RequestIDMiddleware(logging.Middleware(s.mux))
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
		ReadTimeout:       HTTPReadTimeout,
		WriteTimeout:      HTTPWriteTimeout,
		IdleTimeout:       HTTPIdleTimeout,
	}
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
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
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		logging.Error("failed to encode JSON response", "error", err)
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
