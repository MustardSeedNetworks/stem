//go:build !cgo || !linux


// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This file contains stub implementations for non-CGO or non-Linux builds.
// The actual test execution requires CGO and Linux for packet generation.
package dataplane

import (
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/platform"
)

// TestType represents the type of test to execute.
type TestType int

// Test type constants for RFC 2544 and Y.1564 tests.
const (
	TestThroughput     TestType = iota // RFC 2544 throughput test.
	TestLatency                        // RFC 2544 latency test.
	TestFrameLoss                      // RFC 2544 frame loss test.
	TestBackToBack                     // RFC 2544 back-to-back test.
	TestSystemRecovery                 // RFC 2544 system recovery test.
	TestReset                          // RFC 2544 reset test.
	TestY1564Config                    // Y.1564 configuration test.
	TestY1564Perf                      // Y.1564 performance test.
	TestY1564Full                      // Y.1564 full test (config + perf).
)

// TestState represents the current state of a test execution.
type TestState int

// Test state constants.
const (
	StateIdle      TestState = iota // Test not started.
	StateRunning                    // Test in progress.
	StateCompleted                  // Test completed successfully.
	StateFailed                     // Test failed.
	StateCancelled                  // Test was cancelled.
)

// LatencyStats holds latency measurement statistics in nanoseconds.
type LatencyStats struct {
	Count    uint64
	MinNs    float64
	MaxNs    float64
	AvgNs    float64
	JitterNs float64
	P50Ns    float64
	P95Ns    float64
	P99Ns    float64
}

// ThroughputResult holds RFC 2544 throughput test results.
type ThroughputResult struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	MaxRatePps   float64
	FramesTested uint64
	Iterations   uint32
	Latency      LatencyStats
}

// FrameLossPoint holds a single data point from frame loss rate testing.
type FrameLossPoint struct {
	OfferedRatePct float64
	ActualRateMbps float64
	FramesSent     uint64
	FramesRecv     uint64
	LossPct        float64
}

// LatencyResult holds RFC 2544 latency test results.
type LatencyResult struct {
	FrameSize      uint32
	OfferedRatePct float64
	Latency        LatencyStats
}

// BurstResult holds RFC 2544 back-to-back burst test results.
type BurstResult struct {
	FrameSize     uint32
	MaxBurst      uint64
	BurstDuration float64
	Trials        uint32
}

// RecoveryResult holds RFC 2544 system recovery test results.
type RecoveryResult struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResult holds RFC 2544 reset test results.
type ResetResult struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// Y1564SLA holds Y.1564 Service Level Agreement parameters.
type Y1564SLA struct {
	CIRMbps         float64
	EIRMbps         float64
	CBSBytes        uint32
	EBSBytes        uint32
	FDThresholdMs   float64
	FDVThresholdMs  float64
	FLRThresholdPct float64
}

// Y1564Service holds Y.1564 service configuration.
type Y1564Service struct {
	ServiceID   uint32
	ServiceName string
	SLA         Y1564SLA
	FrameSize   uint32
	CoS         uint8
	Enabled     bool
}

// Y1564StepResult holds results for a single Y.1564 configuration test step.
type Y1564StepResult struct {
	Step             uint32
	OfferedRatePct   float64
	AchievedRateMbps float64
	FramesTx         uint64
	FramesRx         uint64
	FLRPct           float64
	FDAvgMs          float64
	FDMinMs          float64
	FDMaxMs          float64
	FDVMs            float64
	FLRPass          bool
	FDPass           bool
	FDVPass          bool
	StepPass         bool
}

// Y1564ConfigResult holds Y.1564 configuration test results.
type Y1564ConfigResult struct {
	ServiceID   uint32
	Steps       [4]Y1564StepResult
	ServicePass bool
}

// Y1564PerfResult holds Y.1564 performance test results.
type Y1564PerfResult struct {
	ServiceID   uint32
	DurationSec uint32
	FramesTx    uint64
	FramesRx    uint64
	FLRPct      float64
	FDAvgMs     float64
	FDMinMs     float64
	FDMaxMs     float64
	FDVMs       float64
	FLRPass     bool
	FDPass      bool
	FDVPass     bool
	ServicePass bool
}

// RFC 2889 configuration and results

type RFC2889Config struct {
	FrameSize         uint32
	DurationSec       uint32
	WarmupSec         uint32
	AddressCount      uint32
	AcceptableLossPct float64
	PortCount         uint32
	Pattern           uint32
}

