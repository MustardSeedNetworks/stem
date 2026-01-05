// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"errors"
	"fmt"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/modules"
	"github.com/krisarmstrong/stem/internal/modules/benchmark"
	"github.com/krisarmstrong/stem/internal/modules/certify"
	"github.com/krisarmstrong/stem/internal/modules/measure"
	"github.com/krisarmstrong/stem/internal/modules/reflector"
	"github.com/krisarmstrong/stem/internal/modules/servicetest"
	"github.com/krisarmstrong/stem/internal/modules/trafficgen"
)

// testExecutor is an interface for module executors that can run tests.
type testExecutor interface {
	Close()
	Execute(testType string, cfg *modules.TestConfig) (*modules.Result, error)
}

// executorFactory creates a new executor for the given interface.
type executorFactory func(iface string) (testExecutor, error)

// moduleExecutorFactories maps module names to their executor factories.
//
//nolint:gochecknoglobals // Static map of executor factories; read-only after initialization.
var moduleExecutorFactories = map[string]executorFactory{
	moduleBenchmark:   func(iface string) (testExecutor, error) { return benchmark.NewExecutor(iface) },
	moduleServicetest: func(iface string) (testExecutor, error) { return servicetest.NewExecutor(iface) },
	moduleTrafficgen:  func(iface string) (testExecutor, error) { return trafficgen.NewExecutor(iface) },
	moduleMeasure:     func(iface string) (testExecutor, error) { return measure.NewExecutor(iface) },
	moduleCertify:     func(iface string) (testExecutor, error) { return certify.NewExecutor(iface) },
}

// executeTest runs the test via the appropriate module executor.
func (s *Server) executeTest(moduleName, testType, iface string, frameSize uint32, duration int) error {
	// Handle reflector separately as it has different lifecycle.
	if moduleName == moduleReflector {
		return s.executeReflector(iface)
	}

	factory, ok := moduleExecutorFactories[moduleName]
	if !ok {
		return fmt.Errorf("executor not implemented for module: %s", moduleName)
	}

	return s.runModuleTest(factory, moduleName, testType, iface, frameSize, duration)
}

// runModuleTest is the generic test execution function that eliminates duplication.
func (s *Server) runModuleTest(
	factory executorFactory,
	moduleName, testType, iface string,
	frameSize uint32,
	duration int,
) error {
	exec, err := factory(iface)
	if err != nil {
		return fmt.Errorf("create %s executor: %w", moduleName, err)
	}

	// Run test in goroutine.
	go func() {
		defer exec.Close()

		s.statsMu.Lock()
		s.testStatus = statusRunning
		s.statsMu.Unlock()

		s.publishTestState(statusRunning, moduleName, testType, nil)

		cfg := &modules.TestConfig{
			Interface: iface,
			FrameSize: frameSize,
			Duration:  duration,
			Params:    nil,
		}

		result, execErr := exec.Execute(testType, cfg)

		s.statsMu.Lock()

		if execErr != nil {
			s.testStatus = statusError
			s.testResult = &TestResultResponse{
				Status:   statusError,
				TestType: testType,
				Module:   moduleName,
				Success:  false,
				Error:    execErr.Error(),
				Message:  "",
				Data:     nil,
			}
			s.currentTest = ""
			s.currentModule = ""
			s.statsMu.Unlock()
			logging.Error("Test failed", "module", moduleName, "testType", testType, "error", execErr)
			s.publishTestState(statusError, moduleName, testType, s.testResult)
			return
		}

		s.testStatus = statusCompleted
		s.testResult = &TestResultResponse{
			Status:   statusCompleted,
			TestType: testType,
			Module:   moduleName,
			Success:  result.Success,
			Error:    result.Error,
			Message:  "",
			Data:     result.Data,
		}
		s.currentTest = ""
		s.currentModule = ""
		s.statsMu.Unlock()
		logging.Info("Test completed", "module", moduleName, "testType", testType, "success", result.Success)
		s.publishTestState(statusCompleted, moduleName, testType, s.testResult)
	}()

	return nil
}

// executeReflector starts the reflector mode.
func (s *Server) executeReflector(iface string) error {
	// Check if already running.
	s.statsMu.Lock()
	if s.reflectorExec != nil && s.reflectorExec.IsRunning() {
		s.statsMu.Unlock()
		return errors.New("reflector already running")
	}

	// Create new executor if needed.
	if s.reflectorExec == nil {
		exec, err := reflector.NewExecutor(iface)
		if err != nil {
			s.statsMu.Unlock()
			return fmt.Errorf("create reflector executor: %w", err)
		}
		s.reflectorExec = exec
	}
	s.statsMu.Unlock()

	// Run reflector in goroutine.
	go func() {
		s.statsMu.Lock()
		exec := s.reflectorExec
		s.testStatus = statusRunning
		s.currentTest = testTypeReflect
		s.currentModule = moduleReflector
		s.statsMu.Unlock()

		s.publishTestState(statusRunning, moduleReflector, testTypeReflect, nil)

		result, err := exec.Execute(testTypeReflect, nil)

		s.statsMu.Lock()

		if err != nil {
			s.testStatus = statusError
			s.testResult = &TestResultResponse{
				Status:   statusError,
				TestType: testTypeReflect,
				Module:   moduleReflector,
				Success:  false,
				Error:    err.Error(),
				Message:  "",
				Data:     nil,
			}
			logging.Error("Reflector start failed", "error", err)
			s.currentModule = ""
			s.statsMu.Unlock()
			s.publishTestState(statusError, moduleReflector, testTypeReflect, s.testResult)
			return
		}

		s.testResult = &TestResultResponse{
			Status:   statusRunning,
			TestType: testTypeReflect,
			Module:   moduleReflector,
			Success:  result.Success,
			Error:    "",
			Message:  "",
			Data:     result.Data,
		}
		logging.Info("Reflector started", "success", result.Success)
		s.currentModule = ""
		s.statsMu.Unlock()
		s.publishTestState(statusRunning, moduleReflector, testTypeReflect, s.testResult)
	}()

	return nil
}
