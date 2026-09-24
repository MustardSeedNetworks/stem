//go:build cgo && linux

package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../../build -lreflector -lpthread -lm
#cgo linux LDFLAGS: -lxdp -lbpf

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include "rfc2544.h"
*/
import "C"
import "fmt"

// RunTSNGateTimingTest executes TSN gate timing test.
func (c *Context) RunTSNGateTimingTest(cfg *TSNConfig) (*TSNTimingResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.tsn_config_t
	C.tsn_default_config(&cCfg)
	fillTSNConfig(&cCfg, cfg)

	var cResult C.tsn_timing_result_v2_t
	ret := C.tsn_gate_timing_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("TSN gate timing test failed: %d", ret)
	}

	return &TSNTimingResult{
		CyclesTested:       uint32(cResult.cycles_tested),
		TimingErrors:       uint32(cResult.timing_errors),
		MaxGateDeviationNs: float64(cResult.max_gate_deviation_ns),
		AvgGateDeviationNs: float64(cResult.avg_gate_deviation_ns),
		GateTimingPassed:   bool(cResult.gate_timing_passed),
	}, nil
}

// RunTSNIsolationTest executes TSN traffic class isolation test.
func (c *Context) RunTSNIsolationTest(cfg *TSNConfig) (*TSNIsolationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.tsn_config_t
	C.tsn_default_config(&cCfg)
	fillTSNConfig(&cCfg, cfg)

	var cResult C.tsn_isolation_result_t
	ret := C.tsn_isolation_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("TSN isolation test failed: %d", ret)
	}

	result := &TSNIsolationResult{
		NumClasses:    uint32(cResult.num_classes),
		OverallPassed: bool(cResult.overall_passed),
	}
	for i := range len(result.ClassResults) {
		cr := cResult.class_results[i]
		result.ClassResults[i] = TSNClassResult{
			FramesTx:         uint64(cr.frames_tx),
			FramesRx:         uint64(cr.frames_rx),
			FramesInterfered: uint64(cr.frames_interfered),
			IsolationPct:     float64(cr.isolation_pct),
			LatencyAvgNs:     float64(cr.latency_avg_ns),
			LatencyMaxNs:     float64(cr.latency_max_ns),
			Passed:           bool(cr.passed),
		}
	}

	return result, nil
}

// RunTSNLatencyTest executes TSN scheduled latency test.
func (c *Context) RunTSNLatencyTest(cfg *TSNConfig) (*TSNLatencyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.tsn_config_t
	C.tsn_default_config(&cCfg)
	fillTSNConfig(&cCfg, cfg)

	trafficClass := uint32(0)
	if cfg != nil && cfg.TrafficClass > 0 {
		trafficClass = cfg.TrafficClass
	}

	var cResult C.tsn_latency_result_t
	ret := C.tsn_scheduled_latency_test(c.ctx, &cCfg, C.uint32_t(trafficClass), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("TSN scheduled latency test failed: %d", ret)
	}

	return &TSNLatencyResult{
		TrafficClass:  uint32(cResult.traffic_class),
		Samples:       uint32(cResult.samples),
		LatencyMinNs:  float64(cResult.latency_min_ns),
		LatencyAvgNs:  float64(cResult.latency_avg_ns),
		LatencyMaxNs:  float64(cResult.latency_max_ns),
		Latency99Ns:   float64(cResult.latency_99_ns),
		Latency999Ns:  float64(cResult.latency_999_ns),
		JitterNs:      float64(cResult.jitter_ns),
		LatencyPassed: bool(cResult.latency_passed),
		JitterPassed:  bool(cResult.jitter_passed),
		OverallPassed: bool(cResult.overall_passed),
	}, nil
}

