// SPDX-License-Identifier: BUSL-1.1

package api

// Configuration constants.
const (
	DefaultPortFilter = 3842
	DefaultOUIFilter  = "00:c0:17" // NetAlly OUI.
	DefaultProfile    = "all"
)

// Status constants for test execution state.
const (
	statusIdle      = "idle"
	statusStarting  = "starting"
	statusRunning   = "running"
	statusCompleted = "completed"
	statusError     = "error"
	statusStopped   = "stopped"
	statusCancelled = "cancelled"
)

// Module name constants.
const (
	moduleReflector   = "reflector"
	moduleBenchmark   = "benchmark"
	moduleServicetest = "servicetest"
	moduleTrafficgen  = "trafficgen"
	moduleMeasure     = "measure"
	moduleCertify     = "certify"
)

// Operating mode constants.
const (
	modeReflector  = "reflector"
	modeTestMaster = "test_master"
)

// Test type constants.
const (
	testTypeReflect    = "reflect"
	testTypeThroughput = "throughput"
)

// StatusResponse for simple status messages.
type StatusResponse struct {
	Status string `json:"status"`
}

// ModeResponse for mode queries.
type ModeResponse struct {
	Mode string `json:"mode"`
}

// ModeUpdateResponse for mode updates.
//
// Status is "updated" when the mode actually changed and "unchanged"
// when a POST asked for the mode the server was already in. Previous
// is always the mode the server was in before this request; when
// Status == "unchanged" it equals Mode.
type ModeUpdateResponse struct {
	Status   string `json:"status"`
	Mode     string `json:"mode"`
	Previous string `json:"previous"`
}

// SettingsResponse for settings queries.
type SettingsResponse struct {
	Mode      string `json:"mode"`
	Interface string `json:"interface"`
	Theme     string `json:"theme"`
}

// SettingsUpdate for settings POST requests.
type SettingsUpdate struct {
	Interface string `json:"interface,omitempty"`
	Theme     string `json:"theme,omitempty"`
}

// HealthResponse for health checks.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Product string `json:"product"`
	Company string `json:"company"`
	Uptime  int64  `json:"uptime"`
}

// TrialStatusResponse for trial queries.
type TrialStatusResponse struct {
	Active        bool `json:"active"`
	DaysRemaining int  `json:"daysRemaining,omitempty"`
}

// ErrorResponse for error messages.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestStartRequest for starting a test.
type TestStartRequest struct {
	TestType  string      `json:"testType"`
	Interface string      `json:"interface,omitempty"`
	Mode      string      `json:"mode,omitempty"`    // reflector or test_master
	Profile   string      `json:"profile,omitempty"` // reflector profile
	Tests     []string    `json:"tests,omitempty"`   // selected test types
	Config    *TestConfig `json:"config,omitempty"`  // full test configuration
}

// TestConfig contains all module-specific configurations.
type TestConfig struct {
	RFC2544    *RFC2544TestConfig    `json:"rfc2544,omitempty"`
	RFC2889    *RFC2889TestConfig    `json:"rfc2889,omitempty"`
	RFC6349    *RFC6349TestConfig    `json:"rfc6349,omitempty"`
	Y1564      *Y1564TestConfig      `json:"y1564,omitempty"`
	Y1731      *Y1731TestConfig      `json:"y1731,omitempty"`
	TSN        *TSNTestConfig        `json:"tsn,omitempty"`
	TrafficGen *TrafficGenTestConfig `json:"trafficGen,omitempty"`
}

// RFC2544TestConfig matches frontend RFC2544Config.
type RFC2544TestConfig struct {
	Duration      int      `json:"duration"`      // seconds
	FrameSizes    []uint32 `json:"frameSizes"`    // frame sizes to test
	Resolution    float64  `json:"resolution"`    // binary search resolution %
	MaxLoss       float64  `json:"maxLoss"`       // max acceptable loss %
	Warmup        int      `json:"warmup"`        // warmup seconds
	Trials        int      `json:"trials"`        // trials per test point
	StepSize      float64  `json:"stepSize"`      // frame loss step size %
	Bidirectional bool     `json:"bidirectional"` // bidirectional testing
}

