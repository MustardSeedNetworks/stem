//go:build cgo && linux

package dataplane

import "time"

// TestType mirrors C test_type_t.
type TestType int

// The RFC 2544, Y.1564 and Y.1731 test types the C dataplane accepts. Order is
// load-bearing: these are iota values passed across the cgo boundary and must
// stay in lockstep with test_type_t in the C header.
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

// TestState mirrors C test_state_t.
type TestState int

// Lifecycle states a dataplane test moves through. Like TestType these are
// iota values shared with C, so the order must match test_state_t.
const (
	StateIdle TestState = iota
	StateRunning
	StateCompleted
	StateFailed
	StateCancelled
)

// LatencyStats contains latency measurements.
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

// ThroughputResult from binary search test.
type ThroughputResult struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	MaxRatePps   float64
	FramesTested uint64
	Iterations   uint32
	Latency      LatencyStats
}

// FrameLossPoint for a single load level.
type FrameLossPoint struct {
	OfferedRatePct float64
	ActualRateMbps float64
	FramesSent     uint64
	FramesRecv     uint64
	LossPct        float64
}

// LatencyResult from latency test.
type LatencyResult struct {
	FrameSize      uint32
	OfferedRatePct float64
	Latency        LatencyStats
}

// BurstResult from back-to-back test.
type BurstResult struct {
	FrameSize     uint32
	MaxBurst      uint64
	BurstDuration float64
	Trials        uint32
}

// RecoveryResult from RFC 2544 Section 26.5 System Recovery test.
type RecoveryResult struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResult from RFC 2544 Section 26.6 Reset test.
type ResetResult struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// Y1564SLA contains SLA parameters for Y.1564 testing.
type Y1564SLA struct {
	CIRMbps         float64
	EIRMbps         float64
	CBSBytes        uint32
	EBSBytes        uint32
	FDThresholdMs   float64
	FDVThresholdMs  float64
	FLRThresholdPct float64
}

// Y1564Service represents a service configuration for Y.1564 testing.
type Y1564Service struct {
	ServiceID   uint32
	ServiceName string
	SLA         Y1564SLA
	FrameSize   uint32
	CoS         uint8
	Enabled     bool
}

// Y1564StepResult from a Y.1564 configuration test step.
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

// Y1564ConfigResult from Y.1564 service configuration test.
type Y1564ConfigResult struct {
	ServiceID   uint32
	Steps       [4]Y1564StepResult
	ServicePass bool
}

// Y1564PerfResult from Y.1564 service performance test.
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

// RFC2889Config holds the shared test parameters for RFC 2889 LAN-switch
// benchmarks (address caching, address learning rate, forwarding, broadcast,
// and congestion control).
type RFC2889Config struct {
	FrameSize         uint32
	DurationSec       uint32
	WarmupSec         uint32
	AddressCount      uint32
	AcceptableLossPct float64
	PortCount         uint32
	Pattern           uint32
}

// RFC2889ForwardingResult is the outcome of the RFC 2889 forwarding-rate
// test: the binary-search maximum rate the DUT forwards without loss across
// PortCount ports using the given traffic Pattern.
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

// RFC2889CachingResult is the outcome of the RFC 2889 address-caching-capacity
// test: whether the DUT sustains AddressCount learned addresses without
// dropping frames.
type RFC2889CachingResult struct {
	AddressCount uint32
	FrameSize    uint32
	PortCount    uint32
	FramesTx     uint64
	FramesRx     uint64
	LossPct      float64
	Passed       bool
}

// RFC2889LearningResult is the outcome of the RFC 2889 address-learning-rate
// test: how quickly the DUT learns new source MAC addresses and whether
// forwarding stays correct once learning completes.
type RFC2889LearningResult struct {
	FrameSize           uint32
	PortCount           uint32
	LearningRateFps     float64
	AddressesLearned    uint32
	LearningTimeMs      float64
	VerificationFrames  uint32
	VerificationLossPct float64
}

// RFC2889BroadcastResult is the outcome of the RFC 2889 broadcast
// forwarding/replication test across IngressPorts and EgressPorts.
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

// RFC2889CongestionResult is the outcome of the RFC 2889 congestion-control
// test: head-of-line blocking and pause-frame/backpressure behavior when the
// DUT is driven past its forwarding rate.
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

// RFC6349Config holds the test parameters for an RFC 6349 TCP throughput
// test: target rate, RTT range, receive window, parallel streams, and MSS.
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

// RFC6349Result is the outcome of an RFC 6349 TCP throughput test: the
// achieved rate against the BDP-derived theoretical maximum, TCP efficiency,
// and buffer delay.
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

// TCPPathInfo captures the network-path characteristics (MTU, RTT, BDP,
// bottleneck bandwidth) used to plan an RFC 6349 test's window and rate.
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

// Y1731Config holds the parameters for an ITU-T Y.1731 Ethernet OAM session:
// maintenance end point (MEP) identity, maintenance entity group (MEG) level
// and ID, and the CCM/measurement cadence.
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

