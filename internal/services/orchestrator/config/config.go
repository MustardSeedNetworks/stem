// SPDX-License-Identifier: BUSL-1.1

// Package config provides YAML configuration support for the Test Master.
//
// Defines configuration structures for all test types including RFC 2544,
// Y.1564, RFC 2889, RFC 6349, Y.1731, MEF, and TSN tests.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Configuration constants for default values.
const (
	defaultTrialDuration     = 60 * time.Second
	defaultWarmupPeriod      = 2 * time.Second
	defaultInitialRatePct    = 100.0
	defaultResolutionPct     = 0.1
	defaultMaxIterations     = 20
	defaultLatencySamples    = 1000
	defaultFrameLossStartPct = 100.0
	defaultFrameLossEndPct   = 10.0
	defaultFrameLossStepPct  = 10.0
	defaultInitialBurst      = 1000
	defaultTrials            = 50
	defaultBatchSize         = 32
	defaultWebUIAddress      = ":8080"

	// RFC 2889 defaults.
	defaultRFC2889PortCount     = 2
	defaultRFC2889AddressCount  = 8192
	defaultRFC2889TrialDuration = 60 * time.Second

	// RFC 6349 defaults.
	defaultRFC6349MSS          = 1460
	defaultRFC6349RWND         = 65535
	defaultRFC6349TestDuration = 30 * time.Second
	defaultRFC6349Streams      = 1

	// Y.1731 defaults.
	defaultY1731MEPID         = 1
	defaultY1731MEGLevel      = 4
	defaultY1731MEGID         = "DEFAULT-MEG"
	defaultY1731CCMInterval   = 1000
	defaultY1731ProbeCount    = 100
	defaultY1731ProbeInterval = time.Second

	// MEF defaults.
	defaultMEFCIRMbps           = 100.0
	defaultMEFCBSBytes          = 12000
	defaultMEFFDThresholdUs     = 10000 // 10ms.
	defaultMEFFDVThresholdUs    = 5000  // 5ms.
	defaultMEFFLRThresholdPct   = 0.01
	defaultMEFAvailThresholdPct = 99.99
	defaultMEFConfigDuration    = 60 * time.Second
	defaultMEFPerfDuration      = 15 * time.Minute

	// TSN defaults.
	defaultTSNNumClasses      = 8
	defaultTSNCycleTimeNs     = 1000000 // 1ms.
	defaultTSNMaxLatencyNs    = 100000  // 100us.
	defaultTSNMaxJitterNs     = 10000   // 10us.
	defaultTSNMaxSyncOffsetNs = 1000    // 1us.
	defaultTSNTestDuration    = 60 * time.Second
	defaultTSNFrameSize       = 128

	// Y.1564 defaults.
	defaultY1564CIRMbps         = 100.0
	defaultY1564CBSBytes        = 12000
	defaultY1564FDThresholdMs   = 10.0
	defaultY1564FDVThresholdMs  = 5.0
	defaultY1564FLRThresholdPct = 0.01
	defaultY1564StepDuration    = 60 * time.Second
	defaultY1564PerfDuration    = 15 * time.Minute

	// Validation constants.
	maxResolutionPct     = 10
	minRFC2889PortCount  = 2
	numY1564DefaultSteps = 4
	numLatencyLoadLevels = 10

	// Frame size constants.
	jumboFrameSize = 9000
)

// TestType represents the RFC 2544 test types.
type TestType string