// RFC2889TestConfig matches frontend RFC2889Config.
type RFC2889TestConfig struct {
	FrameSize      uint32  `json:"frameSize"`
	Duration       uint32  `json:"duration"`
	Warmup         uint32  `json:"warmup"`
	AddressCount   uint32  `json:"addressCount"`
	AcceptableLoss float64 `json:"acceptableLoss"`
	PortCount      uint32  `json:"portCount"`
	Pattern        uint32  `json:"pattern"` // 0=mesh, 1=pair, 2=broadcast
}

// RFC6349TestConfig matches frontend RFC6349Config.
type RFC6349TestConfig struct {
	TargetRateMbps  float64 `json:"targetRateMbps"`
	MinRTTMs        float64 `json:"minRTTMs"`
	MaxRTTMs        float64 `json:"maxRTTMs"`
	RWNDSize        uint32  `json:"rwndSize"`
	Duration        uint32  `json:"duration"`
	ParallelStreams uint32  `json:"parallelStreams"`
	MSS             uint32  `json:"mss"`
	Mode            uint32  `json:"mode"` // 0=bidirectional, 1=upstream, 2=downstream
}

// Y1564TestConfig matches frontend Y1564Config.
type Y1564TestConfig struct {
	CIR                float64  `json:"cir"` // Committed Information Rate Mbps
	EIR                float64  `json:"eir"` // Excess Information Rate Mbps
	CBS                uint32   `json:"cbs"` // Committed Burst Size KB
	EBS                uint32   `json:"ebs"` // Excess Burst Size KB
	FrameSizes         []uint32 `json:"frameSizes"`
	ConfigStepDuration uint32   `json:"configStepDuration"` // seconds
	PerfTestDuration   uint32   `json:"perfTestDuration"`   // seconds
	VlanID             uint16   `json:"vlanId"`
	PCP                uint8    `json:"pcp"` // Priority Code Point
	ColorAware         bool     `json:"colorAware"`
	FLRThreshold       float64  `json:"flrThreshold"` // Frame Loss Ratio %
	FDThreshold        float64  `json:"fdThreshold"`  // Frame Delay ms
	FDVThreshold       float64  `json:"fdvThreshold"` // Frame Delay Variation ms
}

// Y1731TestConfig matches frontend Y1731Config.
type Y1731TestConfig struct {
	MepID          uint32 `json:"mepId"`
	MegLevel       uint32 `json:"megLevel"`
	MegID          string `json:"megId"`
	CCMInterval    uint32 `json:"ccmInterval"`
	Priority       uint8  `json:"priority"`
	Duration       uint32 `json:"duration"`
	IntervalMs     uint32 `json:"intervalMs"`
	Count          uint32 `json:"count"`
	FrameSize      uint32 `json:"frameSize"`
	PriorityTagged bool   `json:"priorityTagged"`
}

// TSNTestConfig matches frontend TSNConfig.
type TSNTestConfig struct {
	Duration          uint32 `json:"duration"`
	Warmup            uint32 `json:"warmup"`
	FrameSize         uint32 `json:"frameSize"`
	MaxLatencyNs      uint32 `json:"maxLatencyNs"`
	MaxJitterNs       uint32 `json:"maxJitterNs"`
	RequirePTPSync    bool   `json:"requirePTPSync"`
	MaxSyncOffsetNs   uint32 `json:"maxSyncOffsetNs"`
	PTPEnabled        bool   `json:"ptpEnabled"`
	PreemptionEnabled bool   `json:"preemptionEnabled"`
	NumTrafficClasses uint32 `json:"numTrafficClasses"`
	BaseTimeNs        uint64 `json:"baseTimeNs"`
	CycleTimeNs       uint32 `json:"cycleTimeNs"`
	TrafficClass      uint32 `json:"trafficClass"`
}

// TrafficGenTestConfig matches frontend TrafficGenConfig.
type TrafficGenTestConfig struct {
	FrameSize       uint32  `json:"frameSize"`
	RatePct         float64 `json:"ratePct"`
	Duration        uint32  `json:"duration"`
	Warmup          uint32  `json:"warmup"`
	StreamID        uint32  `json:"streamId"`
	BurstMode       bool    `json:"burstMode"`
	BurstSize       uint32  `json:"burstSize"`
	InterBurstGapUs uint32  `json:"interBurstGapUs"`
	SrcMac          string  `json:"srcMac"`
	DstMac          string  `json:"dstMac"`
	VlanID          uint16  `json:"vlanId"`
	VlanPriority    uint8   `json:"vlanPriority"`
}