type RFC2889ForwardingResult struct {
	FrameSize         uint32
	PortCount         uint32
	Pattern           uint32
	MaxRatePct        float64
	MaxRateFps        float64
	AggregateRateMbps float64
	FramesTx          uint64
	FramesRx          uint64
}

type RFC2889CachingResult struct {
	AddressCount uint32
	FrameSize    uint32
	PortCount    uint32
	FramesTx     uint64
	FramesRx     uint64
	LossPct      float64
	Passed       bool
}

type RFC2889LearningResult struct {
	FrameSize           uint32
	PortCount           uint32
	LearningRateFps     float64
	AddressesLearned    uint32
	LearningTimeMs      float64
	VerificationFrames  uint32
	VerificationLossPct float64
}

type RFC2889BroadcastResult struct {
	FrameSize         uint32
	IngressPorts      uint32
	EgressPorts       uint32
	BroadcastRateFps  float64
	BroadcastRateMbps float64
	FramesTx          uint64
	FramesRx          uint64
	ReplicationFactor float64
}

type RFC2889CongestionResult struct {
	FrameSize            uint32
	OverloadRatePct      float64
	FramesTx             uint64
	FramesRx             uint64
	FramesDropped        uint64
	HeadOfLineBlocking   float64
	BackpressureObserved bool
	PauseFramesRx        uint64
}

// RFC 6349 configuration and results

type RFC6349Config struct {
	TargetRateMbps  float64
	MinRTTMs        float64
	MaxRTTMs        float64
	RWNDSize        uint32
	DurationSec     uint32
	ParallelStreams uint32
	MSS             uint32
	Mode            uint32
}

type RFC6349Result struct {
	AchievedRateMbps    float64
	TheoreticalRateMbps float64
	RTTMinMs            float64
	RTTAvgMs            float64
	RTTMaxMs            float64
	BDPBytes            uint64
	RWNDUsed            uint32
	BytesTransferred    uint64
	Retransmissions     uint64
	TestDurationMs      uint32
	TCPEfficiency       float64
	BufferDelayPct      float64
	TransferTimeRatio   float64
	Passed              bool
}

type TCPPathInfo struct {
	PathMTU          uint32
	MSS              uint32
	RTTMinMs         float64
	RTTAvgMs         float64
	RTTMaxMs         float64
	BDPBytes         uint64
	IdealRWND        uint32
	BottleneckBWMbps float64
}

// Y.1731 configuration and results

type Y1731Config struct {
	MEPID          uint32
	MEGLevel       uint32
	MEGID          string
	CCMInterval    uint32
	Priority       uint8
	DurationSec    uint32
	IntervalMs     uint32
	Count          uint32
	FrameSize      uint32
	PriorityTagged bool
}

type Y1731DelayResult struct {
	FramesSent       uint32
	FramesReceived   uint32
	FramesLost       uint32
	DelayMinUs       float64
	DelayAvgUs       float64
	DelayMaxUs       float64
	DelayVariationUs float64
}

type Y1731LossResult struct {
	FramesTx         uint64
	FramesRx         uint64
	NearEndLoss      uint64
	FarEndLoss       uint64
	NearEndLossRatio float64
	FarEndLossRatio  float64
	AvailabilityPct  float64
}

type Y1731LoopbackResult struct {
	LBMSent     uint64
	LBRReceived uint64
	RTTMinMs    float64
	RTTAvgMs    float64
	RTTMaxMs    float64
}

// MEF configuration and results

type MEFConfig struct {
	ServiceID         string
	CoS               uint32
	CIRMbps           float64
	EIRMbps           float64
	CBSBytes          uint32
	EBSBytes          uint32
	FDThresholdUs     float64
	FDVThresholdUs    float64
	FLRThresholdPct   float64
	AvailabilityPct   float64
	ConfigDurationSec uint32
	PerfDurationMin   uint32
	FrameSizes        []uint32
}

type MEFStepResult struct {
	StepPct          uint32
	OfferedRateKbps  uint32
	AchievedRateKbps uint32
	FramesTx         uint64
	FramesRx         uint64
	FDUs             float64
	FDMinUs          float64
	FDMaxUs          float64
	FDVUs            float64
	FLRPct           float64
	Passed           bool
}

type MEFConfigResult struct {
	ServiceID     string
	Steps         [4]MEFStepResult
	NumSteps      uint32
	OverallPassed bool
}

