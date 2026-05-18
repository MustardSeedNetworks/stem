// SPDX-License-Identifier: BUSL-1.1

package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/services/orchestrator/config"
)

// Test constants for repeated strings.
const testIfaceEth0 = "eth0"

// ============================================================================
// DefaultConfig Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test default test type.
	if cfg.TestType != config.TestThroughput {
		t.Errorf("Expected TestType=%s, got %s", config.TestThroughput, cfg.TestType)
	}

	// Test default durations.
	if cfg.TrialDuration != 60*time.Second {
		t.Errorf("Expected TrialDuration=60s, got %v", cfg.TrialDuration)
	}

	if cfg.WarmupPeriod != 2*time.Second {
		t.Errorf("Expected WarmupPeriod=2s, got %v", cfg.WarmupPeriod)
	}
}

func TestDefaultConfigThroughput(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Throughput.InitialRatePct != 100.0 {
		t.Errorf("Expected InitialRatePct=100.0, got %f", cfg.Throughput.InitialRatePct)
	}

	if cfg.Throughput.ResolutionPct != 0.1 {
		t.Errorf("Expected ResolutionPct=0.1, got %f", cfg.Throughput.ResolutionPct)
	}

	if cfg.Throughput.MaxIterations != 20 {
		t.Errorf("Expected MaxIterations=20, got %d", cfg.Throughput.MaxIterations)
	}

	if cfg.Throughput.AcceptableLoss != 0.0 {
		t.Errorf("Expected AcceptableLoss=0.0, got %f", cfg.Throughput.AcceptableLoss)
	}
}

func TestDefaultConfigLatency(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Latency.Samples != 1000 {
		t.Errorf("Expected Samples=1000, got %d", cfg.Latency.Samples)
	}

	if len(cfg.Latency.LoadLevels) != 10 {
		t.Errorf("Expected 10 load levels, got %d", len(cfg.Latency.LoadLevels))
	}
}

func TestDefaultConfigFrameLoss(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.FrameLoss.StartPct != 100.0 {
		t.Errorf("Expected StartPct=100.0, got %f", cfg.FrameLoss.StartPct)
	}

	if cfg.FrameLoss.EndPct != 10.0 {
		t.Errorf("Expected EndPct=10.0, got %f", cfg.FrameLoss.EndPct)
	}

	if cfg.FrameLoss.StepPct != 10.0 {
		t.Errorf("Expected StepPct=10.0, got %f", cfg.FrameLoss.StepPct)
	}
}

func TestDefaultConfigBackToBack(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.BackToBack.InitialBurst != 1000 {
		t.Errorf("Expected InitialBurst=1000, got %d", cfg.BackToBack.InitialBurst)
	}

	if cfg.BackToBack.Trials != 50 {
		t.Errorf("Expected Trials=50, got %d", cfg.BackToBack.Trials)
	}
}

// ============================================================================
// Protocol Default Config Tests
// ============================================================================

func TestDefaultY1564Config(t *testing.T) {
	cfg := config.DefaultY1564Config()

	if len(cfg.ConfigSteps) != 4 {
		t.Errorf("Expected 4 config steps, got %d", len(cfg.ConfigSteps))
	}

	expectedSteps := []float64{25, 50, 75, 100}
	for i, step := range cfg.ConfigSteps {
		if step != expectedSteps[i] {
			t.Errorf("Step %d: expected %f, got %f", i, expectedSteps[i], step)
		}
	}

	if cfg.StepDuration != 60*time.Second {
		t.Errorf("Expected StepDuration=60s, got %v", cfg.StepDuration)
	}

	if cfg.PerfDuration != 15*time.Minute {
		t.Errorf("Expected PerfDuration=15m, got %v", cfg.PerfDuration)
	}
}

func TestDefaultY1564SLA(t *testing.T) {
	sla := config.DefaultY1564SLA()

	if sla.CIRMbps != 100.0 {
		t.Errorf("Expected CIRMbps=100.0, got %f", sla.CIRMbps)
	}

	if sla.FDThresholdMs != 10.0 {
		t.Errorf("Expected FDThresholdMs=10.0, got %f", sla.FDThresholdMs)
	}

	if sla.FDVThresholdMs != 5.0 {
		t.Errorf("Expected FDVThresholdMs=5.0, got %f", sla.FDVThresholdMs)
	}

	if sla.FLRThresholdPct != 0.01 {
		t.Errorf("Expected FLRThresholdPct=0.01, got %f", sla.FLRThresholdPct)
	}
}

func TestDefaultRFC2889Config(t *testing.T) {
	cfg := config.DefaultRFC2889Config()

	if cfg.PortCount != 2 {
		t.Errorf("Expected PortCount=2, got %d", cfg.PortCount)
	}

	if cfg.AddressCount != 8192 {
		t.Errorf("Expected AddressCount=8192, got %d", cfg.AddressCount)
	}
}