// RunTSNFullTest executes TSN full test suite.
func (c *Context) RunTSNFullTest(cfg *TSNConfig) (*TSNFullResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.tsn_config_t
	C.tsn_default_config(&cCfg)
	fillTSNConfig(&cCfg, cfg)

	var cResult C.tsn_full_result_t
	ret := C.tsn_full_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("TSN full test failed: %d", ret)
	}

	result := &TSNFullResult{
		TimingResult: TSNTimingResult{
			CyclesTested:       uint32(cResult.timing_result.cycles_tested),
			TimingErrors:       uint32(cResult.timing_result.timing_errors),
			MaxGateDeviationNs: float64(cResult.timing_result.max_gate_deviation_ns),
			AvgGateDeviationNs: float64(cResult.timing_result.avg_gate_deviation_ns),
			GateTimingPassed:   bool(cResult.timing_result.gate_timing_passed),
		},
		IsolationResult: TSNIsolationResult{
			NumClasses:    uint32(cResult.isolation_result.num_classes),
			OverallPassed: bool(cResult.isolation_result.overall_passed),
		},
		PTPResult: TSNPTPResult{
			Samples:        uint32(cResult.ptp_result.samples),
			OffsetAvgNs:    float64(cResult.ptp_result.offset_avg_ns),
			OffsetMaxNs:    float64(cResult.ptp_result.offset_max_ns),
			OffsetStddevNs: float64(cResult.ptp_result.offset_stddev_ns),
			SyncAchieved:   bool(cResult.ptp_result.sync_achieved),
		},
		OverallPassed: bool(cResult.overall_passed),
	}

	for i := range len(result.IsolationResult.ClassResults) {
		cr := cResult.isolation_result.class_results[i]
		result.IsolationResult.ClassResults[i] = TSNClassResult{
			FramesTx:         uint64(cr.frames_tx),
			FramesRx:         uint64(cr.frames_rx),
			FramesInterfered: uint64(cr.frames_interfered),
			IsolationPct:     float64(cr.isolation_pct),
			LatencyAvgNs:     float64(cr.latency_avg_ns),
			LatencyMaxNs:     float64(cr.latency_max_ns),
			Passed:           bool(cr.passed),
		}
	}

	for i := range len(result.LatencyResults) {
		lr := cResult.latency_results[i]
		result.LatencyResults[i] = TSNLatencyResult{
			TrafficClass:  uint32(lr.traffic_class),
			Samples:       uint32(lr.samples),
			LatencyMinNs:  float64(lr.latency_min_ns),
			LatencyAvgNs:  float64(lr.latency_avg_ns),
			LatencyMaxNs:  float64(lr.latency_max_ns),
			Latency99Ns:   float64(lr.latency_99_ns),
			Latency999Ns:  float64(lr.latency_999_ns),
			JitterNs:      float64(lr.jitter_ns),
			LatencyPassed: bool(lr.latency_passed),
			JitterPassed:  bool(lr.jitter_passed),
			OverallPassed: bool(lr.overall_passed),
		}
	}

	return result, nil
}

func fillTSNConfig(cCfg *C.tsn_config_t, cfg *TSNConfig) {
	if cfg == nil {
		return
	}
	if cfg.DurationSec > 0 {
		cCfg.duration_sec = C.uint32_t(cfg.DurationSec)
	}
	if cfg.WarmupSec > 0 {
		cCfg.warmup_sec = C.uint32_t(cfg.WarmupSec)
	}
	if cfg.FrameSize > 0 {
		cCfg.frame_size = C.uint32_t(cfg.FrameSize)
	}
	if cfg.MaxLatencyNs > 0 {
		cCfg.max_latency_ns = C.uint32_t(cfg.MaxLatencyNs)
	}
	if cfg.MaxJitterNs > 0 {
		cCfg.max_jitter_ns = C.uint32_t(cfg.MaxJitterNs)
	}
	cCfg.require_ptp_sync = C.bool(cfg.RequirePTPSync)
	if cfg.MaxSyncOffsetNs > 0 {
		cCfg.max_sync_offset_ns = C.uint32_t(cfg.MaxSyncOffsetNs)
	}
	cCfg.ptp_enabled = C.bool(cfg.PTPEnabled)
	cCfg.preemption_enabled = C.bool(cfg.PreemptionEnabled)
	if cfg.NumTrafficClasses > 0 {
		cCfg.num_traffic_classes = C.uint32_t(cfg.NumTrafficClasses)
	}
	if cfg.BaseTimeNs > 0 {
		cCfg.base_time_ns = C.uint64_t(cfg.BaseTimeNs)
	}
	if cfg.CycleTimeNs > 0 {
		cCfg.cycle_time_ns = C.uint32_t(cfg.CycleTimeNs)
	}
}