// Y1731DelayResult is the outcome of a Y.1731 ETH-DM (delay measurement)
// test.
type Y1731DelayResult struct {
	FramesSent       uint32
	FramesReceived   uint32
	FramesLost       uint32
	DelayMinUs       float64
	DelayAvgUs       float64
	DelayMaxUs       float64
	DelayVariationUs float64
}

// Y1731LossResult is the outcome of a Y.1731 ETH-LM (loss measurement) test,
// including near-end and far-end loss ratios and derived availability.
type Y1731LossResult struct {
	FramesTx         uint64
	FramesRx         uint64
	NearEndLoss      uint64
	FarEndLoss       uint64
	NearEndLossRatio float64
	FarEndLossRatio  float64
	AvailabilityPct  float64
}

// Y1731LoopbackResult is the outcome of a Y.1731 ETH-LB (loopback) test:
// round-trip timing over sent LBM / received LBR frames.
type Y1731LoopbackResult struct {
	LBMSent     uint64
	LBRReceived uint64
	RTTMinMs    float64
	RTTAvgMs    float64
	RTTMaxMs    float64
}

// MEF configuration and results

// MEFConfig holds the MEF/Y.1564 service-attribute configuration for a
// step-load service activation test: CIR/EIR, CBS/EBS, and the frame-delay,
// jitter, loss, and availability thresholds that gate pass/fail.
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

// MEFStepResult is the outcome of a single load step in the MEF/Y.1564
// step-load test.
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

// MEFConfigResult is the overall outcome of the MEF/Y.1564 configuration
// test: the per-CoS step results and the aggregate pass verdict.
type MEFConfigResult struct {
	ServiceID     string
	Steps         [4]MEFStepResult
	NumSteps      uint32
	OverallPassed bool
}

// MEFPerfResult is the outcome of the MEF/Y.1564 sustained performance test
// run at the target rate for DurationSec.
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

// TSNConfig holds the parameters for a TSN test exercising 802.1Qbv gate
// scheduling and 802.1AS time synchronization.
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

// TSNTimingResult is the outcome of the TSN gate-timing-accuracy test.
type TSNTimingResult struct {
	CyclesTested       uint32
	TimingErrors       uint32
	MaxGateDeviationNs float64
	AvgGateDeviationNs float64
	GateTimingPassed   bool
}

// TSNClassResult is the per-traffic-class outcome of a TSN isolation test:
// how much a class's frames were interfered with by other traffic classes.
type TSNClassResult struct {
	FramesTx         uint64
	FramesRx         uint64
	FramesInterfered uint64
	IsolationPct     float64
	LatencyAvgNs     float64
	LatencyMaxNs     float64
	Passed           bool
}

// TSNIsolationResult is the aggregated outcome of the TSN class-isolation
// test across NumClasses traffic classes.
type TSNIsolationResult struct {
	NumClasses    uint32
	ClassResults  [8]TSNClassResult
	OverallPassed bool
}

// TSNLatencyResult is the per-traffic-class latency distribution measured
// during a TSN test.
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

// TSNPTPResult is the outcome of the 802.1AS PTP clock-synchronization
// accuracy test.
type TSNPTPResult struct {
	Samples        uint32
	OffsetAvgNs    float64
	OffsetMaxNs    float64
	OffsetStddevNs float64
	SyncAchieved   bool
}

// TSNFullResult combines the timing, isolation, per-class latency, and PTP
// results of a full TSN test run.
type TSNFullResult struct {
	TimingResult    TSNTimingResult
	IsolationResult TSNIsolationResult
	LatencyResults  [8]TSNLatencyResult
	PTPResult       TSNPTPResult
	OverallPassed   bool
}

// Traffic generation configuration

// TrafficGenConfig holds raw traffic-generation parameters (rate, burst
// pattern, VLAN tagging) used independent of any specific RFC/ITU-T test.
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

// TrafficGenResult is the outcome of a raw traffic-generation run: packets
// and bytes sent/received, achieved rate, and loss.
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

// Config for RFC2544 tests.
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
}

// Stats for real-time monitoring.
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

// ThroughputResultCLI wraps the throughput test result for CLI.
type ThroughputResultCLI struct {
	FrameSize   uint32
	MaxRatePct  float64
	MaxRateMbps float64
	MaxRatePPS  float64
	Iterations  uint32
	Latency     LatencyStats
}

// LatencyResultCLI wraps the latency test result for CLI.
type LatencyResultCLI struct {
	FrameSize uint32
	LoadPct   float64
	Latency   LatencyStats
}

// FrameLossResultCLI wraps the frame loss test result for CLI.
type FrameLossResultCLI struct {
	FrameSize  uint32
	OfferedPct float64
	FramesTx   uint64
	FramesRx   uint64
	LossPct    float64
}

// BackToBackResultCLI wraps the back-to-back test result for CLI.
type BackToBackResultCLI struct {
	FrameSize       uint32
	MaxBurstFrames  uint64
	BurstDurationUs uint64
	Trials          uint32
}

// RecoveryResultCLI wraps the system recovery test result for CLI.
type RecoveryResultCLI struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResultCLI wraps the reset test result for CLI.
type ResetResultCLI struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}
