// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"fmt"

	"github.com/MustardSeedNetworks/stem/internal/logging"
	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/reflector"
)

// Default configuration constants for module tests.
const (
	defaultFrameSize = 1518 // Default Ethernet frame size (bytes).
	defaultDuration  = 60   // Default test duration (seconds).
)

// testExecutor and executorFactory are the API-layer aliases for the canonical
// types in modtypes. The executor factories now live in the module registry
// (services.Factory) instead of a parallel map here, so a module's metadata and
// execution are a single source of truth (registered together in
// services.buildDefaultRegistry). Tests override Server.executorResolver with a
// factory returning a fake modtypes.Executor to drive executeTest without the
// real cgo dataplane.
type (
	testExecutor    = modtypes.Executor
	executorFactory = modtypes.ExecutorFactory
)

// executeTest runs the test via the appropriate module executor.
func (s *Server) executeTest(
	moduleName, testType, iface, profile string,
	config *TestConfig,
) error {
	// Handle reflector separately as it has different lifecycle.
	if moduleName == moduleReflector {
		return s.executeReflector(iface, profile)
	}

	resolver := s.testExecutorResolver()
	if resolver == nil {
		resolver = services.Factory
	}
	factory, ok := resolver(moduleName)
	if !ok {
		return fmt.Errorf("executor not implemented for module: %s", moduleName)
	}

	return s.runModuleTest(factory, moduleName, testType, iface, config)
}

func (s *Server) testExecutorResolver() func(string) (executorFactory, bool) {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	return s.executorResolver
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
	s.statsMu.Lock()
	runID := s.testRunID
	s.activeTestExec = exec
	s.testStatus = statusRunning
	s.statsMu.Unlock()

	// Run test in goroutine.
	go func() {
		defer exec.Close()

		// Convert server config to module config with params map.
		cfg := convertToModuleConfig(iface, testType, config)

		result, execErr := exec.Execute(testType, cfg)

		s.statsMu.Lock()
		if s.testRunID != runID {
			s.statsMu.Unlock()
			return
		}
		s.activeTestExec = nil

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
			logging.Error(
				"Test failed",
				"module",
				moduleName,
				"testType",
				testType,
				"error",
				execErr,
			)
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
		logging.Info(
			"Test completed",
			"module",
			moduleName,
			"testType",
			testType,
			"success",
			result.Success,
		)
	}()

	return nil
}

