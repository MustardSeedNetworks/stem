// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package web provides the embedded web server for the Reflector.
//
// Serves the React UI and provides JSON API endpoints for real-time
// packet statistics, signature filter configuration, and interface status.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/krisarmstrong/stem/internal/reflector/dataplane"
)

// Response types for type safety
type (
	// StatusResponse for simple status messages
	StatusResponse struct {
		Status string `json:"status"`
	}

	// HealthResponse for health checks
	HealthResponse struct {
		Status  string  `json:"status"`
		Running bool    `json:"running"`
		Uptime  float64 `json:"uptime"`
		Version string  `json:"version"`
	}
)

// setCORSHeaders sets CORS headers for local development.
// Restricts to localhost origins for security.
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// Only allow localhost origins (various ports for dev servers)
	if origin == "" ||
		strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "http://127.0.0.1") ||
		strings.HasPrefix(origin, "https://localhost") ||
		strings.HasPrefix(origin, "https://127.0.0.1") {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		// For same-origin requests (no Origin header), no CORS header needed
	}
	// Don't set header for non-localhost origins (blocked by browser)
}

// writeJSON encodes v as JSON and writes it to w.
// If encoding fails, it logs the error.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

//go:embed dist/*
var reactApp embed.FS

// Server holds the web server state
type Server struct {
	dp        *dataplane.Dataplane
	port      int
	startTime time.Time
	mux       *http.ServeMux
}

// StatsResponse is the JSON structure for stats API
type StatsResponse struct {
	Timestamp        string  `json:"timestamp"`
	Uptime           float64 `json:"uptime_seconds"`
	Interface        string  `json:"interface"`
	Running          bool    `json:"running"`
	PacketsReceived  uint64  `json:"packets_received"`
	PacketsReflected uint64  `json:"packets_reflected"`
	BytesReceived    uint64  `json:"bytes_received"`
	BytesReflected   uint64  `json:"bytes_reflected"`
	TxErrors         uint64  `json:"tx_errors"`
	RxInvalid        uint64  `json:"rx_invalid"`
	RatePackets      float64 `json:"rate_pps"`
	RateMbps         float64 `json:"rate_mbps"`
	Signatures       struct {
		ProbeOT uint64 `json:"probeot"`
		DataOT  uint64 `json:"dataot"`
		Latency uint64 `json:"latency"`
		RFC2544 uint64 `json:"rfc2544"`
		Y1564   uint64 `json:"y1564"`
		MSN     uint64 `json:"msn"`
	} `json:"signatures"`
	Latency struct {
		MinUs   float64 `json:"min_us"`
		AvgUs   float64 `json:"avg_us"`
		MaxUs   float64 `json:"max_us"`
		Count   uint64  `json:"count"`
		Enabled bool    `json:"enabled"`
	} `json:"latency"`
}

// ConfigResponse is the JSON structure for config API
type ConfigResponse struct {
	Interface       string `json:"interface"`
	SignatureFilter string `json:"signature_filter"`
	Filtering       struct {
		Port      uint16 `json:"port"`
		FilterOUI bool   `json:"filter_oui"`
		OUI       string `json:"oui"`
		FilterMAC bool   `json:"filter_mac"`
	} `json:"filtering"`
	Reflection struct {
		Mode string `json:"mode"`
	} `json:"reflection"`
	Platform struct {
		Type string `json:"type"`
	} `json:"platform"`
}

// New creates a new web server
func New(dp *dataplane.Dataplane, port int) *Server {
	s := &Server{
		dp:        dp,
		port:      port,
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}

	// API routes
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/stats/reset", s.handleResetStats)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/health", s.handleHealth)

	// Serve embedded React app
	distFS, err := fs.Sub(reactApp, "dist")
	if err != nil {
		// If no dist folder, serve a simple status page
		s.mux.HandleFunc("/", s.handleFallback)
	} else {
		s.mux.Handle("/", http.FileServer(http.FS(distFS)))
	}

	return s
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return server.ListenAndServe()
}

