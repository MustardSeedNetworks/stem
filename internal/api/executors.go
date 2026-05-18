// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"fmt"

	"github.com/krisarmstrong/stem/internal/logging"
	"github.com/krisarmstrong/stem/internal/services/benchmark"
	"github.com/krisarmstrong/stem/internal/services/certify"
	"github.com/krisarmstrong/stem/internal/services/measure"
	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/reflector"
	"github.com/krisarmstrong/stem/internal/services/servicetest"
	"github.com/krisarmstrong/stem/internal/services/trafficgen"
)

// Default configuration constants for module tests.
const (
	defaultFrameSize = 1518 // Default Ethernet frame size (bytes).
	defaultDuration  = 60   // Default test duration (seconds).
)

// testExecutor is an interface for module executors that can run tests.
type testExecutor interface {
	Close()
	Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error)
}

// executorFactory creates a new executor for the given interface.
type executorFactory func(iface string) (testExecutor, error)

// getExecutorFactory returns the executor factory for a given module name.
// This function-based approach satisfies gochecknoglobals while maintaining
// the registry pattern for module executor dispatch.
func getExecutorFactory(moduleName string) (executorFactory, bool) {
	factories := map[string]executorFactory{
		moduleBenchmark:   func(iface string) (testExecutor, error) { return benchmark.NewExecutor(iface) },
		moduleServicetest: func(iface string) (testExecutor, error) { return servicetest.NewExecutor(iface) },
		moduleTrafficgen:  func(iface string) (testExecutor, error) { return trafficgen.NewExecutor(iface) },
		moduleMeasure:     func(iface string) (testExecutor, error) { return measure.NewExecutor(iface) },
		moduleCertify:     func(iface string) (testExecutor, error) { return certify.NewExecutor(iface) },
	}
	factory, ok := factories[moduleName]
	return factory, ok
}

// executeTest runs the test via the appropriate module executor.
func (s *Server) executeTest(moduleName, testType, iface string, config *TestConfig) error {
	// Handle reflector separately as it has different lifecycle.
	if moduleName == moduleReflector {
		return s.executeReflector(iface, config)
	}

	factory, ok := getExecutorFactory(moduleName)
	if !ok {
		return fmt.Errorf("executor not implemented for module: %s", moduleName)
	}

	return s.runModuleTest(factory, moduleName, testType, iface, config)
}

// runModuleTest is the generic test execution function that eliminates duplication.
func (s *Server) runModuleTest(
	factory executorFactory,
	moduleName, testType, iface string,
	config *TestConfig,
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

		// Convert server config to module config with params map.
		cfg := convertToModuleConfig(iface, testType, config)

		result, execErr := exec.Execute(testType, cfg)

		s.statsMu.Lock()

		if execErr != nil {
			s.testStatus = statusError
			errResult := &TestResultResponse{
				Status:   statusError,
				TestType: testType,
				Module:   moduleName,
				Success:  false,
				Error:    execErr.Error(),
				Message:  "",
				Data:     nil,
			}
			s.testResult = errResult
			s.currentTest = ""
			s.currentModule = ""
			s.statsMu.Unlock()
			logging.Error("Test failed", "module", moduleName, "testType", testType, "error", execErr)
			return
		}

		s.testStatus = statusCompleted
		completedResult := &TestResultResponse{
			Status:   statusCompleted,
			TestType: testType,
			Module:   moduleName,
			Success:  result.Success,
			Error:    result.Error,
			Message:  "",
			Data:     result.Data,
		}
		s.testResult = completedResult
		s.currentTest = ""
		s.currentModule = ""
		s.statsMu.Unlock()
		logging.Info("Test completed", "module", moduleName, "testType", testType, "success", result.Success)
	}()

	return nil
}

// executeReflector starts the reflector mode.
func (s *Server) executeReflector(iface string, config *TestConfig) error {
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

		// Note: Reflector uses its own config type (reflector.Config).
		// Currently the reflector's Execute ignores the config parameter.
		// Convert server config to reflector.Config when reflector uses it.
		_ = config // Suppress unused parameter warning.
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
	}()

	return nil
}

