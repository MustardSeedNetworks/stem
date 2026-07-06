//go:build cgo && linux

package dataplane

import "time"

// TestType mirrors C test_type_t
type TestType int

const (
	TestThroughput TestType = iota
	TestLatency
	TestFrameLoss
	TestBackToBack
	TestSystemRecovery
	TestReset
	TestY1564Config
	TestY1564Perf
	TestY1564Full
)

// TestState mirrors C test_state_t
type TestState int

const (
	StateIdle TestState = iota
	StateRunning
	StateCompleted
	StateFailed
	StateCancelled
)

// LatencyStats contains latency measurements
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

// ThroughputResult from binary search test
type ThroughputResult struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	MaxRatePps   float64
	FramesTested uint64
	Iterations   uint32
	Latency      LatencyStats
}

// FrameLossPoint for a single load level
type FrameLossPoint struct {
	OfferedRatePct float64
	ActualRateMbps float64
	FramesSent     uint64
	FramesRecv     uint64
	LossPct        float64
}

// LatencyResult from latency test
type LatencyResult struct {
	FrameSize      uint32
	OfferedRatePct float64
	Latency        LatencyStats
}

// BurstResult from back-to-back test
type BurstResult struct {
	FrameSize     uint32
	MaxBurst      uint64
	BurstDuration float64
	Trials        uint32
}

// RecoveryResult from RFC 2544 Section 26.5 System Recovery test
type RecoveryResult struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResult from RFC 2544 Section 26.6 Reset test
type ResetResult struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// Y1564SLA contains SLA parameters for Y.1564 testing
type Y1564SLA struct {
	CIRMbps         float64
	EIRMbps         float64
	CBSBytes        uint32
	EBSBytes        uint32
	FDThresholdMs   float64
	FDVThresholdMs  float64
	FLRThresholdPct float64
}

// Y1564Service represents a service configuration for Y.1564 testing
type Y1564Service struct {
	ServiceID   uint32
	ServiceName string
	SLA         Y1564SLA
	FrameSize   uint32
	CoS         uint8
	Enabled     bool
}

// Y1564StepResult from a Y.1564 configuration test step
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

// Y1564ConfigResult from Y.1564 service configuration test
type Y1564ConfigResult struct {
	ServiceID   uint32
	Steps       [4]Y1564StepResult
	ServicePass bool
}

// Y1564PerfResult from Y.1564 service performance test
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

// Traffic generation configuration

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

// Traffic generation result

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

// Config for RFC2544 tests
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

// Stats for real-time monitoring
type Stats struct {
	TxPackets   uint64
	TxBytes     uint64
	RxPackets   uint64
	RxBytes     uint64
	CurrentRate float64
	Progress    float64
	Timestamp   time.Time
}

// =============================================================================
// Wrapper types and functions for CLI integration
// =============================================================================

// ThroughputResult wraps the throughput test result for CLI
type ThroughputResultCLI struct {
	FrameSize   uint32
	MaxRatePct  float64
	MaxRateMbps float64
	MaxRatePPS  float64
	Iterations  uint32
	Latency     LatencyStats
}

// LatencyResultCLI wraps the latency test result for CLI
type LatencyResultCLI struct {
	FrameSize uint32
	LoadPct   float64
	Latency   LatencyStats
}

// FrameLossResultCLI wraps the frame loss test result for CLI
type FrameLossResultCLI struct {
	FrameSize  uint32
	OfferedPct float64
	FramesTx   uint64
	FramesRx   uint64
	LossPct    float64
}

// BackToBackResultCLI wraps the back-to-back test result for CLI
type BackToBackResultCLI struct {
	FrameSize       uint32
	MaxBurstFrames  uint64
	BurstDurationUs uint64
	Trials          uint32
}

// RecoveryResultCLI wraps the system recovery test result for CLI
type RecoveryResultCLI struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResultCLI wraps the reset test result for CLI
type ResetResultCLI struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}