func TestDefaultRFC6349Config(t *testing.T) {
	cfg := config.DefaultRFC6349Config()

	if cfg.MSS != 1460 {
		t.Errorf("Expected MSS=1460, got %d", cfg.MSS)
	}

	if cfg.RWND != 65535 {
		t.Errorf("Expected RWND=65535, got %d", cfg.RWND)
	}
}

func TestDefaultY1731Config(t *testing.T) {
	cfg := config.DefaultY1731Config()

	if cfg.MEPID != 1 {
		t.Errorf("Expected MEPID=1, got %d", cfg.MEPID)
	}

	if cfg.MEGLevel != 4 {
		t.Errorf("Expected MEGLevel=4, got %d", cfg.MEGLevel)
	}

	if cfg.MEGID != "DEFAULT-MEG" {
		t.Errorf("Expected MEGID='DEFAULT-MEG', got '%s'", cfg.MEGID)
	}
}

func TestDefaultMEFConfig(t *testing.T) {
	cfg := config.DefaultMEFConfig()

	if cfg.CIRMbps != 100.0 {
		t.Errorf("Expected CIRMbps=100.0, got %f", cfg.CIRMbps)
	}

	if cfg.CBSBytes != 12000 {
		t.Errorf("Expected CBSBytes=12000, got %d", cfg.CBSBytes)
	}
}

func TestDefaultTSNConfig(t *testing.T) {
	cfg := config.DefaultTSNConfig()

	if cfg.NumClasses != 8 {
		t.Errorf("Expected NumClasses=8, got %d", cfg.NumClasses)
	}

	if cfg.CycleTimeNs != 1000000 {
		t.Errorf("Expected CycleTimeNs=1000000, got %d", cfg.CycleTimeNs)
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateNoInterface(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing interface")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidateInvalidTestType(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = "invalid_test"

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid test type")
	}
}

func TestValidateInvalidFrameSize(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.FrameSize = 100 // Not a standard size.

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid frame size")
	}
}

func TestValidateValidFrameSizes(t *testing.T) {
	validSizes := []uint32{0, 64, 128, 256, 512, 1024, 1280, 1518, 9000}

	for _, size := range validSizes {
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.FrameSize = size

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Frame size %d should be valid, got error: %v", size, err)
		}
	}
}

func TestValidateInvalidResolution(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.Throughput.ResolutionPct = 0.0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for zero resolution")
	}

	cfg.Throughput.ResolutionPct = 15.0
	err = cfg.Validate()
	if err == nil {
		t.Error("Expected error for resolution > 10")
	}
}

func TestValidateInvalidFrameLoss(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.FrameLoss.StartPct = 10.0
	cfg.FrameLoss.EndPct = 100.0 // Start < End is invalid.

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for start < end in frame loss")
	}
}

func TestValidateY1564NoServices(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestY1564Full
	cfg.Y1564.Services = []config.Y1564Service{} // No services.

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for Y.1564 without services")
	}
}

func TestValidateY1564ZeroCIR(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestY1564Full
	cfg.Y1564.Services = []config.Y1564Service{
		{
			ServiceID:   1,
			ServiceName: "",
			SLA: config.Y1564SLA{
				CIRMbps:         0,
				EIRMbps:         0,
				CBSBytes:        0,
				EBSBytes:        0,
				FDThresholdMs:   0,
				FDVThresholdMs:  0,
				FLRThresholdPct: 0,
			},
			FrameSize: 0,
			CoS:       0,
			Enabled:   true,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for Y.1564 with zero CIR")
	}
}

func TestValidateRFC2889InsufficientPorts(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestRFC2889Forwarding
	cfg.RFC2889.PortCount = 1 // Need at least 2.

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for RFC 2889 with < 2 ports")
	}
}

func TestValidateRFC6349ZeroMSS(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestRFC6349Throughput
	cfg.RFC6349.MSS = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for RFC 6349 with zero MSS")
	}
}

func TestValidateY1731ZeroMEPID(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestY1731Delay
	cfg.Y1731.MEPID = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for Y.1731 with zero MEP ID")
	}
}

func TestValidateMEFZeroCIR(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestMEFFull
	cfg.MEF.CIRMbps = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for MEF with zero CIR")
	}
}

func TestValidateTSNZeroCycleTime(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestTSNFull
	cfg.TSN.CycleTimeNs = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for TSN with zero cycle time")
	}
}

// ============================================================================
// StandardFrameSizes Tests
// ============================================================================

func TestStandardFrameSizes(t *testing.T) {
	sizes := config.StandardFrameSizes(false)

	expected := []uint32{64, 128, 256, 512, 1024, 1280, 1518}
	if len(sizes) != len(expected) {
		t.Errorf("Expected %d sizes, got %d", len(expected), len(sizes))
	}

	for i, size := range sizes {
		if size != expected[i] {
			t.Errorf("Size %d: expected %d, got %d", i, expected[i], size)
		}
	}
}

