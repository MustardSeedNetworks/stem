//go:build cgo && linux

// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This package wraps the high-performance C library for test execution,
// handling packet generation, timing, and result collection.
package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../../build -lreflector -lpthread -lm
#cgo linux LDFLAGS: -lxdp -lbpf

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include "rfc2544_internal.h"

typedef struct {
    rfc2544_ctx_t *ctx;
    int status;
} stem_rfc2544_init_result_t;
static stem_rfc2544_init_result_t stem_rfc2544_initialize(const char *interface) {
    stem_rfc2544_init_result_t result = {0};
    result.status = rfc2544_init(&result.ctx, interface);
    return result;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

const (
	interfaceNameMax = 63
)

// ErrNotSupported is defined for interface parity across build targets.
// in the CGO build since the dataplane is available.
var ErrNotSupported = errors.New("CGO dataplane not available on this platform")

// Context wraps the C rfc2544_ctx_t.
type Context struct {
	ctx       *C.rfc2544_ctx_t
	mu        sync.Mutex
	ctxMu     sync.RWMutex
	stats     Stats
	config    Config
	frameSize uint32
}

// NewContext creates a new RFC2544 test context.
func NewContext(iface string) (*Context, error) {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))

	result := C.stem_rfc2544_initialize(cIface)
	if result.status < 0 {
		return nil, fmt.Errorf("init failed: %d", result.status)
	}

	return &Context{ctx: result.ctx}, nil
}

// Configure applies test configuration.
func (c *Context) Configure(cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ccfg := C.rfc2544_config_alloc_default()
	if ccfg == nil {
		return errors.New("allocate RFC 2544 configuration")
	}
	defer C.rfc2544_config_free(ccfg)

	// Copy interface name
	cIface := C.CString(cfg.Interface)
	defer C.free(unsafe.Pointer(cIface))
	C.strncpy(&ccfg._interface[0], cIface, interfaceNameMax)

	ccfg.line_rate = C.uint64_t(cfg.LineRate)
	ccfg.auto_detect_nic = C.bool(cfg.AutoDetect)
	ccfg.test_type = C.test_type_t(cfg.TestType)
	ccfg.frame_size = C.uint32_t(cfg.FrameSize)
	ccfg.include_jumbo = C.bool(cfg.IncludeJumbo)
	ccfg.trial_duration_sec = C.uint32_t(cfg.TrialDuration.Seconds())
	ccfg.warmup_sec = C.uint32_t(cfg.WarmupPeriod.Seconds())
	ccfg.initial_rate_pct = C.double(cfg.InitialRatePct)
	ccfg.resolution_pct = C.double(cfg.ResolutionPct)
	ccfg.max_iterations = C.uint32_t(cfg.MaxIterations)
	ccfg.acceptable_loss = C.double(cfg.AcceptableLoss)
	ccfg.hw_timestamp = C.bool(cfg.HWTimestamp)
	ccfg.measure_latency = C.bool(cfg.MeasureLatency)
	ccfg.use_pacing = C.bool(cfg.UsePacing)
	ccfg.batch_size = C.uint32_t(cfg.BatchSize)
	if cfg.Y1564StepDuration > 0 {
		ccfg.y1564.step_duration_sec = C.uint32_t(cfg.Y1564StepDuration.Seconds())
	}

	if ret := C.rfc2544_configure(c.ctx, ccfg); ret < 0 {
		return fmt.Errorf("configure failed: %d", ret)
	}
	peer, err := resolvePeer(cfg.Interface, cfg.Peer, cfg.PeerPort)
	if err != nil {
		return err
	}
	if ret := C.rfc2544_set_peer(c.ctx, (*C.uint8_t)(&peer.remoteMAC[0]),
		(*C.uint8_t)(&peer.localIP[0]), (*C.uint8_t)(&peer.remoteIP[0]),
		C.uint16_t(peer.sourcePort), C.uint16_t(peer.remotePort)); ret < 0 {
		return fmt.Errorf("configure peer failed: %d", ret)
	}

	return nil
}