// Test type constants for various network testing standards.
const (
	// TestThroughput is RFC 2544 Section 26.1 throughput test.
	TestThroughput TestType = "throughput"
	// TestLatency is RFC 2544 Section 26.2 latency test.
	TestLatency TestType = "latency"
	// TestFrameLoss is RFC 2544 Section 26.3 frame loss test.
	TestFrameLoss TestType = "frame_loss"
	// TestBackToBack is RFC 2544 Section 26.4 back-to-back test.
	TestBackToBack TestType = "back_to_back"
	// TestSystemRecovery is RFC 2544 Section 26.5 system recovery test.
	TestSystemRecovery TestType = "system_recovery"
	// TestReset is RFC 2544 Section 26.6 reset test.
	TestReset TestType = "reset"

	// TestY1564Config is ITU-T Y.1564 service configuration test.
	TestY1564Config TestType = "y1564_config"
	// TestY1564Perf is ITU-T Y.1564 service performance test.
	TestY1564Perf TestType = "y1564_perf"
	// TestY1564Full is ITU-T Y.1564 full test (config + perf).
	TestY1564Full TestType = "y1564"

	// TestRFC2889Forwarding is RFC 2889 forwarding rate test.
	TestRFC2889Forwarding TestType = "rfc2889_forwarding"
	// TestRFC2889Caching is RFC 2889 address caching test.
	TestRFC2889Caching TestType = "rfc2889_caching"
	// TestRFC2889Learning is RFC 2889 address learning test.
	TestRFC2889Learning TestType = "rfc2889_learning"
	// TestRFC2889Broadcast is RFC 2889 broadcast forwarding test.
	TestRFC2889Broadcast TestType = "rfc2889_broadcast"
	// TestRFC2889Congestion is RFC 2889 congestion control test.
	TestRFC2889Congestion TestType = "rfc2889_congestion"

	// TestRFC6349Throughput is RFC 6349 TCP throughput test.
	TestRFC6349Throughput TestType = "rfc6349_throughput"
	// TestRFC6349Path is RFC 6349 path analysis test.
	TestRFC6349Path TestType = "rfc6349_path"

	// TestY1731Delay is ITU-T Y.1731 delay measurement test.
	TestY1731Delay TestType = "y1731_delay"
	// TestY1731Loss is ITU-T Y.1731 loss measurement test.
	TestY1731Loss TestType = "y1731_loss"
	// TestY1731SLM is ITU-T Y.1731 synthetic loss measurement test.
	TestY1731SLM TestType = "y1731_slm"
	// TestY1731Loopback is ITU-T Y.1731 loopback test.
	TestY1731Loopback TestType = "y1731_loopback"

	// TestMEFConfig is MEF service activation configuration test.
	TestMEFConfig TestType = "mef_config"
	// TestMEFPerf is MEF service activation performance test.
	TestMEFPerf TestType = "mef_perf"
	// TestMEFFull is full MEF service activation test.
	TestMEFFull TestType = "mef"

	// TestTSNTiming is IEEE 802.1Qbv gate timing accuracy test.
	TestTSNTiming TestType = "tsn_timing"
	// TestTSNIsolation is IEEE 802.1Qbv traffic class isolation test.
	TestTSNIsolation TestType = "tsn_isolation"
	// TestTSNLatency is IEEE 802.1Qbv scheduled latency test.
	TestTSNLatency TestType = "tsn_latency"
	// TestTSNFull is full IEEE 802.1Qbv TSN test suite.
	TestTSNFull TestType = "tsn"
)

// OutputFormat for results.
type OutputFormat string

// Output format constants.
const (
	// FormatText outputs results as human-readable text.
	FormatText OutputFormat = "text"
	// FormatJSON outputs results as JSON.
	FormatJSON OutputFormat = "json"
	// FormatCSV outputs results as CSV.
	FormatCSV OutputFormat = "csv"
)

// Config represents the full configuration.
type Config struct {
	// Interface settings.
	Interface    string `yaml:"interface"`
	LineRateMbps uint64 `yaml:"line_rate_mbps"` // 0 = auto-detect
	AutoDetect   bool   `yaml:"auto_detect_nic"`

	// Test selection.
	TestType     TestType `yaml:"test_type"`
	FrameSize    uint32   `yaml:"frame_size"`    // 0 = all standard sizes
	IncludeJumbo bool     `yaml:"include_jumbo"` // Include 9000 byte frames

	// Timing.
	TrialDuration time.Duration `yaml:"trial_duration"` // Default: 60s.
	WarmupPeriod  time.Duration `yaml:"warmup_period"`  // Default: 2s.

	// Throughput test (Section 26.1).
	Throughput ThroughputConfig `yaml:"throughput"`

	// Latency test (Section 26.2).
	Latency LatencyConfig `yaml:"latency"`

	// Frame loss test (Section 26.3).
	FrameLoss FrameLossConfig `yaml:"frame_loss"`

	// Back-to-back test (Section 26.4).
	BackToBack BackToBackConfig `yaml:"back_to_back"`

	// Features.
	HWTimestamp    bool `yaml:"hw_timestamp"`
	MeasureLatency bool `yaml:"measure_latency"`

	// Output.
	OutputFormat OutputFormat `yaml:"output_format"`
	Verbose      bool         `yaml:"verbose"`

	// Platform.
	UseDPDK  bool   `yaml:"use_dpdk"`
	DPDKArgs string `yaml:"dpdk_args"`

	// Rate control.
	UsePacing bool   `yaml:"use_pacing"`
	BatchSize uint32 `yaml:"batch_size"`

	// Web UI.
	WebUI WebUIConfig `yaml:"web_ui"`

	// ITU-T Y.1564 (EtherSAM) configuration.
	Y1564 Y1564Config `yaml:"y1564"`

	// Extended protocol test configurations.
	RFC2889 RFC2889Config `yaml:"rfc2889"` // RFC 2889 LAN Switch tests.
	RFC6349 RFC6349Config `yaml:"rfc6349"` // RFC 6349 TCP tests.
	Y1731   Y1731Config   `yaml:"y1731"`   // Y.1731 OAM tests.
	MEF     MEFConfig     `yaml:"mef"`     // MEF Service Activation tests.
	TSN     TSNConfig     `yaml:"tsn"`     // TSN tests.
}

