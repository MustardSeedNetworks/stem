// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/modules"
	"github.com/krisarmstrong/stem/internal/netif"
)

var errTestAlreadyRunning = errors.New("test already running")

// handleTestStart starts a test run via the module system.
func (s *Server) handleTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var req TestStartRequest
	if !decodeJSONStrict(w, r, &req, maxRequestBodySize) {
		return
	}

	if req.TestType == "" {
		req.TestType = testTypeThroughput
	}

	mod, modErr := s.resolveTestModule(req.TestType)
	if modErr != nil {
		WriteInvalidRequest(w, "Unknown or unsupported test type")
		return
	}

	iface, ifaceErr := s.resolveTestInterface(req.Interface)
	if ifaceErr != nil {
		WriteInvalidRequest(w, "No network interface specified")
		return
	}

	// Re-validate interface exists before starting test (#42).
	if !s.validateInterfaceForTest(w, iface) {
		return
	}

	beginErr := s.beginTestRun(req.TestType, mod.Name())
	if beginErr != nil {
		if errors.Is(beginErr, errTestAlreadyRunning) {
			WriteConflict(w, "A test is already running")
			return
		}
		WriteInternalError(w, beginErr)
		return
	}

	logging.Info("Starting test via module system",
		"testType", req.TestType,
		"module", mod.Name(),
		"interface", iface,
	)

	execErr := s.executeTest(mod.Name(), req.TestType, iface, req.Config)
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
		WriteMethodNotAllowed(w)
		return
	}

	s.statsMu.Lock()
	testType := s.currentTest
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
		WriteInvalidRequest(w, "No test is currently running")
		return
	}

	s.testStatus = statusCancelled
	s.currentTest = ""
	s.currentModule = ""
	s.statsMu.Unlock()

	logging.Info("Test cancelled", "testType", testType)
	writeJSON(w, StatusResponse{Status: statusStopped})
}

// handleTestResult returns the result of the last completed test.
func (s *Server) handleTestResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
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

func (s *Server) resolveTestModule(testType string) (modules.Module, error) {
	mod := modules.GetModuleForTest(testType)
	if mod == nil {
		return nil, fmt.Errorf("unknown test type: %s", testType)
	}
	if !mod.CanRun(testType) {
		return nil, fmt.Errorf("module %s cannot run test type: %s", mod.Name(), testType)
	}
	return mod, nil
}

func (s *Server) resolveTestInterface(requested string) (string, error) {
	if requested != "" {
		return requested, nil
	}

	s.statsMu.RLock()
	iface := s.selectedIface
	s.statsMu.RUnlock()

	if iface == "" {
		return "", errors.New("no interface specified")
	}
	return iface, nil
}

// validateInterfaceForTest re-validates the interface exists before starting a test.
func (s *Server) validateInterfaceForTest(w http.ResponseWriter, ifaceName string) bool {
	ifaces, err := netif.DetectInterfaces()
	if err != nil {
		logging.Error("failed to detect interfaces for test validation", "error", err)
		WriteInternalError(w, err)
		return false
	}

	for _, iface := range ifaces {
		if iface.Name == ifaceName {
			return true
		}
	}

	WriteInvalidRequest(w, "Specified network interface no longer exists")
	return false
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
	// Store a sanitized error message for internal state - don't leak details.
	s.testResult = &TestResultResponse{
		Status:   statusError,
		TestType: testType,
		Module:   module,
		Success:  false,
		Error:    "Test execution failed",
		Message:  "",
		Data:     nil,
	}
	s.statsMu.Unlock()

	// Use the centralized error mapping for test errors.
	apiErr := MapTestError(execErr)
	logging.Error("Test execution failed",
		"testType", testType,
		"module", module,
		"error", execErr,
	)
	WriteAPIError(w, apiErr)
}