type MEFPerfResult struct {
	ServiceID       string
	DurationSec     uint32
	FramesTx        uint64
	FramesRx        uint64
	ThroughputKbps  uint32
	FDMinUs         float64
	FDAvgUs         float64
	FDMaxUs         float64
	FDVUs           float64
	FLRPct          float64
	AvailabilityPct float64
	FDPassed        bool
	FDVPassed       bool
	FLRPassed       bool
	AvailPassed     bool
	OverallPassed   bool
}

// TSN configuration and results

type TSNConfig struct {
	DurationSec       uint32
	WarmupSec         uint32
	FrameSize         uint32
	MaxLatencyNs      uint32
	MaxJitterNs       uint32
	RequirePTPSync    bool
	MaxSyncOffsetNs   uint32
	PTPEnabled        bool
	PreemptionEnabled bool
	NumTrafficClasses uint32
	BaseTimeNs        uint64
	CycleTimeNs       uint32
	TrafficClass      uint32
}

type TSNTimingResult struct {
	CyclesTested       uint32
	TimingErrors       uint32
	MaxGateDeviationNs float64
	AvgGateDeviationNs float64
	GateTimingPassed   bool
}

type TSNClassResult struct {
	FramesTx         uint64
	FramesRx         uint64
	FramesInterfered uint64
	IsolationPct     float64
	LatencyAvgNs     float64
	LatencyMaxNs     float64
	Passed           bool
}

type TSNIsolationResult struct {
	NumClasses    uint32
	ClassResults  [8]TSNClassResult
	OverallPassed bool
}

type TSNLatencyResult struct {
	TrafficClass  uint32
	Samples       uint32
	LatencyMinNs  float64
	LatencyAvgNs  float64
	LatencyMaxNs  float64
	Latency99Ns   float64
	Latency999Ns  float64
	JitterNs      float64
	LatencyPassed bool
	JitterPassed  bool
	OverallPassed bool
}

type TSNPTPResult struct {
	Samples        uint32
	OffsetAvgNs    float64
	OffsetMaxNs    float64
	OffsetStddevNs float64
	SyncAchieved   bool
}

type TSNFullResult struct {
	TimingResult    TSNTimingResult
	IsolationResult TSNIsolationResult
	LatencyResults  [8]TSNLatencyResult
	PTPResult       TSNPTPResult
	OverallPassed   bool
}

// Traffic generation configuration and results

type TrafficGenConfig struct {
	FrameSize       uint32
	RatePct         float64
	DurationSec     uint32
	WarmupSec       uint32
	StreamID        uint32
	BurstMode       bool
	BurstSize       uint32
	InterBurstGapUs uint32
	SrcMac          string
	DstMac          string
	VlanID          uint16
	VlanPriority    uint8
}

type TrafficGenResult struct {
	PacketsSent  uint64
	PacketsRecv  uint64
	BytesSent    uint64
	LossPct      float64
	ElapsedSec   float64
	AchievedPPS  float64
	AchievedMbps float64
	Latency      LatencyStats
}

// Config holds dataplane test configuration parameters.
type Config struct {
	Interface      string
	LineRate       uint64
	AutoDetect     bool
	TestType       TestType
	FrameSize      uint32
	IncludeJumbo   bool
	TrialDuration  time.Duration
	WarmupPeriod   time.Duration
	InitialRatePct float64
	ResolutionPct  float64
	MaxIterations  uint32
	AcceptableLoss float64
	HWTimestamp    bool
	MeasureLatency bool
	UsePacing      bool
	BatchSize      uint32
	UseDPDK        bool
	DPDKArgs       string
}

// Context holds the test execution context and state.
type Context struct {
	frameSize uint32
}

// Stats holds real-time test execution statistics.
type Stats struct {
	TxPackets   uint64
	TxBytes     uint64
	RxPackets   uint64
	RxBytes     uint64
	CurrentRate float64
	Progress    float64
	Timestamp   time.Time
}

// ThroughputResultCLI holds throughput test results for CLI output.
type ThroughputResultCLI struct {
	FrameSize   uint32
	MaxRatePct  float64
	MaxRateMbps float64
	MaxRatePPS  float64
	Iterations  uint32
	Latency     LatencyStats
}

// LatencyResultCLI holds latency test results for CLI output.
type LatencyResultCLI struct {
	FrameSize uint32
	LoadPct   float64
	Latency   LatencyStats
}