// ThroughputConfig for binary search throughput test.
type ThroughputConfig struct {
	InitialRatePct float64 `yaml:"initial_rate_pct"` // Default: 100.
	ResolutionPct  float64 `yaml:"resolution_pct"`   // Default: 0.1.
	MaxIterations  uint32  `yaml:"max_iterations"`   // Default: 20.
	AcceptableLoss float64 `yaml:"acceptable_loss"`  // Default: 0.0.
}

// LatencyConfig for latency test.
type LatencyConfig struct {
	Samples    uint32    `yaml:"samples"`     // Number of samples per trial.
	LoadLevels []float64 `yaml:"load_levels"` // Load levels to test (% of throughput).
}

// FrameLossConfig for frame loss test.
type FrameLossConfig struct {
	StartPct float64 `yaml:"start_pct"` // Starting offered load %.
	EndPct   float64 `yaml:"end_pct"`   // Ending offered load %.
	StepPct  float64 `yaml:"step_pct"`  // Step size.
}

// BackToBackConfig for burst capacity test.
type BackToBackConfig struct {
	InitialBurst uint64 `yaml:"initial_burst"` // Starting burst size.
	Trials       uint32 `yaml:"trials"`        // Trials per burst size.
}

// WebUIConfig for web interface.
type WebUIConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"` // e.g., ":8080".
}

// Y1564SLA defines SLA parameters for Y.1564 testing.
type Y1564SLA struct {
	CIRMbps         float64 `yaml:"cir_mbps"`          // Committed Information Rate.
	EIRMbps         float64 `yaml:"eir_mbps"`          // Excess Information Rate.
	CBSBytes        uint32  `yaml:"cbs_bytes"`         // Committed Burst Size.
	EBSBytes        uint32  `yaml:"ebs_bytes"`         // Excess Burst Size.
	FDThresholdMs   float64 `yaml:"fd_threshold_ms"`   // Frame Delay threshold (ms).
	FDVThresholdMs  float64 `yaml:"fdv_threshold_ms"`  // Frame Delay Variation threshold (ms).
	FLRThresholdPct float64 `yaml:"flr_threshold_pct"` // Frame Loss Ratio threshold (%).
}

// Y1564Service defines a service for Y.1564 testing.
type Y1564Service struct {
	ServiceID   uint32   `yaml:"service_id"`
	ServiceName string   `yaml:"service_name"`
	SLA         Y1564SLA `yaml:"sla"`
	FrameSize   uint32   `yaml:"frame_size"`
	CoS         uint8    `yaml:"cos"` // Class of Service (DSCP value).
	Enabled     bool     `yaml:"enabled"`
}

// Y1564Config for ITU-T Y.1564 testing.
type Y1564Config struct {
	Services      []Y1564Service `yaml:"services"`
	ConfigSteps   []float64      `yaml:"config_steps"`    // Step percentages (default: 25, 50, 75, 100).
	StepDuration  time.Duration  `yaml:"step_duration"`   // Duration per step (default: 60s).
	PerfDuration  time.Duration  `yaml:"perf_duration"`   // Performance test duration (default: 15m).
	RunConfigTest bool           `yaml:"run_config_test"` // Run configuration test.
	RunPerfTest   bool           `yaml:"run_perf_test"`   // Run performance test.
}

