//go:build cgo && linux

package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../../build -lreflector -lpthread -lm
#cgo linux LDFLAGS: -lxdp -lbpf

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include "rfc2544.h"
*/
import "C"
import "fmt"

const (
	mefServiceIDMaxLength = 31
	kilobitsPerMegabit    = 1000
	mefMaxFrameSizes      = 7
)

// RunMEFConfigTest executes MEF configuration test.
func (c *Context) RunMEFConfigTest(cfg *MEFConfig) (*MEFConfigResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.mef_config_t
	C.mef_default_config(&cCfg)
	fillMEFConfig(&cCfg, cfg)

	var cResult C.mef_config_result_t
	ret := C.mef_config_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("MEF config test failed: %d", ret)
	}

	result := &MEFConfigResult{
		ServiceID:     C.GoString(&cResult.service_id[0]),
		NumSteps:      uint32(cResult.num_steps),
		OverallPassed: bool(cResult.overall_passed),
	}

	numSteps := min(int(cResult.num_steps), len(result.Steps))

	for i := range numSteps {
		step := cResult.steps[i]
		result.Steps[i] = MEFStepResult{
			StepPct:          uint32(step.step_pct),
			OfferedRateKbps:  uint32(step.offered_rate_kbps),
			AchievedRateKbps: uint32(step.achieved_rate_kbps),
			FramesTx:         uint64(step.frames_tx),
			FramesRx:         uint64(step.frames_rx),
			FDUs:             float64(step.fd_us),
			FDMinUs:          float64(step.fd_min_us),
			FDMaxUs:          float64(step.fd_max_us),
			FDVUs:            float64(step.fdv_us),
			FLRPct:           float64(step.flr_pct),
			Passed:           bool(step.passed),
		}
	}

	return result, nil
}

// RunMEFPerfTest executes MEF performance test.
func (c *Context) RunMEFPerfTest(cfg *MEFConfig) (*MEFPerfResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.mef_config_t
	C.mef_default_config(&cCfg)
	fillMEFConfig(&cCfg, cfg)

	var cResult C.mef_perf_result_t
	ret := C.mef_perf_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("MEF performance test failed: %d", ret)
	}

	return &MEFPerfResult{
		ServiceID:       C.GoString(&cResult.service_id[0]),
		DurationSec:     uint32(cResult.duration_sec),
		FramesTx:        uint64(cResult.frames_tx),
		FramesRx:        uint64(cResult.frames_rx),
		ThroughputKbps:  uint32(cResult.throughput_kbps),
		FDMinUs:         float64(cResult.fd_min_us),
		FDAvgUs:         float64(cResult.fd_avg_us),
		FDMaxUs:         float64(cResult.fd_max_us),
		FDVUs:           float64(cResult.fdv_us),
		FLRPct:          float64(cResult.flr_pct),
		AvailabilityPct: float64(cResult.availability_pct),
		FDPassed:        bool(cResult.fd_passed),
		FDVPassed:       bool(cResult.fdv_passed),
		FLRPassed:       bool(cResult.flr_passed),
		AvailPassed:     bool(cResult.avail_passed),
		OverallPassed:   bool(cResult.overall_passed),
	}, nil
}