func TestStandardFrameSizesWithJumbo(t *testing.T) {
	sizes := config.StandardFrameSizes(true)

	expected := []uint32{64, 128, 256, 512, 1024, 1280, 1518, 9000}
	if len(sizes) != len(expected) {
		t.Errorf("Expected %d sizes, got %d", len(expected), len(sizes))
	}

	// Check last one is jumbo.
	if sizes[len(sizes)-1] != 9000 {
		t.Errorf("Expected last size to be 9000, got %d", sizes[len(sizes)-1])
	}
}

// ============================================================================
// Load/Save Tests
// ============================================================================

func TestSaveAndLoad(t *testing.T) {
	// Use t.TempDir() for automatic cleanup.
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Create config.
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestLatency
	cfg.FrameSize = 1518

	// Save.
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load.
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify.
	if loaded.Interface != cfg.Interface {
		t.Errorf("Interface: expected %s, got %s", cfg.Interface, loaded.Interface)
	}

	if loaded.TestType != cfg.TestType {
		t.Errorf("TestType: expected %s, got %s", cfg.TestType, loaded.TestType)
	}

	if loaded.FrameSize != cfg.FrameSize {
		t.Errorf("FrameSize: expected %d, got %d", cfg.FrameSize, loaded.FrameSize)
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML.
	err := writeTestFile(t, configPath, "{{{{invalid yaml")
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = config.Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")

	// Write config missing interface.
	err := writeTestFile(t, configPath, "test_type: throughput\n")
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = config.Load(configPath)
	if err == nil {
		t.Error("Expected validation error")
	}
}

// writeTestFile is a helper to write test files.
func writeTestFile(t *testing.T, path, content string) error {
	t.Helper()
	const fileMode = 0o600
	return os.WriteFile(path, []byte(content), fileMode)
}

// ============================================================================
// Test Type Tests
// ============================================================================

func TestTestTypeConstants(t *testing.T) {
	// Verify test type string values (exhaustive map of all TestType constants).
	testTypes := map[config.TestType]string{
		config.TestThroughput:        "throughput",
		config.TestLatency:           "latency",
		config.TestFrameLoss:         "frame_loss",
		config.TestBackToBack:        "back_to_back",
		config.TestSystemRecovery:    "system_recovery",
		config.TestReset:             "reset",
		config.TestY1564Config:       "y1564_config",
		config.TestY1564Perf:         "y1564_perf",
		config.TestY1564Full:         "y1564",
		config.TestRFC2889Forwarding: "rfc2889_forwarding",
		config.TestRFC2889Caching:    "rfc2889_caching",
		config.TestRFC2889Learning:   "rfc2889_learning",
		config.TestRFC2889Broadcast:  "rfc2889_broadcast",
		config.TestRFC2889Congestion: "rfc2889_congestion",
		config.TestRFC6349Throughput: "rfc6349_throughput",
		config.TestRFC6349Path:       "rfc6349_path",
		config.TestY1731Delay:        "y1731_delay",
		config.TestY1731Loss:         "y1731_loss",
		config.TestY1731SLM:          "y1731_slm",
		config.TestY1731Loopback:     "y1731_loopback",
		config.TestMEFConfig:         "mef_config",
		config.TestMEFPerf:           "mef_perf",
		config.TestMEFFull:           "mef",
		config.TestTSNTiming:         "tsn_timing",
		config.TestTSNIsolation:      "tsn_isolation",
		config.TestTSNLatency:        "tsn_latency",
		config.TestTSNFull:           "tsn",
	}

	for tt, expected := range testTypes {
		if string(tt) != expected {
			t.Errorf("TestType %v: expected string '%s', got '%s'", tt, expected, string(tt))
		}
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestOutputFormatConstants(t *testing.T) {
	formats := map[config.OutputFormat]string{
		config.FormatText: "text",
		config.FormatJSON: "json",
		config.FormatCSV:  "csv",
	}

	for fmt, expected := range formats {
		if string(fmt) != expected {
			t.Errorf("OutputFormat %v: expected '%s', got '%s'", fmt, expected, string(fmt))
		}
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// ============================================================================
// Additional Validation Tests for Coverage
// ============================================================================

// TestSaveInvalidPath tests Save with an invalid path.
func TestSaveInvalidPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0

	// Try to save to a non-existent directory.
	err := cfg.Save("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

// TestValidateAllRFC2889TestTypes tests all RFC 2889 test type validations.
func TestValidateAllRFC2889TestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestRFC2889Forwarding,
		config.TestRFC2889Caching,
		config.TestRFC2889Learning,
		config.TestRFC2889Broadcast,
		config.TestRFC2889Congestion,
	}

	for _, tt := range testTypes {
		// Valid config.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.RFC2889.PortCount = 2

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (insufficient ports).
		cfg.RFC2889.PortCount = 1
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for insufficient ports", tt)
		}
	}
}

// TestValidateAllRFC6349TestTypes tests all RFC 6349 test type validations.
func TestValidateAllRFC6349TestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestRFC6349Throughput,
		config.TestRFC6349Path,
	}

	for _, tt := range testTypes {
		// Valid config.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.RFC6349.MSS = 1460

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (zero MSS).
		cfg.RFC6349.MSS = 0
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for zero MSS", tt)
		}
	}
}