// handleStats returns current statistics as JSON
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	elapsed := time.Since(s.startTime).Seconds()
	resp := StatsResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    elapsed,
	}

	// Handle nil dataplane (for testing without CGO)
	if s.dp != nil {
		stats := s.dp.GetStats()
		pps := float64(0)
		mbps := float64(0)
		if elapsed > 0 {
			pps = float64(stats.PacketsReflected) / elapsed
			mbps = float64(stats.BytesReflected) * 8.0 / (elapsed * 1000000.0)
		}

		resp.Interface = s.dp.Interface()
		resp.Running = s.dp.IsRunning()
		resp.PacketsReceived = stats.PacketsReceived
		resp.PacketsReflected = stats.PacketsReflected
		resp.BytesReceived = stats.BytesReceived
		resp.BytesReflected = stats.BytesReflected
		resp.TxErrors = stats.TxErrors
		resp.RxInvalid = stats.RxInvalid
		resp.RatePackets = pps
		resp.RateMbps = mbps
		resp.Signatures.ProbeOT = stats.SigProbeOT
		resp.Signatures.DataOT = stats.SigDataOT
		resp.Signatures.Latency = stats.SigLatency
		resp.Signatures.RFC2544 = stats.SigRFC2544
		resp.Signatures.Y1564 = stats.SigY1564
		resp.Signatures.MSN = stats.SigMSN
		resp.Latency.MinUs = stats.LatencyMin
		resp.Latency.AvgUs = stats.LatencyAvg
		resp.Latency.MaxUs = stats.LatencyMax
		resp.Latency.Count = stats.LatencyCount
		resp.Latency.Enabled = stats.LatencyCount > 0
	}

	setCORSHeaders(w, r)
	writeJSON(w, resp)
}

// handleConfig handles GET (read) and POST (update) for configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		setCORSHeaders(w, r)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	setCORSHeaders(w, r)

	// Require dataplane for config operations
	if s.dp == nil {
		writeJSON(w, ConfigResponse{Platform: struct {
			Type string `json:"type"`
		}{Type: "none"}})
		return
	}

	if r.Method == http.MethodPost {
		// Update configuration
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.dp.UpdateConfig(updates); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, StatusResponse{Status: "ok"})
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := s.dp.Config()
	resp := ConfigResponse{
		Interface:       cfg.Interface,
		SignatureFilter: cfg.SignatureFilter,
	}
	resp.Filtering.Port = cfg.Filtering.Port
	resp.Filtering.FilterOUI = cfg.Filtering.FilterOUI
	resp.Filtering.OUI = cfg.Filtering.OUI
	resp.Filtering.FilterMAC = cfg.Filtering.FilterMAC
	resp.Reflection.Mode = cfg.Reflection.Mode

	if cfg.Platform.UseDPDK {
		resp.Platform.Type = "DPDK"
	} else if cfg.Platform.UseAFXDP {
		resp.Platform.Type = "AF_XDP"
	} else {
		resp.Platform.Type = "AF_PACKET"
	}

	writeJSON(w, resp)
}

// handleResetStats resets statistics counters
func (s *Server) handleResetStats(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.dp != nil {
		s.dp.ResetStats()
	}
	s.startTime = time.Now() // Reset uptime too
	writeJSON(w, StatusResponse{Status: "ok"})
}

// handleHealth returns a simple health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
	running := false
	if s.dp != nil {
		running = s.dp.IsRunning()
	}
	writeJSON(w, HealthResponse{
		Status:  "ok",
		Running: running,
		Uptime:  time.Since(s.startTime).Seconds(),
		Version: "2.0.0",
	})
}

// handleFallback serves a simple HTML page when React app isn't built
func (s *Server) handleFallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var iface, status = "N/A", "No dataplane"
	var rxPkts, txPkts, rxBytes, txBytes uint64

	if s.dp != nil {
		stats := s.dp.GetStats()
		iface = s.dp.Interface()
		if s.dp.IsRunning() {
			status = "Running"
		} else {
			status = "Stopped"
		}
		rxPkts = stats.PacketsReceived
		txPkts = stats.PacketsReflected
		rxBytes = stats.BytesReceived
		txBytes = stats.BytesReflected
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Reflector 2.0</title>
    <meta http-equiv="refresh" content="5">
    <style>
        body { font-family: monospace; background: #1a1a2e; color: #eee; padding: 20px; }
        .stat { margin: 5px 0; }
        .label { color: #00d4ff; }
        h1 { color: #00ff88; }
    </style>
</head>
<body>
    <h1>Reflector 2.0</h1>
    <p>Interface: %s | Status: %s</p>
    <div class="stat"><span class="label">RX Packets:</span> %d</div>
    <div class="stat"><span class="label">TX Packets:</span> %d</div>
    <div class="stat"><span class="label">RX Bytes:</span> %d</div>
    <div class="stat"><span class="label">TX Bytes:</span> %d</div>
    <hr>
    <p><small>Auto-refresh every 5 seconds. Build React UI for full dashboard.</small></p>
</body>
</html>`, iface, status, rxPkts, txPkts, rxBytes, txBytes)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