// Run starts the configured test.
func (c *Context) Run() error {
	c.ctxMu.RLock()
	defer c.ctxMu.RUnlock()
	if c.ctx == nil {
		return errors.New("dataplane context is closed")
	}
	ret := C.rfc2544_run(c.ctx)
	if ret < 0 {
		return fmt.Errorf("run failed: %d", ret)
	}
	return nil
}

// Cancel stops a running test.
func (c *Context) Cancel() {
	c.ctxMu.RLock()
	defer c.ctxMu.RUnlock()
	if c.ctx != nil {
		C.rfc2544_cancel(c.ctx)
	}
}

// State returns the current test state.
func (c *Context) State() TestState {
	c.ctxMu.RLock()
	defer c.ctxMu.RUnlock()
	if c.ctx == nil {
		return StateIdle
	}
	return TestState(C.rfc2544_get_state(c.ctx))
}

// Close cleans up resources.
func (c *Context) Close() {
	c.ctxMu.Lock()
	defer c.ctxMu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ctx != nil {
		C.rfc2544_cleanup(c.ctx)
		c.ctx = nil
	}
}

// RunCustomStreamTest executes a custom traffic stream.
func (c *Context) RunCustomStreamTest(cfg *TrafficGenConfig) (*TrafficGenResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	frameSize, ratePct, durationSec, warmupSec, streamID := customStreamParameters(cfg)

	signature := C.CString("CUSTOM ")
	defer C.free(unsafe.Pointer(signature))

	var cResult C.trial_result_t
	ret := C.run_trial_custom(c.ctx, C.uint32_t(frameSize), C.double(ratePct), C.uint32_t(durationSec),
		C.uint32_t(warmupSec), signature, C.uint32_t(streamID), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("custom stream test failed: %d", ret)
	}

	return &TrafficGenResult{
		PacketsSent:  uint64(cResult.packets_sent),
		PacketsRecv:  uint64(cResult.packets_recv),
		BytesSent:    uint64(cResult.bytes_sent),
		LossPct:      float64(cResult.loss_pct),
		ElapsedSec:   float64(cResult.elapsed_sec),
		AchievedPPS:  float64(cResult.achieved_pps),
		AchievedMbps: float64(cResult.achieved_mbps),
		Latency:      newLatencyStats(cResult.latency),
	}, nil
}

// RunSystemRecoveryTest runs RFC 2544 Section 26.5 System Recovery test.
func (c *Context) RunSystemRecoveryTest(throughputPct float64, overloadSec uint32) (*RecoveryResultCLI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.recovery_result_t

	ret := C.rfc2544_system_recovery_test(c.ctx, C.uint32_t(c.frameSize),
		C.double(throughputPct), C.uint32_t(overloadSec), &result)
	if ret < 0 {
		return nil, fmt.Errorf("system recovery test failed: %d", ret)
	}

	return &RecoveryResultCLI{
		FrameSize:       uint32(result.frame_size),
		OverloadRatePct: float64(result.overload_rate_pct),
		RecoveryRatePct: float64(result.recovery_rate_pct),
		OverloadSec:     uint32(result.overload_sec),
		RecoveryTimeMs:  float64(result.recovery_time_ms),
		FramesLost:      uint64(result.frames_lost),
		Trials:          uint32(result.trials),
	}, nil
}

// RunResetTest runs RFC 2544 Section 26.6 Reset test.
func (c *Context) RunResetTest() (*ResetResultCLI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.reset_result_t

	ret := C.rfc2544_reset_test(c.ctx, C.uint32_t(c.frameSize), &result)
	if ret < 0 {
		return nil, fmt.Errorf("reset test failed: %d", ret)
	}

	return &ResetResultCLI{
		FrameSize:   uint32(result.frame_size),
		ResetTimeMs: float64(result.reset_time_ms),
		FramesLost:  uint64(result.frames_lost),
		Trials:      uint32(result.trials),
		ManualReset: bool(result.manual_reset),
	}, nil
}