// convertToModuleConfig converts server TestConfig to modtypes.TestConfig with params map.
func convertToModuleConfig(iface, testType string, cfg *TestConfig) *modtypes.TestConfig {
	modCfg := &modtypes.TestConfig{
		Interface: iface,
		FrameSize: defaultFrameSize,
		Duration:  defaultDuration,
		Params:    make(map[string]any),
	}

	if cfg == nil {
		return modCfg
	}

	// Route config based on test type prefix.
	switch {
	case isRFC2544Test(testType) && cfg.RFC2544 != nil:
		populateRFC2544Params(modCfg, cfg.RFC2544)
	case isRFC2889Test(testType) && cfg.RFC2889 != nil:
		populateRFC2889Params(modCfg, cfg.RFC2889)
	case isRFC6349Test(testType) && cfg.RFC6349 != nil:
		populateRFC6349Params(modCfg, cfg.RFC6349)
	case isY1564Test(testType) && cfg.Y1564 != nil:
		populateY1564Params(modCfg, cfg.Y1564)
	case isY1731Test(testType) && cfg.Y1731 != nil:
		populateY1731Params(modCfg, cfg.Y1731)
	case isTSNTest(testType) && cfg.TSN != nil:
		populateTSNParams(modCfg, cfg.TSN)
	case isTrafficGenTest(testType) && cfg.TrafficGen != nil:
		populateTrafficGenParams(modCfg, cfg.TrafficGen)
	}

	return modCfg
}

// Test type classification helpers.
func isRFC2544Test(testType string) bool {
	return len(testType) >= 7 && testType[:7] == "rfc2544"
}

func isRFC2889Test(testType string) bool {
	return len(testType) >= 7 && testType[:7] == "rfc2889"
}

func isRFC6349Test(testType string) bool {
	return len(testType) >= 7 && testType[:7] == "rfc6349"
}

func isY1564Test(testType string) bool {
	return len(testType) >= 5 && testType[:5] == "y1564"
}

func isY1731Test(testType string) bool {
	return len(testType) >= 5 && testType[:5] == "y1731"
}

func isTSNTest(testType string) bool {
	return len(testType) >= 3 && testType[:3] == "tsn"
}

func isTrafficGenTest(testType string) bool {
	return testType == "custom_stream" || testType == "trafficgen"
}

// populateRFC2544Params populates the params map with RFC 2544 config.
func populateRFC2544Params(modCfg *modtypes.TestConfig, c *RFC2544TestConfig) {
	modCfg.Duration = c.Duration
	if len(c.FrameSizes) > 0 {
		modCfg.FrameSize = c.FrameSizes[0]
	}
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["frameSizes"] = c.FrameSizes
	modCfg.Params["resolution"] = c.Resolution
	modCfg.Params["maxLoss"] = c.MaxLoss
	modCfg.Params["warmup"] = c.Warmup
	modCfg.Params["trials"] = c.Trials
	modCfg.Params["stepSize"] = c.StepSize
	modCfg.Params["bidirectional"] = c.Bidirectional
}

// populateRFC2889Params populates the params map with RFC 2889 config.
func populateRFC2889Params(modCfg *modtypes.TestConfig, c *RFC2889TestConfig) {
	modCfg.FrameSize = c.FrameSize
	modCfg.Duration = int(c.Duration)
	modCfg.Params["frameSize"] = c.FrameSize
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["warmup"] = c.Warmup
	modCfg.Params["addressCount"] = c.AddressCount
	modCfg.Params["acceptableLoss"] = c.AcceptableLoss
	modCfg.Params["portCount"] = c.PortCount
	modCfg.Params["pattern"] = c.Pattern
}

