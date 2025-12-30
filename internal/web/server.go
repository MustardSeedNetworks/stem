// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package web provides the unified HTTP server for The Stem WebUI.
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
package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/krisarmstrong/stem/internal/interfaces"
	"github.com/krisarmstrong/stem/internal/license"
	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/modules"
	"github.com/krisarmstrong/stem/internal/modules/benchmark"
	"github.com/krisarmstrong/stem/internal/modules/servicetest"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
	"github.com/krisarmstrong/stem/internal/version"
)

// Configuration constants
const (
	DefaultPortFilter = 3842
	DefaultOUIFilter  = "00:c0:17" // NetAlly OUI
	DefaultProfile    = "all"
)

// HTTP server timeout constants
const (
	HTTPReadHeaderTimeout = 10 * time.Second
	HTTPReadTimeout       = 30 * time.Second
	HTTPWriteTimeout      = 30 * time.Second
	HTTPIdleTimeout       = 120 * time.Second
)

// Response types for type safety
type (
	// StatusResponse for simple status messages
	StatusResponse struct {
		Status string `json:"status"`
	}

	// ModeResponse for mode queries
	ModeResponse struct {
		Mode string `json:"mode"`
	}

	// ModeUpdateResponse for mode updates
	ModeUpdateResponse struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}

	// SettingsResponse for settings queries
	SettingsResponse struct {
		Mode      string `json:"mode"`
		Interface string `json:"interface"`
		Theme     string `json:"theme"`
	}

	// HealthResponse for health checks
	HealthResponse struct {
		Status  string `json:"status"`
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Product string `json:"product"`
		Company string `json:"company"`
		Uptime  int64  `json:"uptime"`
	}

	// TrialStatusResponse for trial queries
	TrialStatusResponse struct {
		Active        bool `json:"active"`
		DaysRemaining int  `json:"daysRemaining,omitempty"`
	}

	// ErrorResponse for error messages
	ErrorResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	// TestStartRequest for starting a test
	TestStartRequest struct {
		TestType  string `json:"testType"`
		Interface string `json:"interface,omitempty"`
		FrameSize uint32 `json:"frameSize,omitempty"`
		Duration  int    `json:"duration,omitempty"`
	}

	// TestStartResponse for test start confirmation
	TestStartResponse struct {
		Status   string `json:"status"`
		TestType string `json:"testType"`
		Module   string `json:"module"`
		Message  string `json:"message,omitempty"`
	}

	// TestResultResponse for completed test results
	TestResultResponse struct {
		Status   string      `json:"status"`
		TestType string      `json:"testType,omitempty"`
		Module   string      `json:"module,omitempty"`
		Success  bool        `json:"success,omitempty"`
		Error    string      `json:"error,omitempty"`
		Message  string      `json:"message,omitempty"`
		Data     interface{} `json:"data,omitempty"`
	}
)

// writeJSON encodes v as JSON and writes it to w.
// If encoding fails, it logs the error and sends a 500 response.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logging.Error("failed to encode JSON response", "error", err)
		// Don't try to write error response - headers already sent
	}
}

//go:embed dist/*
var staticFiles embed.FS

// Server represents the web server
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
	licenseManager  *license.Manager
}

// ReflectorConfig holds reflector-specific settings
type ReflectorConfig struct {
	Profile         string   `json:"profile"` // netally, msn, all, custom
	SignatureFilter []string `json:"signatureFilter"`
	OUIFilter       string   `json:"ouiFilter"`
	PortFilter      int      `json:"portFilter"`
}

// Stats holds runtime statistics
type Stats struct {
	PacketsReceived uint64  `json:"packetsReceived"`
	PacketsSent     uint64  `json:"packetsSent"`
	BytesReceived   uint64  `json:"bytesReceived"`
	BytesSent       uint64  `json:"bytesSent"`
	CurrentPPS      float64 `json:"currentPps"`
	CurrentMbps     float64 `json:"currentMbps"`
	Uptime          int64   `json:"uptime"`
	TestStatus      string  `json:"testStatus"`
	CurrentTest     *string `json:"currentTest"`
}

