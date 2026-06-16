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

typedef struct rfc2544_ctx rfc2544_ctx_t;

typedef enum {
    MEF_EPL = 0,
    MEF_EVPL = 1,
    MEF_EP_LAN = 2,
    MEF_EVP_LAN = 3,
    MEF_EP_TREE = 4,
    MEF_EVP_TREE = 5
} mef_service_type_t;

typedef enum {
    MEF_COS_BEST_EFFORT = 0,
    MEF_COS_LOW = 1,
    MEF_COS_MEDIUM = 2,
    MEF_COS_HIGH = 3,
    MEF_COS_CRITICAL = 4,
    MEF_COS_HIGH_PRIORITY = 3
} mef_cos_t;

typedef enum {
    MEF_TIER_STANDARD = 0,
    MEF_TIER_PREMIUM = 1,
    MEF_TIER_MISSION_CRITICAL = 2
} mef_perf_tier_t;

typedef struct {
    double fd_threshold_us;
    double fdv_threshold_us;
    double flr_threshold_pct;
    double availability_pct;
    uint32_t mttr_minutes;
    uint32_t mtbf_hours;
} mef_sla_t;

typedef struct {
    uint32_t step_pct;
    uint32_t offered_rate_kbps;
    uint32_t achieved_rate_kbps;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double fd_us;
    double fd_min_us;
    double fd_max_us;
    double fdv_us;
    double flr_pct;
    bool passed;
} mef_step_result_t;

typedef struct {
    uint32_t cir_kbps;
    uint32_t cbs_bytes;
    uint32_t eir_kbps;
    uint32_t ebs_bytes;
    bool color_mode;
    bool coupling_flag;
} mef_bandwidth_profile_t;

typedef struct {
    mef_service_type_t service_type;
    mef_cos_t cos;
    char service_id[32];
    mef_bandwidth_profile_t bw_profile;
    mef_sla_t sla;
    uint32_t config_test_duration_sec;
    uint32_t perf_test_duration_min;
    uint32_t frame_sizes[7];
    uint32_t num_frame_sizes;
} mef_config_t;

typedef struct {
    char service_id[32];
    mef_step_result_t steps[4];
    uint32_t num_steps;
    bool overall_passed;
} mef_config_result_t;

typedef struct {
    char service_id[32];
    uint32_t duration_sec;
    uint64_t frames_tx;
    uint64_t frames_rx;
    uint32_t throughput_kbps;
    double fd_min_us;
    double fd_avg_us;
    double fd_max_us;
    double fdv_us;
    double flr_pct;
    double availability_pct;
    bool fd_passed;
    bool fdv_passed;
    bool flr_passed;
    bool avail_passed;
    bool overall_passed;
} mef_perf_result_t;

extern int mef_config_test(rfc2544_ctx_t *ctx, const mef_config_t *config, mef_config_result_t *result);
extern int mef_perf_test(rfc2544_ctx_t *ctx, const mef_config_t *config, mef_perf_result_t *result);
extern int mef_full_test(rfc2544_ctx_t *ctx, const mef_config_t *config,
                        mef_config_result_t *config_result, mef_perf_result_t *perf_result);
extern void mef_default_config(mef_config_t *config);
*/
import "C"
import "fmt"

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

	numSteps := int(cResult.num_steps)
	if numSteps > len(result.Steps) {
		numSteps = len(result.Steps)
	}

	for i := 0; i < numSteps; i++ {
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
	for i := 0; i < len(configResult.Steps); i++ {
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
	if cfg.ServiceID != "" {
		idBytes := []byte(cfg.ServiceID)
		for i := 0; i < len(idBytes) && i < 31; i++ {
			cCfg.service_id[i] = C.char(idBytes[i])
		}
		cCfg.service_id[31] = 0
	}
	if cfg.CoS > 0 {
		cCfg.cos = C.mef_cos_t(cfg.CoS)
	}
	if cfg.CIRMbps > 0 {
		cCfg.bw_profile.cir_kbps = C.uint32_t(cfg.CIRMbps * 1000)
	}
	if cfg.EIRMbps > 0 {
		cCfg.bw_profile.eir_kbps = C.uint32_t(cfg.EIRMbps * 1000)
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
	if len(cfg.FrameSizes) > 0 {
		count := len(cfg.FrameSizes)
		if count > 7 {
			count = 7
		}
		for i := 0; i < count; i++ {
			cCfg.frame_sizes[i] = C.uint32_t(cfg.FrameSizes[i])
		}
		cCfg.num_frame_sizes = C.uint32_t(count)
	}
}
