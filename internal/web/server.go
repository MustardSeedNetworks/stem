// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package web provides the HTTP server for the Seed Test Suite WebUI.
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
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/krisarmstrong/stem/internal/interfaces"
	"github.com/krisarmstrong/stem/internal/license"
)

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
		log.Printf("Warning: Failed to initialize license manager: %v", err)
	}

	s := &Server{
		port:       port,
		mux:        http.NewServeMux(),
		stats:      &Stats{},
		testStatus: "idle",
		startTime:  time.Now(),
		mode:       "test_master",
		reflectorConfig: ReflectorConfig{
			Profile:    "all",
			PortFilter: 3842,
			OUIFilter:  "00:c0:17",
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

	// Static files (embedded UI)
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		log.Printf("Warning: Could not load embedded UI: %v", err)
		// Serve a simple fallback page
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Seed Test Suite</title></head>
<body>
<h1>Seed Test Suite</h1>
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ifaces)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleSettings handles settings get/update
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return current settings
		settings := map[string]interface{}{
			"mode":      "test_master",
			"interface": s.selectedIface,
			"theme":     "system",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)

	case http.MethodPost:
		// Update settings
		var settings map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if iface, ok := settings["interface"].(string); ok {
			s.selectedIface = iface
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

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

	health := map[string]interface{}{
		"status":  "healthy",
		"version": "3.0.0-dev",
		"product": "Seed Test Suite",
		"company": "Mustard Seed Networks",
		"uptime":  int64(time.Since(s.startTime).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
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

// handleMode handles mode switching between reflector and test_master
func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(map[string]string{"mode": s.mode})

	case http.MethodPost:
		var req struct {
			Mode string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Mode != "reflector" && req.Mode != "test_master" {
			http.Error(w, "Invalid mode (must be 'reflector' or 'test_master')", http.StatusBadRequest)
			return
		}

		s.mode = req.Mode
		json.NewEncoder(w).Encode(map[string]string{"status": "updated", "mode": s.mode})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleReflectorConfig handles reflector configuration
func (s *Server) handleReflectorConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(s.reflectorConfig)

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

		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reflectorStats)
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

	w.Header().Set("Content-Type", "application/json")

	if s.licenseManager == nil {
		json.NewEncoder(w).Encode(LicenseStatus{
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

	json.NewEncoder(w).Encode(status)
}

// handleLicenseActivate activates a license key
func (s *Server) handleLicenseActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.licenseManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "License manager not initialized",
		})
		return
	}

	var req struct {
		LicenseKey string `json:"licenseKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.LicenseKey == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "License key is required",
		})
		return
	}

	result := s.licenseManager.Activate(req.LicenseKey)
	json.NewEncoder(w).Encode(result)
}

// handleLicenseTrial starts or checks trial status
func (s *Server) handleLicenseTrial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.licenseManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "License manager not initialized",
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Check trial status
		if s.licenseManager.IsTrialValid() {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"active":        true,
				"daysRemaining": s.licenseManager.TrialDaysRemaining(),
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"active": false,
			})
		}

	case http.MethodPost:
		// Start trial
		result := s.licenseManager.StartTrial()
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Run starts the web server
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting Seed Test Suite web server on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.mux)
}