// NewServer creates a new web server
func NewServer(port int) *Server {
	// Initialize license manager
	licMgr, err := license.NewManager()
	if err != nil {
		logging.Warn("Failed to initialize license manager", "error", err)
	}

	s := &Server{
		port:       port,
		mux:        http.NewServeMux(),
		stats:      &Stats{},
		testStatus: "idle",
		startTime:  time.Now(),
		mode:       "test_master",
		reflectorConfig: ReflectorConfig{
			Profile:    DefaultProfile,
			PortFilter: DefaultPortFilter,
			OUIFilter:  DefaultOUIFilter,
		},
		licenseManager: licMgr,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("/api/interfaces", s.handleInterfaces)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/test/start", s.handleTestStart)
	s.mux.HandleFunc("/api/test/stop", s.handleTestStop)
	s.mux.HandleFunc("/api/test/result", s.handleTestResult)
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/health", s.handleHealth)

	// Reflector-specific routes
	s.mux.HandleFunc("/api/reflector/config", s.handleReflectorConfig)
	s.mux.HandleFunc("/api/reflector/stats", s.handleReflectorStats)
	s.mux.HandleFunc("/api/mode", s.handleMode)

	// License routes
	s.mux.HandleFunc("/api/license", s.handleLicense)
	s.mux.HandleFunc("/api/license/activate", s.handleLicenseActivate)
	s.mux.HandleFunc("/api/license/trial", s.handleLicenseTrial)

	// Module routes (new module-oriented API)
	s.mux.HandleFunc("/api/modules", s.handleModules)
	s.mux.HandleFunc("/api/modules/", s.handleModuleByName)

	// Static files (embedded UI)
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		logging.Warn("Could not load embedded UI", "error", err)
		// Serve a simple fallback page
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

// handleInterfaces returns the list of network interfaces
func (s *Server) handleInterfaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ifaces, err := interfaces.DetectInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, ifaces)
}

// handleStats returns current runtime statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.RLock()
	stats := *s.stats
	stats.Uptime = int64(time.Since(s.startTime).Seconds())
	stats.TestStatus = s.testStatus
	if s.currentTest != "" {
		stats.CurrentTest = &s.currentTest
	}
	s.statsMu.RUnlock()

	writeJSON(w, stats)
}

// handleTestStart starts a test run via the module system
func (s *Server) handleTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req TestStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Default to throughput if no test type specified
	if req.TestType == "" {
		req.TestType = "throughput"
	}

	// Look up the module for this test type
	mod := modules.GetModuleForTest(req.TestType)
	if mod == nil {
		http.Error(w, fmt.Sprintf("Unknown test type: %s", req.TestType), http.StatusBadRequest)
		return
	}

	// Verify the module can run this test
	if !mod.CanRun(req.TestType) {
		http.Error(w, fmt.Sprintf("Module %s cannot run test type: %s", mod.Name(), req.TestType), http.StatusBadRequest)
		return
	}

	// Use provided interface or fall back to selected interface
	iface := req.Interface
	if iface == "" {
		iface = s.selectedIface
	}
	if iface == "" {
		http.Error(w, "No interface specified", http.StatusBadRequest)
		return
	}

	// Check if test is already running
	s.statsMu.Lock()
	if s.testStatus == "running" {
		s.statsMu.Unlock()
		http.Error(w, "Test already running", http.StatusConflict)
		return
	}
	s.testStatus = "starting"
	s.currentTest = req.TestType
	s.testResult = nil
	s.statsMu.Unlock()

	logging.Info("Starting test via module system",
		"testType", req.TestType,
		"module", mod.Name(),
		"interface", iface,
	)

	// Try to create executor and start test
	err := s.executeTest(mod.Name(), req.TestType, iface, req.FrameSize, req.Duration)
	if err != nil {
		s.statsMu.Lock()
		s.testStatus = "error"
		s.testResult = &TestResultResponse{
			Status:   "error",
			TestType: req.TestType,
			Module:   mod.Name(),
			Success:  false,
			Error:    err.Error(),
		}
		s.statsMu.Unlock()

		// Check if this is a platform limitation
		if errors.Is(err, dataplane.ErrNotSupported) {
			logging.Warn("Test execution not supported on this platform",
				"testType", req.TestType,
				"error", err,
			)
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, TestStartResponse{
				Status:   "unavailable",
				TestType: req.TestType,
				Module:   mod.Name(),
				Message:  "Test execution requires Linux with CGO support. This platform cannot execute tests.",
			})
			return
		}

		logging.Error("Failed to start test",
			"testType", req.TestType,
			"error", err,
		)
		http.Error(w, fmt.Sprintf("Failed to start test: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, TestStartResponse{
		Status:   "started",
		TestType: req.TestType,
		Module:   mod.Name(),
		Message:  "Test execution started",
	})
}