// RFC2889Config for LAN switch benchmarking tests.
type RFC2889Config struct {
	PortCount         uint32        `yaml:"port_count"`          // Number of ports.
	AddressCount      uint32        `yaml:"address_count"`       // MAC addresses for caching test.
	TrialDuration     time.Duration `yaml:"trial_duration"`      // Duration per trial.
	AcceptableLossPct float64       `yaml:"acceptable_loss_pct"` // Acceptable loss percentage.
}

// RFC6349Config for TCP throughput testing.
type RFC6349Config struct {
	TargetRateMbps  float64       `yaml:"target_rate_mbps"` // Target rate (0 = auto).
	MSS             uint32        `yaml:"mss"`              // Maximum Segment Size.
	RWND            uint32        `yaml:"rwnd"`             // Receive Window Size.
	TestDuration    time.Duration `yaml:"test_duration"`    // Test duration.
	ParallelStreams uint32        `yaml:"parallel_streams"` // Number of parallel streams.
}

// Y1731Config for Ethernet OAM testing.
type Y1731Config struct {
	MEPID         uint32        `yaml:"mep_id"`         // MEP identifier.
	MEGLevel      uint8         `yaml:"meg_level"`      // MEG level (0-7).
	MEGID         string        `yaml:"meg_id"`         // MEG identifier.
	CCMInterval   uint32        `yaml:"ccm_interval"`   // CCM interval (ms).
	ProbeCount    uint32        `yaml:"probe_count"`    // Number of probes.
	ProbeInterval time.Duration `yaml:"probe_interval"` // Interval between probes.
}

// MEFConfig for service activation testing.
type MEFConfig struct {
	CIRMbps           float64       `yaml:"cir_mbps"`            // Committed Information Rate.
	EIRMbps           float64       `yaml:"eir_mbps"`            // Excess Information Rate.
	CBSBytes          uint32        `yaml:"cbs_bytes"`           // Committed Burst Size.
	EBSBytes          uint32        `yaml:"ebs_bytes"`           // Excess Burst Size.
	FDThresholdUs     float64       `yaml:"fd_threshold_us"`     // Frame Delay threshold (us).
	FDVThresholdUs    float64       `yaml:"fdv_threshold_us"`    // Frame Delay Variation (us).
	FLRThresholdPct   float64       `yaml:"flr_threshold_pct"`   // Frame Loss Ratio threshold.
	AvailThresholdPct float64       `yaml:"avail_threshold_pct"` // Availability threshold.
	ConfigDuration    time.Duration `yaml:"config_duration"`     // Config test duration.
	PerfDuration      time.Duration `yaml:"perf_duration"`       // Perf test duration.
}

// TSNConfig for Time-Sensitive Networking testing.
type TSNConfig struct {
	NumClasses      uint32        `yaml:"num_classes"`        // Number of traffic classes.
	CycleTimeNs     uint64        `yaml:"cycle_time_ns"`      // GCL cycle time.
	MaxLatencyNs    uint64        `yaml:"max_latency_ns"`     // Maximum latency threshold.
	MaxJitterNs     uint64        `yaml:"max_jitter_ns"`      // Maximum jitter threshold.
	MaxSyncOffsetNs uint64        `yaml:"max_sync_offset_ns"` // Maximum PTP sync offset.
	TestDuration    time.Duration `yaml:"test_duration"`      // Test duration.
	FrameSize       uint32        `yaml:"frame_size"`         // Frame size for testing.
}

// DefaultRFC2889Config returns default RFC 2889 configuration.
func DefaultRFC2889Config() RFC2889Config {
	return RFC2889Config{
		PortCount:         defaultRFC2889PortCount,
		AddressCount:      defaultRFC2889AddressCount,
		TrialDuration:     defaultRFC2889TrialDuration,
		AcceptableLossPct: 0.0,
	}
}

// DefaultRFC6349Config returns default RFC 6349 configuration.
func DefaultRFC6349Config() RFC6349Config {
	return RFC6349Config{
		TargetRateMbps:  0, // Auto-detect.
		MSS:             defaultRFC6349MSS,
		RWND:            defaultRFC6349RWND,
		TestDuration:    defaultRFC6349TestDuration,
		ParallelStreams: defaultRFC6349Streams,
	}
}