// executeReflector starts the reflector mode.
func (s *Server) executeReflector(iface, profile string) error {
	s.reflectorMu.Lock()
	defer s.reflectorMu.Unlock()

	if err := s.validateReflectorProfile(profile); err != nil {
		return err
	}
	// Check if already running.
	s.statsMu.Lock()
	if s.reflectorExec != nil && s.reflectorExec.IsRunning() {
		s.statsMu.Unlock()
		return errors.New("reflector already running")
	}

	// Recreate the executor when the selected interface changes.
	if s.reflectorExec != nil && s.reflectorExec.Dataplane().Interface() != iface {
		oldExec := s.reflectorExec
		s.reflectorExec = nil
		s.statsMu.Unlock()
		oldExec.Close()
		s.statsMu.Lock()
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
	exec := s.reflectorExec
	storedConfig := s.reflectorConfig
	s.statsMu.Unlock()

	if profile != "" {
		storedConfig.Profile = profile
	}
	update := reflectorStartConfig(storedConfig)
	if err := exec.Dataplane().UpdateConfig(update); err != nil {
		return fmt.Errorf("apply reflector config: %w", err)
	}

	result, err := exec.Execute(testTypeReflect, nil)
	if err != nil {
		return err
	}

	s.statsMu.Lock()
	s.testStatus = statusRunning
	s.currentTest = testTypeReflect
	s.currentModule = ""
	s.testResult = &TestResultResponse{
		Status:   statusRunning,
		TestType: testTypeReflect,
		Module:   moduleReflector,
		Success:  result.Success,
		Error:    "",
		Message:  "",
		Data:     result.Data,
	}
	s.statsMu.Unlock()
	logging.Info("Reflector started", "success", result.Success)

	return nil
}

func reflectorStartConfig(storedConfig ReflectorConfig) *reflectorDP.ConfigUpdate {
	settings := reflectorConfig.SettingsForProfile(storedConfig.Profile)
	update := &reflectorDP.ConfigUpdate{
		Port:            &settings.Port,
		Mode:            &settings.Mode,
		SignatureFilter: &settings.SignatureFilter,
	}
	if storedConfig.Profile != reflectorConfig.ProfileCustom {
		filterOUI := false
		update.FilterOUI = &filterOUI
		return update
	}
	if port, ok := safeIntToUint16(storedConfig.PortFilter); ok && port > 0 {
		update.Port = &port
	}
	if len(storedConfig.SignatureFilter) == 1 {
		filter := storedConfig.SignatureFilter[0]
		update.SignatureFilter = &filter
	}
	if storedConfig.OUIFilter != "" {
		filterOUI := true
		update.FilterOUI = &filterOUI
		update.OUI = &storedConfig.OUIFilter
	}
	return update
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
	modCfg.Params["frame_sizes"] = c.FrameSizes
	modCfg.Params["resolution"] = c.Resolution
	modCfg.Params["max_loss"] = c.MaxLoss
	modCfg.Params["warmup"] = c.Warmup
	modCfg.Params["trials"] = c.Trials
	modCfg.Params["step_size"] = c.StepSize
	modCfg.Params["bidirectional"] = c.Bidirectional
}

// populateRFC2889Params populates the params map with RFC 2889 config.
func populateRFC2889Params(modCfg *modtypes.TestConfig, c *RFC2889TestConfig) {
	modCfg.FrameSize = c.FrameSize
	modCfg.Duration = int(c.Duration)
	modCfg.Params["frame_size"] = c.FrameSize
	modCfg.Params["duration_sec"] = c.Duration
	modCfg.Params["warmup_sec"] = c.Warmup
	modCfg.Params["address_count"] = c.AddressCount
	modCfg.Params["acceptable_loss_pct"] = c.AcceptableLoss
	modCfg.Params["port_count"] = c.PortCount
	modCfg.Params["pattern"] = c.Pattern
}

// populateRFC6349Params populates the params map with RFC 6349 config.
func populateRFC6349Params(modCfg *modtypes.TestConfig, c *RFC6349TestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.Params["target_rate_mbps"] = c.TargetRateMbps
	modCfg.Params["min_rtt_ms"] = c.MinRTTMs
	modCfg.Params["max_rtt_ms"] = c.MaxRTTMs
	modCfg.Params["rwnd_size"] = c.RWNDSize
	modCfg.Params["duration_sec"] = c.Duration
	modCfg.Params["parallel_streams"] = c.ParallelStreams
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
	modCfg.Params["frame_sizes"] = c.FrameSizes
	modCfg.Params["config_duration_sec"] = c.ConfigStepDuration
	modCfg.Params["perf_duration_sec"] = c.PerfTestDuration
	modCfg.Params["vlan_id"] = c.VlanID
	modCfg.Params["cos"] = c.PCP
	modCfg.Params["color_aware"] = c.ColorAware
	modCfg.Params["flr_threshold_pct"] = c.FLRThreshold
	modCfg.Params["fd_threshold_ms"] = c.FDThreshold
	modCfg.Params["fdv_threshold_ms"] = c.FDVThreshold
}

// populateY1731Params populates the params map with Y.1731 config.
func populateY1731Params(modCfg *modtypes.TestConfig, c *Y1731TestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["mep_id"] = c.MepID
	modCfg.Params["meg_level"] = c.MegLevel
	modCfg.Params["meg_id"] = c.MegID
	modCfg.Params["ccm_interval"] = c.CCMInterval
	modCfg.Params["priority"] = c.Priority
	modCfg.Params["duration"] = c.Duration
	modCfg.Params["interval_ms"] = c.IntervalMs
	modCfg.Params["count"] = c.Count
	modCfg.Params["frame_size"] = c.FrameSize
	modCfg.Params["priority_tagged"] = c.PriorityTagged
}

// populateTSNParams populates the params map with TSN config.
func populateTSNParams(modCfg *modtypes.TestConfig, c *TSNTestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["duration_sec"] = c.Duration
	modCfg.Params["warmup_sec"] = c.Warmup
	modCfg.Params["frame_size"] = c.FrameSize
	modCfg.Params["max_latency_ns"] = c.MaxLatencyNs
	modCfg.Params["max_jitter_ns"] = c.MaxJitterNs
	modCfg.Params["require_ptp_sync"] = c.RequirePTPSync
	modCfg.Params["max_sync_offset_ns"] = c.MaxSyncOffsetNs
	modCfg.Params["ptp_enabled"] = c.PTPEnabled
	modCfg.Params["preemption_enabled"] = c.PreemptionEnabled
	modCfg.Params["num_traffic_classes"] = c.NumTrafficClasses
	modCfg.Params["base_time_ns"] = c.BaseTimeNs
	modCfg.Params["cycle_time_ns"] = c.CycleTimeNs
	modCfg.Params["traffic_class"] = c.TrafficClass
}

// populateTrafficGenParams populates the params map with TrafficGen config.
func populateTrafficGenParams(modCfg *modtypes.TestConfig, c *TrafficGenTestConfig) {
	modCfg.Duration = int(c.Duration)
	modCfg.FrameSize = c.FrameSize
	modCfg.Params["frame_size"] = c.FrameSize
	modCfg.Params["rate_pct"] = c.RatePct
	modCfg.Params["duration_sec"] = c.Duration
	modCfg.Params["warmup_sec"] = c.Warmup
	modCfg.Params["stream_id"] = c.StreamID
	modCfg.Params["burst_mode"] = c.BurstMode
	modCfg.Params["burst_size"] = c.BurstSize
	modCfg.Params["inter_burst_gap_us"] = c.InterBurstGapUs
	modCfg.Params["src_mac"] = c.SrcMac
	modCfg.Params["dst_mac"] = c.DstMac
	modCfg.Params["vlan_id"] = c.VlanID
	modCfg.Params["vlan_priority"] = c.VlanPriority
}