// executeTest runs the test via the appropriate module executor
func (s *Server) executeTest(moduleName, testType, iface string, frameSize uint32, duration int) error {
	switch moduleName {
	case "benchmark":
		return s.executeBenchmarkTest(testType, iface, frameSize, duration)
	case "servicetest":
		return s.executeServiceTest(testType, iface, frameSize, duration)
	default:
		return fmt.Errorf("executor not implemented for module: %s", moduleName)
	}
}

// executeBenchmarkTest runs RFC 2544 tests via the benchmark executor
func (s *Server) executeBenchmarkTest(testType, iface string, frameSize uint32, duration int) error {
	exec, err := benchmark.NewExecutor(iface)
	if err != nil {
		return fmt.Errorf("create benchmark executor: %w", err)
	}

	// Run test in goroutine
	go func() {
		defer exec.Close()

		s.statsMu.Lock()
		s.testStatus = "running"
		s.statsMu.Unlock()

		cfg := &benchmark.TestConfig{
			Interface: iface,
			FrameSize: frameSize,
			Duration:  duration,
		}

		result, err := exec.Execute(testType, cfg)

		s.statsMu.Lock()
		defer s.statsMu.Unlock()

		if err != nil {
			s.testStatus = "error"
			s.testResult = &TestResultResponse{
				Status:   "error",
				TestType: testType,
				Module:   "benchmark",
				Success:  false,
				Error:    err.Error(),
			}
			logging.Error("Benchmark test failed", "testType", testType, "error", err)
			return
		}

		s.testStatus = "completed"
		s.testResult = &TestResultResponse{
			Status:   "completed",
			TestType: testType,
			Module:   "benchmark",
			Success:  result.Success,
			Data:     result.Data,
		}
		if !result.Success {
			s.testResult.Error = result.Error
		}
		logging.Info("Benchmark test completed", "testType", testType, "success", result.Success)
	}()

	return nil
}

// executeServiceTest runs Y.1564/MEF tests via the servicetest executor
func (s *Server) executeServiceTest(testType, iface string, frameSize uint32, duration int) error {
	exec, err := servicetest.NewExecutor(iface)
	if err != nil {
		return fmt.Errorf("create servicetest executor: %w", err)
	}

	// Run test in goroutine
	go func() {
		defer exec.Close()

		s.statsMu.Lock()
		s.testStatus = "running"
		s.statsMu.Unlock()

		cfg := &servicetest.TestConfig{
			Interface: iface,
			FrameSize: frameSize,
			Duration:  duration,
		}

		result, err := exec.Execute(testType, cfg)

		s.statsMu.Lock()
		defer s.statsMu.Unlock()

		if err != nil {
			s.testStatus = "error"
			s.testResult = &TestResultResponse{
				Status:   "error",
				TestType: testType,
				Module:   "servicetest",
				Success:  false,
				Error:    err.Error(),
			}
			logging.Error("ServiceTest failed", "testType", testType, "error", err)
			return
		}

		s.testStatus = "completed"
		s.testResult = &TestResultResponse{
			Status:   "completed",
			TestType: testType,
			Module:   "servicetest",
			Success:  result.Success,
			Data:     result.Data,
		}
		if !result.Success {
			s.testResult.Error = result.Error
		}
		logging.Info("ServiceTest completed", "testType", testType, "success", result.Success)
	}()

	return nil
}