// newLatencyStats copies one C latency block across the boundary. Three call
// sites spelled this out identically before.
func newLatencyStats(l C.latency_stats_t) LatencyStats {
	return LatencyStats{
		Count:    uint64(l.count),
		MinNs:    float64(l.min_ns),
		MaxNs:    float64(l.max_ns),
		AvgNs:    float64(l.avg_ns),
		JitterNs: float64(l.jitter_ns),
		P50Ns:    float64(l.p50_ns),
		P95Ns:    float64(l.p95_ns),
		P99Ns:    float64(l.p99_ns),
	}
}

// Internal wrappers for the existing methods.
func (c *Context) runThroughputTestInternal(frameSize uint32) ([]ThroughputResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 8
	results := make([]C.throughput_result_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_throughput_test(c.ctx, C.uint32_t(frameSize), &results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("throughput test failed: %d", ret)
	}

	// MaxRate* are what the generator measurably put on the wire and
	// OfferedRatePct is the rate the binary search settled on; they are
	// separate fields because reporting the second as the first made a run
	// that carried ~110 pps claim 99.90 % of 10 Gbps (#1233).
	goResults := make([]ThroughputResult, count)
	for i := range int(count) {
		goResults[i] = ThroughputResult{
			FrameSize:        uint32(results[i].frame_size),
			MaxRatePct:       float64(results[i].max_rate_pct),
			MaxRateMbps:      float64(results[i].max_rate_mbps),
			MaxRatePps:       float64(results[i].max_rate_pps),
			OfferedRatePct:   float64(results[i].offered_rate_pct),
			GeneratorLimited: bool(results[i].generator_limited),
			FramesTested:     uint64(results[i].frames_tested),
			FramesReceived:   uint64(results[i].frames_received),
			Iterations:       uint32(results[i].iterations),
			Latency:          newLatencyStats(results[i].latency),
		}
	}

	return goResults, nil
}

func (c *Context) runLatencyTestInternal(frameSize uint32, loadPct float64) (*LatencyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.latency_result_t
	ret := C.rfc2544_latency_test(c.ctx, C.uint32_t(frameSize), C.double(loadPct), &result)
	if ret < 0 {
		return nil, fmt.Errorf("latency test failed: %d", ret)
	}

	return &LatencyResult{
		FrameSize:      uint32(result.frame_size),
		OfferedRatePct: float64(result.offered_rate_pct),
		Latency:        newLatencyStats(result.latency),
	}, nil
}

func (c *Context) runFrameLossTestInternal(frameSize uint32) ([]FrameLossPoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 20
	results := make([]C.frame_loss_point_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_frame_loss_test(c.ctx, C.uint32_t(frameSize), &results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("frame loss test failed: %d", ret)
	}

	goResults := make([]FrameLossPoint, count)
	for i := range int(count) {
		goResults[i] = FrameLossPoint{
			OfferedRatePct: float64(results[i].offered_rate_pct),
			ActualRateMbps: float64(results[i].actual_rate_mbps),
			FramesSent:     uint64(results[i].frames_sent),
			FramesRecv:     uint64(results[i].frames_recv),
			LossPct:        float64(results[i].loss_pct),
		}
	}

	return goResults, nil
}

func (c *Context) runBackToBackTestInternal(frameSize uint32) (*BurstResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.burst_result_t
	ret := C.rfc2544_back_to_back_test(c.ctx, C.uint32_t(frameSize), &result)
	if ret < 0 {
		return nil, fmt.Errorf("back-to-back test failed: %d", ret)
	}

	return &BurstResult{
		FrameSize:     uint32(result.frame_size),
		MaxBurst:      uint64(result.max_burst),
		BurstDuration: float64(result.burst_duration),
		Trials:        uint32(result.trials),
	}, nil
}