// DefaultY1731Config returns default Y.1731 configuration.
func DefaultY1731Config() Y1731Config {
	return Y1731Config{
		MEPID:         defaultY1731MEPID,
		MEGLevel:      defaultY1731MEGLevel,
		MEGID:         defaultY1731MEGID,
		CCMInterval:   defaultY1731CCMInterval,
		ProbeCount:    defaultY1731ProbeCount,
		ProbeInterval: defaultY1731ProbeInterval,
	}
}

// DefaultMEFConfig returns default MEF configuration.
func DefaultMEFConfig() MEFConfig {
	return MEFConfig{
		CIRMbps:           defaultMEFCIRMbps,
		EIRMbps:           0,
		CBSBytes:          defaultMEFCBSBytes,
		EBSBytes:          0,
		FDThresholdUs:     defaultMEFFDThresholdUs,
		FDVThresholdUs:    defaultMEFFDVThresholdUs,
		FLRThresholdPct:   defaultMEFFLRThresholdPct,
		AvailThresholdPct: defaultMEFAvailThresholdPct,
		ConfigDuration:    defaultMEFConfigDuration,
		PerfDuration:      defaultMEFPerfDuration,
	}
}

// DefaultTSNConfig returns default TSN configuration.
func DefaultTSNConfig() TSNConfig {
	return TSNConfig{
		NumClasses:      defaultTSNNumClasses,
		CycleTimeNs:     defaultTSNCycleTimeNs,
		MaxLatencyNs:    defaultTSNMaxLatencyNs,
		MaxJitterNs:     defaultTSNMaxJitterNs,
		MaxSyncOffsetNs: defaultTSNMaxSyncOffsetNs,
		TestDuration:    defaultTSNTestDuration,
		FrameSize:       defaultTSNFrameSize,
	}
}

// DefaultY1564SLA returns default SLA parameters.
func DefaultY1564SLA() Y1564SLA {
	return Y1564SLA{
		CIRMbps:         defaultY1564CIRMbps,
		EIRMbps:         0.0,
		CBSBytes:        defaultY1564CBSBytes,
		EBSBytes:        0,
		FDThresholdMs:   defaultY1564FDThresholdMs,
		FDVThresholdMs:  defaultY1564FDVThresholdMs,
		FLRThresholdPct: defaultY1564FLRThresholdPct,
	}
}

// DefaultY1564Config returns default Y.1564 configuration.
func DefaultY1564Config() Y1564Config {
	return Y1564Config{
		Services:      []Y1564Service{},
		ConfigSteps:   []float64{25, 50, 75, 100},
		StepDuration:  defaultY1564StepDuration,
		PerfDuration:  defaultY1564PerfDuration,
		RunConfigTest: true,
		RunPerfTest:   true,
	}
}

// DefaultConfig returns a configuration with RFC 2544 recommended defaults.
func DefaultConfig() *Config {
	return &Config{
		Interface:     "",
		LineRateMbps:  0,
		AutoDetect:    true,
		TestType:      TestThroughput,
		FrameSize:     0, // All standard sizes.
		IncludeJumbo:  false,
		TrialDuration: defaultTrialDuration,
		WarmupPeriod:  defaultWarmupPeriod,

		Throughput: ThroughputConfig{
			InitialRatePct: defaultInitialRatePct,
			ResolutionPct:  defaultResolutionPct,
			MaxIterations:  defaultMaxIterations,
			AcceptableLoss: 0.0,
		},

		Latency: LatencyConfig{
			Samples:    defaultLatencySamples,
			LoadLevels: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},

		FrameLoss: FrameLossConfig{
			StartPct: defaultFrameLossStartPct,
			EndPct:   defaultFrameLossEndPct,
			StepPct:  defaultFrameLossStepPct,
		},

		BackToBack: BackToBackConfig{
			InitialBurst: defaultInitialBurst,
			Trials:       defaultTrials,
		},

		HWTimestamp:    true,
		MeasureLatency: true,
		OutputFormat:   FormatText,
		Verbose:        false,
		UseDPDK:        false,
		DPDKArgs:       "",
		UsePacing:      true,
		BatchSize:      defaultBatchSize,

		WebUI: WebUIConfig{
			Enabled: false,
			Address: defaultWebUIAddress,
		},

		Y1564: DefaultY1564Config(),

		// Extended protocol test defaults.
		RFC2889: DefaultRFC2889Config(),
		RFC6349: DefaultRFC6349Config(),
		Y1731:   DefaultY1731Config(),
		MEF:     DefaultMEFConfig(),
		TSN:     DefaultTSNConfig(),
	}
}