// handleTestStop stops the current test
func (s *Server) handleTestStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.Lock()
	if s.testStatus != "running" && s.testStatus != "starting" {
		s.statsMu.Unlock()
		http.Error(w, "No test running", http.StatusBadRequest)
		return
	}
	s.testStatus = "cancelled"
	testType := s.currentTest
	s.currentTest = ""
	s.statsMu.Unlock()

	logging.Info("Test cancelled", "testType", testType)
	writeJSON(w, StatusResponse{Status: "stopped"})
}

// handleTestResult returns the result of the last completed test
func (s *Server) handleTestResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.RLock()
	result := s.testResult
	status := s.testStatus
	currentTest := s.currentTest
	s.statsMu.RUnlock()

	if result != nil {
		writeJSON(w, result)
		return
	}

	// No result available, return current status
	writeJSON(w, TestResultResponse{
		Status:   status,
		TestType: currentTest,
		Message:  "No test result available",
	})
}

// SettingsUpdate for settings POST requests
type SettingsUpdate struct {
	Interface string `json:"interface,omitempty"`
	Theme     string `json:"theme,omitempty"`
}

// handleSettings handles settings get/update
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, SettingsResponse{
			Mode:      s.mode,
			Interface: s.selectedIface,
			Theme:     "system",
		})

	case http.MethodPost:
		var update SettingsUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if update.Interface != "" {
			// Validate that the interface exists
			ifaces, err := interfaces.DetectInterfaces()
			if err != nil {
				logging.Error("failed to detect interfaces for validation", "error", err)
				http.Error(w, "Failed to validate interface", http.StatusInternalServerError)
				return
			}

			found := false
			for _, iface := range ifaces {
				if iface.Name == update.Interface {
					found = true
					break
				}
			}
			if !found {
				http.Error(w, fmt.Sprintf("Interface '%s' not found", update.Interface), http.StatusBadRequest)
				return
			}

			s.selectedIface = update.Interface
			logging.Info("interface selected", "interface", update.Interface)
		}

		writeJSON(w, StatusResponse{Status: "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, HealthResponse{
		Status:  "healthy",
		Version: version.Version,
		Commit:  version.Commit,
		Product: "The Stem",
		Company: "Mustard Seed Networks",
		Uptime:  int64(time.Since(s.startTime).Seconds()),
	})
}

// UpdateStats updates the runtime statistics (called by test runner)
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

// ModeRequest for mode POST requests
type ModeRequest struct {
	Mode string `json:"mode"`
}