// TestStartResponse for test start confirmation.
type TestStartResponse struct {
	Status   string `json:"status"`
	TestType string `json:"testType"`
	Module   string `json:"module"`
	Message  string `json:"message,omitempty"`
}

// AuthLoginRequest captures credentials supplied to /api/auth/login.
type AuthLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthLoginResponse returns the issued JWT tokens.
type AuthLoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// AuthRefreshRequest captures the refresh token for /api/auth/refresh.
type AuthRefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// TestResultResponse for completed test results.
type TestResultResponse struct {
	Status   string `json:"status"`
	TestType string `json:"testType,omitempty"`
	Module   string `json:"module,omitempty"`
	Success  bool   `json:"success,omitempty"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	Data     any    `json:"data,omitempty"`
}

// ModeRequest for mode POST requests. Valid values are mirrored from
// the modeReflector / modeTestMaster constants above.
type ModeRequest struct {
	Mode string `json:"mode" validate:"required,oneof=reflector test_master"`
}

// ReflectorConfig holds reflector-specific settings.
type ReflectorConfig struct {
	Profile         string   `json:"profile"` // netally, msn, all, custom.
	SignatureFilter []string `json:"signatureFilter"`
	OUIFilter       string   `json:"ouiFilter"`
	PortFilter      int      `json:"portFilter"`
}

// ReflectorStats holds reflector-specific statistics.
type ReflectorStats struct {
	Running          bool    `json:"running"`
	PacketsReceived  uint64  `json:"packetsReceived"`
	PacketsReflected uint64  `json:"packetsReflected"`
	BytesReceived    uint64  `json:"bytesReceived"`
	BytesReflected   uint64  `json:"bytesReflected"`
	TxErrors         uint64  `json:"txErrors"`
	RxInvalid        uint64  `json:"rxInvalid"`
	RatePPS          float64 `json:"ratePps"`
	RateMbps         float64 `json:"rateMbps"`
	Signatures       struct {
		ProbeOT uint64 `json:"probeot"`
		DataOT  uint64 `json:"dataot"`
		Latency uint64 `json:"latency"`
		RFC2544 uint64 `json:"rfc2544"`
		Y1564   uint64 `json:"y1564"`
		MSN     uint64 `json:"msn"`
	} `json:"signatures"`
	Latency struct {
		MinUs float64 `json:"minUs"`
		AvgUs float64 `json:"avgUs"`
		MaxUs float64 `json:"maxUs"`
		Count uint64  `json:"count"`
	} `json:"latency"`
	Uptime float64 `json:"uptime"`
}

// Stats holds runtime statistics.
type Stats struct {
	PacketsReceived uint64  `json:"packetsReceived"`
	PacketsSent     uint64  `json:"packetsSent"`
	BytesReceived   uint64  `json:"bytesReceived"`
	BytesSent       uint64  `json:"bytesSent"`
	CurrentPPS      float64 `json:"currentPps"`
	CurrentMbps     float64 `json:"currentMbps"`
	Uptime          int64   `json:"uptime"`
	TestStatus      string  `json:"testStatus"`
	CurrentTest     *string `json:"currentTest"`
}

// LicenseStatus represents the license status response.
type LicenseStatus struct {
	Activated     bool     `json:"activated"`
	IsTrialMode   bool     `json:"isTrialMode"`
	Tier          int      `json:"tier"`
	TierName      string   `json:"tierName"`
	DaysRemaining int      `json:"daysRemaining"`
	Features      []string `json:"features"`
	DeviceHash    string   `json:"deviceHash"`
	LicenseKey    string   `json:"licenseKey,omitempty"`
	Message       string   `json:"message,omitempty"`
}

// LicenseActivateRequest for license activation.
type LicenseActivateRequest struct {
	LicenseKey string `json:"licenseKey" validate:"required"`
}

// LivenessResponse for Kubernetes liveness probe.
type LivenessResponse struct {
	Status string `json:"status"`
}

// ReadinessCheck represents an individual readiness check result.
type ReadinessCheck struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// ReadinessResponse for Kubernetes readiness probe.
type ReadinessResponse struct {
	Status string                    `json:"status"`
	Checks map[string]ReadinessCheck `json:"checks"`
}
