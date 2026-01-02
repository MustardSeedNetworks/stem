//go:build !cgo || !linux

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This file contains stub implementations for non-CGO or non-Linux builds.
// The actual test execution requires CGO and Linux for packet generation.
package dataplane

import (
	"errors"
	"time"
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
	config    Config //nolint:unused // Placeholder for CGO implementation.
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

// ErrNotSupported is returned when CGO dataplane is not available.
var ErrNotSupported = errors.New("CGO dataplane not available on this platform")

// NewContext creates a new test context for the given interface (stub).
func NewContext(_ string) (*Context, error) {
	return nil, ErrNotSupported
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

// GetLineRate returns the line rate for an interface (stub).
func GetLineRate(_ string) uint64 {
	return 0
}

// CalcPPS calculates packets per second for given line rate and frame size (stub).
func CalcPPS(_ uint64, _ uint32) uint64 {
	return 0
}