// handleMode handles mode switching between reflector and test_master
func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, ModeResponse{Mode: s.mode})

	case http.MethodPost:
		var req ModeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logging.Warn("mode update failed: invalid JSON", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Mode != "reflector" && req.Mode != "test_master" {
			logging.Warn("mode update failed: invalid mode", "mode", req.Mode)
			http.Error(w, "Invalid mode (must be 'reflector' or 'test_master')", http.StatusBadRequest)
			return
		}

		oldMode := s.mode
		s.mode = req.Mode
		logging.Info("mode changed", "from", oldMode, "to", s.mode)
		writeJSON(w, ModeUpdateResponse{Status: "updated", Mode: s.mode})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleReflectorConfig handles reflector configuration
func (s *Server) handleReflectorConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.reflectorConfig)

	case http.MethodPost:
		var cfg ReflectorConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			logging.Warn("reflector config update failed: invalid JSON", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate profile
		validProfiles := map[string]bool{"netally": true, "msn": true, "all": true, "custom": true}
		if cfg.Profile != "" && !validProfiles[cfg.Profile] {
			logging.Warn("reflector config update failed: invalid profile", "profile", cfg.Profile)
			http.Error(w, "Invalid profile", http.StatusBadRequest)
			return
		}

		// Track what changed for logging
		var changes []string

		// Update config
		if cfg.Profile != "" {
			s.reflectorConfig.Profile = cfg.Profile
			changes = append(changes, "profile="+cfg.Profile)
		}
		if cfg.OUIFilter != "" {
			s.reflectorConfig.OUIFilter = cfg.OUIFilter
			changes = append(changes, "ouiFilter="+cfg.OUIFilter)
		}
		if cfg.PortFilter > 0 {
			s.reflectorConfig.PortFilter = cfg.PortFilter
			changes = append(changes, fmt.Sprintf("portFilter=%d", cfg.PortFilter))
		}
		if cfg.SignatureFilter != nil {
			s.reflectorConfig.SignatureFilter = cfg.SignatureFilter
			changes = append(changes, "signatureFilter updated")
		}

		if len(changes) > 0 {
			logging.Info("reflector config updated", "changes", changes)
		}

		writeJSON(w, StatusResponse{Status: "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ReflectorStats holds reflector-specific statistics
type ReflectorStats struct {
	Running          bool    `json:"running"`
	PacketsReceived  uint64  `json:"packetsReceived"`
	PacketsReflected uint64  `json:"packetsReflected"`
	BytesReceived    uint64  `json:"bytesReceived"`
	BytesReflected   uint64  `json:"bytesReflected"`
	TxErrors         uint64  `json:"txErrors"`
	RxInvalid        uint64  `json:"rxInvalid"`
	RatePPS          float64 `json:"ratePps"`
	RateMbps         float64 `json:"rateMbps"`
	Signatures       struct {
		ProbeOT uint64 `json:"probeot"`
		DataOT  uint64 `json:"dataot"`
		Latency uint64 `json:"latency"`
		RFC2544 uint64 `json:"rfc2544"`
		Y1564   uint64 `json:"y1564"`
		MSN     uint64 `json:"msn"`
	} `json:"signatures"`
	Latency struct {
		MinUs float64 `json:"minUs"`
		AvgUs float64 `json:"avgUs"`
		MaxUs float64 `json:"maxUs"`
		Count uint64  `json:"count"`
	} `json:"latency"`
	Uptime float64 `json:"uptime"`
}

// handleReflectorStats returns reflector-specific statistics
func (s *Server) handleReflectorStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.RLock()
	elapsed := time.Since(s.startTime).Seconds()

	// Calculate rates
	pps := float64(0)
	mbps := float64(0)
	if elapsed > 0 && s.stats.PacketsSent > 0 {
		pps = float64(s.stats.PacketsSent) / elapsed
		mbps = float64(s.stats.BytesSent) * 8.0 / (elapsed * 1000000.0)
	}

	reflectorStats := ReflectorStats{
		Running:          s.mode == "reflector" && s.testStatus == "running",
		PacketsReceived:  s.stats.PacketsReceived,
		PacketsReflected: s.stats.PacketsSent,
		BytesReceived:    s.stats.BytesReceived,
		BytesReflected:   s.stats.BytesSent,
		RatePPS:          pps,
		RateMbps:         mbps,
		Uptime:           elapsed,
	}
	s.statsMu.RUnlock()

	writeJSON(w, reflectorStats)
}

// LicenseStatus represents the license status response
type LicenseStatus struct {
	Activated     bool     `json:"activated"`
	IsTrialMode   bool     `json:"isTrialMode"`
	Tier          int      `json:"tier"`
	TierName      string   `json:"tierName"`
	DaysRemaining int      `json:"daysRemaining"`
	Features      []string `json:"features"`
	DeviceHash    string   `json:"deviceHash"`
	LicenseKey    string   `json:"licenseKey,omitempty"`
	Message       string   `json:"message,omitempty"`
}

// handleLicense returns current license status
func (s *Server) handleLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.licenseManager == nil {
		writeJSON(w, LicenseStatus{
			Activated: false,
			Message:   "License manager not initialized",
		})
		return
	}

	state := s.licenseManager.GetState()
	fp := s.licenseManager.GetFingerprint()

	status := LicenseStatus{
		DeviceHash: fp.Hash(),
	}

	if state == nil {
		status.Activated = false
		status.Message = "No license. Start a trial or enter a license key."
	} else if state.IsTrialMode {
		status.Activated = true
		status.IsTrialMode = true
		status.Tier = int(license.TierTestSuite)
		status.TierName = "Trial"
		status.DaysRemaining = s.licenseManager.TrialDaysRemaining()
		status.Features = state.Features
		status.Message = fmt.Sprintf("Trial mode: %d days remaining", status.DaysRemaining)
	} else {
		status.Activated = s.licenseManager.IsActivated()
		status.IsTrialMode = false
		status.Tier = int(state.Tier)
		status.TierName = state.Tier.String()
		status.Features = state.Features
		status.LicenseKey = license.FormatKey(state.LicenseKey)
		if status.Activated {
			status.Message = fmt.Sprintf("Licensed: %s", state.Tier)
		} else {
			status.Message = "License expired or invalid"
		}
	}

	writeJSON(w, status)
}

// LicenseActivateRequest for license activation
type LicenseActivateRequest struct {
	LicenseKey string `json:"licenseKey"`
}

// handleLicenseActivate activates a license key
func (s *Server) handleLicenseActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.licenseManager == nil {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License manager not initialized",
		})
		return
	}

	var req LicenseActivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.LicenseKey == "" {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License key is required",
		})
		return
	}

	result := s.licenseManager.Activate(req.LicenseKey)
	writeJSON(w, result)
}