// Validation error messages.
var (
	errInterfaceRequired      = errors.New("interface is required")
	errRFC2889PortCount       = errors.New("RFC 2889 tests require at least 2 ports")
	errRFC6349MSSRequired     = errors.New("RFC 6349 tests require MSS > 0")
	errY1731MEPIDRequired     = errors.New("Y.1731 tests require MEP ID > 0")
	errMEFCIRRequired         = errors.New("MEF tests require CIR > 0")
	errTSNCycleTimeRequired   = errors.New("TSN tests require cycle_time_ns > 0")
	errY1564NoServices        = errors.New("Y.1564 test requires at least one service configured")
	errResolutionOutOfRange   = errors.New("resolution must be between 0 and 10%")
	errFrameLossStartLessThan = errors.New("frame loss start must be >= end")
)

// Load reads configuration from a YAML file.
func Load(path string) (*Config, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// Save writes configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// validateTestTypeConfig validates test-type specific configuration.
func (c *Config) validateTestTypeConfig() error {
	switch c.TestType {
	case TestThroughput, TestLatency, TestFrameLoss, TestBackToBack,
		TestSystemRecovery, TestReset:
		return nil // Valid RFC 2544 test types.

	case TestY1564Config, TestY1564Perf, TestY1564Full:
		return c.validateY1564Config()

	case TestRFC2889Forwarding, TestRFC2889Caching, TestRFC2889Learning,
		TestRFC2889Broadcast, TestRFC2889Congestion:
		if c.RFC2889.PortCount < minRFC2889PortCount {
			return errRFC2889PortCount
		}
		return nil

	case TestRFC6349Throughput, TestRFC6349Path:
		if c.RFC6349.MSS == 0 {
			return errRFC6349MSSRequired
		}
		return nil

	case TestY1731Delay, TestY1731Loss, TestY1731SLM, TestY1731Loopback:
		if c.Y1731.MEPID == 0 {
			return errY1731MEPIDRequired
		}
		return nil

	case TestMEFConfig, TestMEFPerf, TestMEFFull:
		if c.MEF.CIRMbps <= 0 {
			return errMEFCIRRequired
		}
		return nil

	case TestTSNTiming, TestTSNIsolation, TestTSNLatency, TestTSNFull:
		if c.TSN.CycleTimeNs == 0 {
			return errTSNCycleTimeRequired
		}
		return nil

	default:
		return fmt.Errorf("invalid test type: %s", c.TestType)
	}
}

// validateY1564Config validates Y.1564 specific configuration.
func (c *Config) validateY1564Config() error {
	if len(c.Y1564.Services) == 0 {
		return errY1564NoServices
	}
	for i, svc := range c.Y1564.Services {
		if svc.Enabled && svc.SLA.CIRMbps <= 0 {
			return fmt.Errorf("service %d: CIR must be > 0", i+1)
		}
	}
	return nil
}

// Validate checks configuration for errors.
func (c *Config) Validate() error {
	if c.Interface == "" {
		return errInterfaceRequired
	}

	err := c.validateTestTypeConfig()
	if err != nil {
		return err
	}

	// Validate frame size.
	validSizes := map[uint32]bool{
		0: true, 64: true, 128: true, 256: true, 512: true,
		1024: true, 1280: true, 1518: true, 9000: true,
	}
	if !validSizes[c.FrameSize] {
		return fmt.Errorf("invalid frame size: %d", c.FrameSize)
	}

	// Validate throughput config.
	if c.Throughput.ResolutionPct <= 0 || c.Throughput.ResolutionPct > maxResolutionPct {
		return errResolutionOutOfRange
	}

	// Validate frame loss config.
	if c.FrameLoss.StartPct < c.FrameLoss.EndPct {
		return errFrameLossStartLessThan
	}

	return nil
}

// StandardFrameSizes returns the RFC 2544 standard frame sizes.
func StandardFrameSizes(includeJumbo bool) []uint32 {
	sizes := []uint32{64, 128, 256, 512, 1024, 1280, 1518}
	if includeJumbo {
		sizes = append(sizes, jumboFrameSize)
	}
	return sizes
}