// RunMEFFullTest executes MEF configuration + performance tests.
func (c *Context) RunMEFFullTest(cfg *MEFConfig) (*MEFConfigResult, *MEFPerfResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.mef_config_t
	C.mef_default_config(&cCfg)
	fillMEFConfig(&cCfg, cfg)

	var cConfig C.mef_config_result_t
	var cPerf C.mef_perf_result_t
	ret := C.mef_full_test(c.ctx, &cCfg, &cConfig, &cPerf)
	if ret < 0 {
		return nil, nil, fmt.Errorf("MEF full test failed: %d", ret)
	}

	configResult := &MEFConfigResult{
		ServiceID:     C.GoString(&cConfig.service_id[0]),
		NumSteps:      uint32(cConfig.num_steps),
		OverallPassed: bool(cConfig.overall_passed),
	}
	for i := range len(configResult.Steps) {
		step := cConfig.steps[i]
		configResult.Steps[i] = MEFStepResult{
			StepPct:          uint32(step.step_pct),
			OfferedRateKbps:  uint32(step.offered_rate_kbps),
			AchievedRateKbps: uint32(step.achieved_rate_kbps),
			FramesTx:         uint64(step.frames_tx),
			FramesRx:         uint64(step.frames_rx),
			FDUs:             float64(step.fd_us),
			FDMinUs:          float64(step.fd_min_us),
			FDMaxUs:          float64(step.fd_max_us),
			FDVUs:            float64(step.fdv_us),
			FLRPct:           float64(step.flr_pct),
			Passed:           bool(step.passed),
		}
	}

	perfResult := &MEFPerfResult{
		ServiceID:       C.GoString(&cPerf.service_id[0]),
		DurationSec:     uint32(cPerf.duration_sec),
		FramesTx:        uint64(cPerf.frames_tx),
		FramesRx:        uint64(cPerf.frames_rx),
		ThroughputKbps:  uint32(cPerf.throughput_kbps),
		FDMinUs:         float64(cPerf.fd_min_us),
		FDAvgUs:         float64(cPerf.fd_avg_us),
		FDMaxUs:         float64(cPerf.fd_max_us),
		FDVUs:           float64(cPerf.fdv_us),
		FLRPct:          float64(cPerf.flr_pct),
		AvailabilityPct: float64(cPerf.availability_pct),
		FDPassed:        bool(cPerf.fd_passed),
		FDVPassed:       bool(cPerf.fdv_passed),
		FLRPassed:       bool(cPerf.flr_passed),
		AvailPassed:     bool(cPerf.avail_passed),
		OverallPassed:   bool(cPerf.overall_passed),
	}

	return configResult, perfResult, nil
}

func fillMEFConfig(cCfg *C.mef_config_t, cfg *MEFConfig) {
	if cfg == nil {
		return
	}
	fillMEFServiceID(cCfg, cfg.ServiceID)
	if cfg.CoS > 0 {
		cCfg.cos = C.mef_cos_t(cfg.CoS)
	}
	if cfg.CIRMbps > 0 {
		cCfg.bw_profile.cir_kbps = C.uint32_t(cfg.CIRMbps * kilobitsPerMegabit)
	}
	if cfg.EIRMbps > 0 {
		cCfg.bw_profile.eir_kbps = C.uint32_t(cfg.EIRMbps * kilobitsPerMegabit)
	}
	if cfg.CBSBytes > 0 {
		cCfg.bw_profile.cbs_bytes = C.uint32_t(cfg.CBSBytes)
	}
	if cfg.EBSBytes > 0 {
		cCfg.bw_profile.ebs_bytes = C.uint32_t(cfg.EBSBytes)
	}
	if cfg.FDThresholdUs > 0 {
		cCfg.sla.fd_threshold_us = C.double(cfg.FDThresholdUs)
	}
	if cfg.FDVThresholdUs > 0 {
		cCfg.sla.fdv_threshold_us = C.double(cfg.FDVThresholdUs)
	}
	if cfg.FLRThresholdPct > 0 {
		cCfg.sla.flr_threshold_pct = C.double(cfg.FLRThresholdPct)
	}
	if cfg.AvailabilityPct > 0 {
		cCfg.sla.availability_pct = C.double(cfg.AvailabilityPct)
	}
	if cfg.ConfigDurationSec > 0 {
		cCfg.config_test_duration_sec = C.uint32_t(cfg.ConfigDurationSec)
	}
	if cfg.PerfDurationMin > 0 {
		cCfg.perf_test_duration_min = C.uint32_t(cfg.PerfDurationMin)
	}
	fillMEFFrameSizes(cCfg, cfg.FrameSizes)
}

func fillMEFServiceID(cCfg *C.mef_config_t, serviceID string) {
	if serviceID == "" {
		return
	}
	idBytes := []byte(serviceID)
	for i := range min(len(idBytes), mefServiceIDMaxLength) {
		cCfg.service_id[i] = C.char(idBytes[i])
	}
	cCfg.service_id[mefServiceIDMaxLength] = 0
}

func fillMEFFrameSizes(cCfg *C.mef_config_t, frameSizes []uint32) {
	count := min(len(frameSizes), mefMaxFrameSizes)
	for i := range count {
		cCfg.frame_sizes[i] = C.uint32_t(frameSizes[i])
	}
	cCfg.num_frame_sizes = C.uint32_t(count)
}