// FrameLossResultCLI holds frame loss test results for CLI output.
type FrameLossResultCLI struct {
	FrameSize  uint32
	OfferedPct float64
	FramesTx   uint64
	FramesRx   uint64
	LossPct    float64
}

// BackToBackResultCLI holds back-to-back test results for CLI output.
type BackToBackResultCLI struct {
	FrameSize       uint32
	MaxBurstFrames  uint64
	BurstDurationUs uint64
	Trials          uint32
}

// RecoveryResultCLI holds system recovery test results for CLI output.
type RecoveryResultCLI struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResultCLI holds reset test results for CLI output.
type ResetResultCLI struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// newPlatformError creates a detailed error about platform requirements.
func newPlatformError() error {
	// In the stub build, platform is always unsupported.
	// Check provides detailed requirements info.
	err := platform.CheckDataplaneSupport()
	if err != nil {
		return fmt.Errorf("dataplane unavailable: %w", err)
	}
	// Fallback should never happen in stub build.
	return &platform.Error{
		Info:    platform.Detect(),
		Message: "dataplane stub: CGO or Linux required for test execution",
	}
}

// ErrNotSupported is returned when CGO dataplane is not available.
// It provides detailed platform requirements and suggestions.
var ErrNotSupported = newPlatformError()

// NewContext creates a new test context for the given interface (stub).
func NewContext(_ string) (*Context, error) {
	return nil, ErrNotSupported
}

// NewTestContext creates a test context for unit testing purposes.
// Unlike NewContext, this returns a valid (but non-functional) Context
// that can be used to test code paths that require a non-nil context.
// All test execution methods will still return ErrNotSupported.
func NewTestContext() *Context {
	return &Context{}
}

// New creates a new test context with the given configuration (stub).
func New(_ Config) (*Context, error) {
	return nil, ErrNotSupported
}

// Configure applies configuration to the context (stub).
func (c *Context) Configure(_ *Config) error {
	return ErrNotSupported
}

// Run starts test execution (stub).
func (c *Context) Run() error {
	return ErrNotSupported
}

// Cancel stops the running test (stub).
func (c *Context) Cancel() {}

// State returns the current test state (stub).
func (c *Context) State() TestState {
	return StateIdle
}

// Close releases context resources (stub).
func (c *Context) Close() {}

// SetFrameSize sets the frame size for test execution.
func (c *Context) SetFrameSize(frameSize uint32) {
	if c != nil {
		c.frameSize = frameSize
	}
}

// RunThroughputTest executes RFC 2544 throughput test (stub).
func (c *Context) RunThroughputTest() (*ThroughputResultCLI, error) {
	return nil, ErrNotSupported
}

// RunLatencyTest executes RFC 2544 latency test at specified load levels (stub).
func (c *Context) RunLatencyTest(_ []float64) ([]LatencyResultCLI, error) {
	return nil, ErrNotSupported
}

// RunFrameLossTest executes RFC 2544 frame loss rate test (stub).
func (c *Context) RunFrameLossTest(_, _, _ float64) ([]FrameLossResultCLI, error) {
	return nil, ErrNotSupported
}

// RunBackToBackTest executes RFC 2544 back-to-back frames test (stub).
func (c *Context) RunBackToBackTest(_ uint64, _ uint32) (*BackToBackResultCLI, error) {
	return nil, ErrNotSupported
}

// RunSystemRecoveryTest executes RFC 2544 system recovery test (stub).
func (c *Context) RunSystemRecoveryTest(_ float64, _ uint32) (*RecoveryResultCLI, error) {
	return nil, ErrNotSupported
}

// RunResetTest executes RFC 2544 reset test (stub).
func (c *Context) RunResetTest() (*ResetResultCLI, error) {
	return nil, ErrNotSupported
}

// RunY1564ConfigTest executes Y.1564 configuration test (stub).
func (c *Context) RunY1564ConfigTest(_ *Y1564Service) (*Y1564ConfigResult, error) {
	return nil, ErrNotSupported
}

// RunY1564PerfTest executes Y.1564 performance test (stub).
func (c *Context) RunY1564PerfTest(_ *Y1564Service, _ uint32) (*Y1564PerfResult, error) {
	return nil, ErrNotSupported
}

// RunRFC2889ForwardingTest executes RFC 2889 forwarding test (stub).
func (c *Context) RunRFC2889ForwardingTest(_ *RFC2889Config) (*RFC2889ForwardingResult, error) {
	return nil, ErrNotSupported
}

