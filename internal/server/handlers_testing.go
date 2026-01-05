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

var errTestAlreadyRunning = errors.New("test already running")

// handleTestStart starts a test run via the module system.
func (s *Server) handleTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestStartRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&req)
	if decodeErr != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if req.TestType == "" {
		req.TestType = testTypeThroughput
	}

	mod, modErr := s.resolveTestModule(req.TestType)
	if modErr != nil {
		http.Error(w, modErr.Error(), http.StatusBadRequest)
		return
	}

	iface, ifaceErr := s.resolveTestInterface(req.Interface)
	if ifaceErr != nil {
		http.Error(w, ifaceErr.Error(), http.StatusBadRequest)
		return
	}

	beginErr := s.beginTestRun(req.TestType, mod.Name())
	if beginErr != nil {
		if errors.Is(beginErr, errTestAlreadyRunning) {
			http.Error(w, beginErr.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Unable to start test", http.StatusInternalServerError)
		return
	}

	s.publishTestState(statusStarting, mod.Name(), req.TestType, nil)

	logging.Info("Starting test via module system",
		"testType", req.TestType,
		"module", mod.Name(),
		"interface", iface,
	)

	execErr := s.executeTest(mod.Name(), req.TestType, iface, req.FrameSize, req.Duration)
	if execErr != nil {
		s.respondTestExecutionError(w, execErr, mod.Name(), req.TestType)
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

func (s *Server) resolveTestModule(testType string) (*modules.Module, error) {
	mod := modules.GetModuleForTest(testType)
	if mod == nil {
		return nil, fmt.Errorf("Unknown test type: %s", testType)
	}
	if !mod.CanRun(testType) {
		return nil, fmt.Errorf("Module %s cannot run test type: %s", mod.Name(), testType)
	}
	return mod, nil
}

func (s *Server) resolveTestInterface(requested string) (string, error) {
	if requested != "" {
		return requested, nil
	}
	if s.selectedIface == "" {
		return "", errors.New("No interface specified")
	}
	return s.selectedIface, nil
}

func (s *Server) beginTestRun(testType, module string) error {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if s.testStatus == statusRunning {
		return errTestAlreadyRunning
	}
	s.testStatus = statusStarting
	s.currentTest = testType
	s.currentModule = module
	s.testResult = nil
	return nil
}

func (s *Server) respondTestExecutionError(w http.ResponseWriter, execErr error, module, testType string) {
	s.statsMu.Lock()
	s.testStatus = statusError
	s.testResult = &TestResultResponse{
		Status:   statusError,
		TestType: testType,
		Module:   module,
		Success:  false,
		Error:    execErr.Error(),
		Message:  "",
		Data:     nil,
	}
	s.statsMu.Unlock()

	if errors.Is(execErr, dataplane.ErrNotSupported) {
		logging.Warn("Test execution not supported on this platform",
			"testType", testType,
			"error", execErr,
		)
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, TestStartResponse{
			Status:   "unavailable",
			TestType: testType,
			Module:   module,
			Message:  "Test execution requires Linux with CGO support. This platform cannot execute tests.",
		})
		return
	}

	logging.Error("Failed to start test",
		"testType", testType,
		"error", execErr,
	)
	http.Error(w, fmt.Sprintf("Failed to start test: %v", execErr), http.StatusInternalServerError)
}
