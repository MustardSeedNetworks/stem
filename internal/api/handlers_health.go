// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"time"

	"github.com/krisarmstrong/stem/internal/version"
)

// handleHealth returns server health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, HealthResponse{
		Status:  "healthy",
		Version: version.GetVersion(),
		Commit:  version.GetCommit(),
		Product: "The Stem",
		Company: "Mustard Seed Networks",
		Uptime:  int64(time.Since(s.startTime).Seconds()),
	})
}

// handleStats returns current runtime statistics.
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

// handleHealthLive is the Kubernetes liveness probe endpoint.
// Returns 200 OK if the server process is running.
func (s *Server) handleHealthLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, LivenessResponse{
		Status: "ok",
	})
}

// handleHealthReady is the Kubernetes readiness probe endpoint.
// Returns 200 OK if the server is ready to accept traffic.
// Checks: authentication system, license manager.
func (s *Server) handleHealthReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checks := make(map[string]ReadinessCheck)
	allHealthy := true

	// Check auth manager readiness.
	if s.authManager != nil {
		checks["auth"] = ReadinessCheck{Status: "ok", Error: ""}
	} else {
		checks["auth"] = ReadinessCheck{Status: "error", Error: "auth manager not initialized"}
		allHealthy = false
	}

	// Check license manager readiness (optional - warn but don't fail).
	if s.licenseManager != nil {
		checks["license"] = ReadinessCheck{Status: "ok", Error: ""}
	} else {
		checks["license"] = ReadinessCheck{Status: "ok", Error: "license manager not initialized (optional)"}
	}

	// Check that server has started properly.
	if s.startTime.IsZero() {
		checks["server"] = ReadinessCheck{Status: "error", Error: "server not started"}
		allHealthy = false
	} else {
		checks["server"] = ReadinessCheck{Status: "ok", Error: ""}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	writeJSON(w, ReadinessResponse{
		Status: status,
		Checks: checks,
	})
}