// RunRFC2889CachingTest executes RFC 2889 caching test (stub).
func (c *Context) RunRFC2889CachingTest(_ *RFC2889Config) (*RFC2889CachingResult, error) {
	return nil, ErrNotSupported
}

// RunRFC2889LearningTest executes RFC 2889 learning test (stub).
func (c *Context) RunRFC2889LearningTest(_ *RFC2889Config) (*RFC2889LearningResult, error) {
	return nil, ErrNotSupported
}

// RunRFC2889BroadcastTest executes RFC 2889 broadcast test (stub).
func (c *Context) RunRFC2889BroadcastTest(_ *RFC2889Config) (*RFC2889BroadcastResult, error) {
	return nil, ErrNotSupported
}

// RunRFC2889CongestionTest executes RFC 2889 congestion test (stub).
func (c *Context) RunRFC2889CongestionTest(_ *RFC2889Config) (*RFC2889CongestionResult, error) {
	return nil, ErrNotSupported
}

// RunRFC6349PathTest executes RFC 6349 path test (stub).
func (c *Context) RunRFC6349PathTest(_ *RFC6349Config) (*TCPPathInfo, error) {
	return nil, ErrNotSupported
}

// RunRFC6349ThroughputTest executes RFC 6349 throughput test (stub).
func (c *Context) RunRFC6349ThroughputTest(_ *RFC6349Config) (*RFC6349Result, error) {
	return nil, ErrNotSupported
}

// RunY1731DelayTest executes Y.1731 delay test (stub).
func (c *Context) RunY1731DelayTest(_ *Y1731Config) (*Y1731DelayResult, error) {
	return nil, ErrNotSupported
}

// RunY1731LossTest executes Y.1731 loss test (stub).
func (c *Context) RunY1731LossTest(_ *Y1731Config) (*Y1731LossResult, error) {
	return nil, ErrNotSupported
}

// RunY1731SyntheticLossTest executes Y.1731 synthetic loss test (stub).
func (c *Context) RunY1731SyntheticLossTest(_ *Y1731Config) (*Y1731LossResult, error) {
	return nil, ErrNotSupported
}

// RunY1731LoopbackTest executes Y.1731 loopback test (stub).
func (c *Context) RunY1731LoopbackTest(_ *Y1731Config) (*Y1731LoopbackResult, error) {
	return nil, ErrNotSupported
}

// RunMEFConfigTest executes MEF config test (stub).
func (c *Context) RunMEFConfigTest(_ *MEFConfig) (*MEFConfigResult, error) {
	return nil, ErrNotSupported
}

// RunMEFPerfTest executes MEF performance test (stub).
func (c *Context) RunMEFPerfTest(_ *MEFConfig) (*MEFPerfResult, error) {
	return nil, ErrNotSupported
}

// RunMEFFullTest executes MEF full test (stub).
func (c *Context) RunMEFFullTest(_ *MEFConfig) (*MEFConfigResult, *MEFPerfResult, error) {
	return nil, nil, ErrNotSupported
}

// RunTSNGateTimingTest executes TSN gate timing test (stub).
func (c *Context) RunTSNGateTimingTest(_ *TSNConfig) (*TSNTimingResult, error) {
	return nil, ErrNotSupported
}

// RunTSNIsolationTest executes TSN isolation test (stub).
func (c *Context) RunTSNIsolationTest(_ *TSNConfig) (*TSNIsolationResult, error) {
	return nil, ErrNotSupported
}

// RunTSNLatencyTest executes TSN latency test (stub).
func (c *Context) RunTSNLatencyTest(_ *TSNConfig) (*TSNLatencyResult, error) {
	return nil, ErrNotSupported
}

// RunTSNFullTest executes TSN full test (stub).
func (c *Context) RunTSNFullTest(_ *TSNConfig) (*TSNFullResult, error) {
	return nil, ErrNotSupported
}

// RunCustomStreamTest executes custom traffic generation (stub).
func (c *Context) RunCustomStreamTest(_ *TrafficGenConfig) (*TrafficGenResult, error) {
	return nil, ErrNotSupported
}

// GetLineRate returns the line rate for an interface (stub).
func GetLineRate(_ string) uint64 {
	return 0
}

// CalcPPS calculates packets per second for given line rate and frame size (stub).
func CalcPPS(_ uint64, _ uint32) uint64 {
	return 0
}
