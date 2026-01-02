// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

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
type ModeUpdateResponse struct {
	Status string `json:"status"`
	Mode   string `json:"mode"`
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
	TestType  string `json:"testType"`
	Interface string `json:"interface,omitempty"`
	FrameSize uint32 `json:"frameSize,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

// TestStartResponse for test start confirmation.
type TestStartResponse struct {
	Status   string `json:"status"`
	TestType string `json:"testType"`
	Module   string `json:"module"`
	Message  string `json:"message,omitempty"`
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

// ModeRequest for mode POST requests.
type ModeRequest struct {
	Mode string `json:"mode"`
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
	LicenseKey string `json:"licenseKey"`
}