// handleLicenseTrial starts or checks trial status
func (s *Server) handleLicenseTrial(w http.ResponseWriter, r *http.Request) {
	if s.licenseManager == nil {
		writeJSON(w, ErrorResponse{
			Success: false,
			Message: "License manager not initialized",
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Check trial status
		if s.licenseManager.IsTrialValid() {
			writeJSON(w, TrialStatusResponse{
				Active:        true,
				DaysRemaining: s.licenseManager.TrialDaysRemaining(),
			})
		} else {
			writeJSON(w, TrialStatusResponse{Active: false})
		}

	case http.MethodPost:
		// Start trial
		result := s.licenseManager.StartTrial()
		writeJSON(w, result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleModules returns the list of all modules
func (s *Server) handleModules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	moduleInfos := modules.GetAllModuleInfos()
	writeJSON(w, map[string]interface{}{
		"modules": moduleInfos,
		"count":   len(moduleInfos),
	})
}

// handleModuleByName handles requests for specific modules
// Supports: GET /api/modules/{name} and GET /api/modules/{name}/tests
func (s *Server) handleModuleByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/modules/{name} or /api/modules/{name}/tests
	path := strings.TrimPrefix(r.URL.Path, "/api/modules/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	moduleName := parts[0]
	module := modules.GetModule(moduleName)

	if module == nil {
		http.Error(w, fmt.Sprintf("Module not found: %s", moduleName), http.StatusNotFound)
		return
	}

	// Check for /tests subpath
	if len(parts) > 1 && parts[1] == "tests" {
		// Return just the test types for this module
		writeJSON(w, map[string]interface{}{
			"module": moduleName,
			"tests":  module.TestTypes(),
			"count":  len(module.TestTypes()),
		})
		return
	}

	// Return full module info
	writeJSON(w, modules.ToInfo(module))
}

// Run starts the web server
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	logging.Info("Starting The Stem web server",
		"address", fmt.Sprintf("http://localhost%s", addr),
		"version", version.Version,
	)

	// Wrap with logging middleware
	handler := logging.RequestIDMiddleware(logging.LoggingMiddleware(s.mux))
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
		ReadTimeout:       HTTPReadTimeout,
		WriteTimeout:      HTTPWriteTimeout,
		IdleTimeout:       HTTPIdleTimeout,
	}
	return server.ListenAndServe()
}