// TestValidateAllY1731TestTypes tests all Y.1731 test type validations.
func TestValidateAllY1731TestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestY1731Delay,
		config.TestY1731Loss,
		config.TestY1731SLM,
		config.TestY1731Loopback,
	}

	for _, tt := range testTypes {
		// Valid config.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.Y1731.MEPID = 1

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (zero MEPID).
		cfg.Y1731.MEPID = 0
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for zero MEPID", tt)
		}
	}
}

// TestValidateAllMEFTestTypes tests all MEF test type validations.
func TestValidateAllMEFTestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestMEFConfig,
		config.TestMEFPerf,
		config.TestMEFFull,
	}

	for _, tt := range testTypes {
		// Valid config.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.MEF.CIRMbps = 100.0

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (zero CIR).
		cfg.MEF.CIRMbps = 0
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for zero CIR", tt)
		}
	}
}

// TestValidateAllTSNTestTypes tests all TSN test type validations.
func TestValidateAllTSNTestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestTSNTiming,
		config.TestTSNIsolation,
		config.TestTSNLatency,
		config.TestTSNFull,
	}

	for _, tt := range testTypes {
		// Valid config.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.TSN.CycleTimeNs = 1000000

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (zero cycle time).
		cfg.TSN.CycleTimeNs = 0
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for zero cycle time", tt)
		}
	}
}

// TestValidateAllY1564TestTypes tests all Y.1564 test type validations.
func TestValidateAllY1564TestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestY1564Config,
		config.TestY1564Perf,
		config.TestY1564Full,
	}

	for _, tt := range testTypes {
		// Valid config with enabled service.
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt
		cfg.Y1564.Services = []config.Y1564Service{
			{
				ServiceID:   1,
				ServiceName: "test-service",
				SLA: config.Y1564SLA{
					CIRMbps:         100.0,
					EIRMbps:         0,
					CBSBytes:        0,
					EBSBytes:        0,
					FDThresholdMs:   0,
					FDVThresholdMs:  0,
					FLRThresholdPct: 0,
				},
				FrameSize: 1518,
				CoS:       0,
				Enabled:   true,
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error with valid config, got %v", tt, err)
		}

		// Invalid config (no services).
		cfg.Y1564.Services = []config.Y1564Service{}
		err = cfg.Validate()
		if err == nil {
			t.Errorf("TestType %s: expected error for no services", tt)
		}
	}
}

// TestValidateY1564DisabledServiceZeroCIR tests that disabled services with zero CIR pass validation.
func TestValidateY1564DisabledServiceZeroCIR(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0
	cfg.TestType = config.TestY1564Full
	cfg.Y1564.Services = []config.Y1564Service{
		{
			ServiceID:   1,
			ServiceName: "disabled-service",
			SLA: config.Y1564SLA{
				CIRMbps:         0, // Zero CIR but disabled.
				EIRMbps:         0,
				CBSBytes:        0,
				EBSBytes:        0,
				FDThresholdMs:   0,
				FDVThresholdMs:  0,
				FLRThresholdPct: 0,
			},
			FrameSize: 0,
			CoS:       0,
			Enabled:   false,
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected no error for disabled service with zero CIR, got %v", err)
	}
}

// TestValidateRFC2544TestTypes tests RFC 2544 test types pass validation.
func TestValidateRFC2544TestTypes(t *testing.T) {
	testTypes := []config.TestType{
		config.TestThroughput,
		config.TestLatency,
		config.TestFrameLoss,
		config.TestBackToBack,
		config.TestSystemRecovery,
		config.TestReset,
	}

	for _, tt := range testTypes {
		cfg := config.DefaultConfig()
		cfg.Interface = testIfaceEth0
		cfg.TestType = tt

		err := cfg.Validate()
		if err != nil {
			t.Errorf("TestType %s: expected no error, got %v", tt, err)
		}
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkDefaultConfig(b *testing.B) {
	for b.Loop() {
		_ = config.DefaultConfig()
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.Interface = testIfaceEth0

	b.ResetTimer()
	for b.Loop() {
		_ = cfg.Validate()
	}
}
