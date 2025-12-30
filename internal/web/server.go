// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package web provides the HTTP server for The Stem WebUI.
//
// The server embeds the React frontend build and provides REST API endpoints
// for interface detection, license management, test execution, and real-time
// statistics monitoring via polling.
package web

import (
	"embed"
	"encoding/json"
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
	"github.com/krisarmstrong/stem/internal/version"
)

// Configuration constants
const (
	DefaultPortFilter = 3842
	DefaultOUIFilter  = "00:c0:17" // NetAlly OUI
	DefaultProfile    = "all"
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
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html>
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

// handleTestStart starts a test run
func (s *Server) handleTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.Lock()
	if s.testStatus == "running" {
		s.statsMu.Unlock()
		http.Error(w, "Test already running", http.StatusConflict)
		return
	}
	s.testStatus = "running"
	s.currentTest = "throughput" // Default test
	s.statsMu.Unlock()

	writeJSON(w, StatusResponse{Status: "started"})
}

// handleTestStop stops the current test
func (s *Server) handleTestStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.Lock()
	if s.testStatus != "running" {
		s.statsMu.Unlock()
		http.Error(w, "No test running", http.StatusBadRequest)
		return
	}
	s.testStatus = "completed"
	s.currentTest = ""
	s.statsMu.Unlock()

	writeJSON(w, StatusResponse{Status: "stopped"})
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
			s.selectedIface = update.Interface
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
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Mode != "reflector" && req.Mode != "test_master" {
			http.Error(w, "Invalid mode (must be 'reflector' or 'test_master')", http.StatusBadRequest)
			return
		}

		s.mode = req.Mode
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
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate profile
		validProfiles := map[string]bool{"netally": true, "msn": true, "all": true, "custom": true}
		if cfg.Profile != "" && !validProfiles[cfg.Profile] {
			http.Error(w, "Invalid profile", http.StatusBadRequest)
			return
		}

		// Update config
		if cfg.Profile != "" {
			s.reflectorConfig.Profile = cfg.Profile
		}
		if cfg.OUIFilter != "" {
			s.reflectorConfig.OUIFilter = cfg.OUIFilter
		}
		if cfg.PortFilter > 0 {
			s.reflectorConfig.PortFilter = cfg.PortFilter
		}
		if cfg.SignatureFilter != nil {
			s.reflectorConfig.SignatureFilter = cfg.SignatureFilter
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
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return server.ListenAndServe()
}
