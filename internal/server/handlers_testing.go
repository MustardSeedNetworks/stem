// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/modules"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
)

// handleTestStart starts a test run via the module system.
func (s *Server) handleTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body.
	var req TestStartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Default to throughput if no test type specified.
	if req.TestType == "" {
		req.TestType = testTypeThroughput
	}

	// Look up the module for this test type.
	mod := modules.GetModuleForTest(req.TestType)
	if mod == nil {
		http.Error(w, fmt.Sprintf("Unknown test type: %s", req.TestType), http.StatusBadRequest)
		return
	}

	// Verify the module can run this test.
	if !mod.CanRun(req.TestType) {
		msg := fmt.Sprintf("Module %s cannot run test type: %s", mod.Name(), req.TestType)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Use provided interface or fall back to selected interface.
	iface := req.Interface
	if iface == "" {
		iface = s.selectedIface
	}
	if iface == "" {
		http.Error(w, "No interface specified", http.StatusBadRequest)
		return
	}

	// Check if test is already running.
	s.statsMu.Lock()
	if s.testStatus == statusRunning {
		s.statsMu.Unlock()
		http.Error(w, "Test already running", http.StatusConflict)
		return
	}
	s.testStatus = statusStarting
	s.currentTest = req.TestType
	s.currentModule = mod.Name()
	s.testResult = nil
	s.statsMu.Unlock()

	s.publishTestState(statusStarting, mod.Name(), req.TestType, nil)

	logging.Info("Starting test via module system",
		"testType", req.TestType,
		"module", mod.Name(),
		"interface", iface,
	)

	// Try to create executor and start test.
	err = s.executeTest(mod.Name(), req.TestType, iface, req.FrameSize, req.Duration)
	if err != nil {
		s.statsMu.Lock()
		s.testStatus = statusError
		s.testResult = &TestResultResponse{
			Status:   statusError,
			TestType: req.TestType,
			Module:   mod.Name(),
			Success:  false,
			Error:    err.Error(),
			Message:  "",
			Data:     nil,
		}
		s.statsMu.Unlock()

		// Check if this is a platform limitation.
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

// handleTestStop stops the current test or reflector.
func (s *Server) handleTestStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.statsMu.Lock()
	testType := s.currentTest
	module := s.currentModule
	exec := s.reflectorExec

	// Check if reflector is running.
	if exec != nil && exec.IsRunning() {
		exec.Stop()
		s.testStatus = statusStopped
		s.currentTest = ""
		s.statsMu.Unlock()
		logging.Info("Reflector stopped via API")
		writeJSON(w, StatusResponse{Status: statusStopped})
		return
	}

	// Check if a test is running.
	if s.testStatus != statusRunning && s.testStatus != statusStarting {
		s.statsMu.Unlock()
		http.Error(w, "No test running", http.StatusBadRequest)
		return
	}

	s.testStatus = statusCancelled
	s.currentTest = ""
	s.currentModule = ""
	s.statsMu.Unlock()

	logging.Info("Test cancelled", "testType", testType)
	s.publishTestState(statusCancelled, module, testType, nil)
	writeJSON(w, StatusResponse{Status: statusStopped})
}

// handleTestResult returns the result of the last completed test.
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

	// No result available, return current status.
	writeJSON(w, TestResultResponse{
		Status:   status,
		TestType: currentTest,
		Module:   "",
		Success:  false,
		Error:    "",
		Message:  "No test result available",
		Data:     nil,
	})
}
