//go:build !cgo || !linux

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This file contains stub implementations for non-CGO or non-Linux builds.
// The actual test execution requires CGO and Linux for packet generation.
package dataplane

import (
	"fmt"
	"time"
)

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

type TestState int

const (
	StateIdle TestState = iota
	StateRunning
	StateCompleted
	StateFailed
	StateCancelled
)

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

type ThroughputResult struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	MaxRatePps   float64
	FramesTested uint64
	Iterations   uint32
	Latency      LatencyStats
}

type FrameLossPoint struct {
	OfferedRatePct float64
	ActualRateMbps float64
	FramesSent     uint64
	FramesRecv     uint64
	LossPct        float64
}

type LatencyResult struct {
	FrameSize      uint32
	OfferedRatePct float64
	Latency        LatencyStats
}

type BurstResult struct {
	FrameSize     uint32
	MaxBurst      uint64
	BurstDuration float64
	Trials        uint32
}

type RecoveryResult struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

type ResetResult struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

type Y1564SLA struct {
	CIRMbps         float64
	EIRMbps         float64
	CBSBytes        uint32
	EBSBytes        uint32
	FDThresholdMs   float64
	FDVThresholdMs  float64
	FLRThresholdPct float64
}

type Y1564Service struct {
	ServiceID   uint32
	ServiceName string
	SLA         Y1564SLA
	FrameSize   uint32
	CoS         uint8
	Enabled     bool
}

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

type Y1564ConfigResult struct {
	ServiceID   uint32
	Steps       [4]Y1564StepResult
	ServicePass bool
}

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

type Context struct {
	config    Config //nolint:unused // placeholder for CGO implementation
	frameSize uint32
}

type Stats struct {
	TxPackets   uint64
	TxBytes     uint64
	RxPackets   uint64
	RxBytes     uint64
	CurrentRate float64
	Progress    float64
	Timestamp   time.Time
}

type ThroughputResultCLI struct {
	FrameSize   uint32
	MaxRatePct  float64
	MaxRateMbps float64
	MaxRatePPS  float64
	Iterations  uint32
	Latency     LatencyStats
}

type LatencyResultCLI struct {
	FrameSize uint32
	LoadPct   float64
	Latency   LatencyStats
}

type FrameLossResultCLI struct {
	FrameSize  uint32
	OfferedPct float64
	FramesTx   uint64
	FramesRx   uint64
	LossPct    float64
}

type BackToBackResultCLI struct {
	FrameSize       uint32
	MaxBurstFrames  uint64
	BurstDurationUs uint64
	Trials          uint32
}

type RecoveryResultCLI struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

type ResetResultCLI struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

var ErrNotSupported = fmt.Errorf("CGO dataplane not available on this platform")

func NewContext(iface string) (*Context, error) {
	return nil, ErrNotSupported
}

func New(cfg Config) (*Context, error) {
	return nil, ErrNotSupported
}

func (c *Context) Configure(cfg *Config) error {
	return ErrNotSupported
}

func (c *Context) Run() error {
	return ErrNotSupported
}

func (c *Context) Cancel() {}

func (c *Context) State() TestState {
	return StateIdle
}

func (c *Context) Close() {}

func (c *Context) SetFrameSize(frameSize uint32) {
	if c != nil {
		c.frameSize = frameSize
	}
}

func (c *Context) RunThroughputTest() (*ThroughputResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunLatencyTest(loadLevels []float64) ([]LatencyResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunFrameLossTest(startPct, endPct, stepPct float64) ([]FrameLossResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunBackToBackTest(initialBurst uint64, trials uint32) (*BackToBackResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunSystemRecoveryTest(throughputPct float64, overloadSec uint32) (*RecoveryResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunResetTest() (*ResetResultCLI, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunY1564ConfigTest(service *Y1564Service) (*Y1564ConfigResult, error) {
	return nil, ErrNotSupported
}

func (c *Context) RunY1564PerfTest(service *Y1564Service, durationSec uint32) (*Y1564PerfResult, error) {
	return nil, ErrNotSupported
}

func GetLineRate(iface string) uint64 {
	return 0
}

func CalcPPS(lineRate uint64, frameSize uint32) uint64 {
	return 0
}