// populateRFC6349Params populates the params map with RFC 6349 config.
func populateRFC6349Params(modCfg *modtypes.TestConfig, c *RFC6349TestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.Params["targetRateMbps"] = c.TargetRateMbps
	modCfg.Params["minRTTMs"] = c.MinRTTMs
	modCfg.Params["maxRTTMs"] = c.MaxRTTMs
	modCfg.Params["rwndSize"] = c.RWNDSize
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["parallelStreams"] = c.ParallelStreams
	modCfg.Params["mss"] = c.MSS
	modCfg.Params["mode"] = c.Mode
}

// populateY1564Params populates the params map with Y.1564 config.
func populateY1564Params(modCfg *modtypes.TestConfig, c *Y1564TestConfig) {
	modCfg.Duration = int(c.PerfTestDuration)
	if len(c.FrameSizes) > 0 {
		modCfg.FrameSize = c.FrameSizes[0]
	}
	modCfg.Params["cir"] = c.CIR
	modCfg.Params["eir"] = c.EIR
	modCfg.Params["cbs"] = c.CBS
	modCfg.Params["ebs"] = c.EBS
	modCfg.Params["frameSizes"] = c.FrameSizes
	modCfg.Params["configStepDuration"] = c.ConfigStepDuration
	modCfg.Params["perfTestDuration"] = c.PerfTestDuration
	modCfg.Params["vlanId"] = c.VlanID
	modCfg.Params["pcp"] = c.PCP
	modCfg.Params["colorAware"] = c.ColorAware
	modCfg.Params["flrThreshold"] = c.FLRThreshold
	modCfg.Params["fdThreshold"] = c.FDThreshold
	modCfg.Params["fdvThreshold"] = c.FDVThreshold
}

// populateY1731Params populates the params map with Y.1731 config.
func populateY1731Params(modCfg *modtypes.TestConfig, c *Y1731TestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["mepId"] = c.MepID
	modCfg.Params["megLevel"] = c.MegLevel
	modCfg.Params["megId"] = c.MegID
	modCfg.Params["ccmInterval"] = c.CCMInterval
	modCfg.Params["priority"] = c.Priority
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["intervalMs"] = c.IntervalMs
	modCfg.Params["count"] = c.Count
	modCfg.Params["frameSize"] = c.FrameSize
	modCfg.Params["priorityTagged"] = c.PriorityTagged
}

// populateTSNParams populates the params map with TSN config.
func populateTSNParams(modCfg *modtypes.TestConfig, c *TSNTestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["warmup"] = c.Warmup
	modCfg.Params["frameSize"] = c.FrameSize
	modCfg.Params["maxLatencyNs"] = c.MaxLatencyNs
	modCfg.Params["maxJitterNs"] = c.MaxJitterNs
	modCfg.Params["requirePTPSync"] = c.RequirePTPSync
	modCfg.Params["maxSyncOffsetNs"] = c.MaxSyncOffsetNs
	modCfg.Params["ptpEnabled"] = c.PTPEnabled
	modCfg.Params["preemptionEnabled"] = c.PreemptionEnabled
	modCfg.Params["numTrafficClasses"] = c.NumTrafficClasses
	modCfg.Params["baseTimeNs"] = c.BaseTimeNs
	modCfg.Params["cycleTimeNs"] = c.CycleTimeNs
	modCfg.Params["trafficClass"] = c.TrafficClass
}

// populateTrafficGenParams populates the params map with TrafficGen config.
func populateTrafficGenParams(modCfg *modtypes.TestConfig, c *TrafficGenTestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["frameSize"] = c.FrameSize
	modCfg.Params["ratePct"] = c.RatePct
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["warmup"] = c.Warmup
	modCfg.Params["streamId"] = c.StreamID
	modCfg.Params["burstMode"] = c.BurstMode
	modCfg.Params["burstSize"] = c.BurstSize
	modCfg.Params["interBurstGapUs"] = c.InterBurstGapUs
	modCfg.Params["srcMac"] = c.SrcMac
	modCfg.Params["dstMac"] = c.DstMac
	modCfg.Params["vlanId"] = c.VlanID
	modCfg.Params["vlanPriority"] = c.VlanPriority
}
