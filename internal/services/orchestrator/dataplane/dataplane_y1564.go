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

typedef struct {
    double cir_mbps;
    double eir_mbps;
    uint32_t cbs_bytes;
    uint32_t ebs_bytes;
    double fd_threshold_ms;
    double fdv_threshold_ms;
    double flr_threshold_pct;
} y1564_sla_t;

typedef struct {
    uint32_t service_id;
    char service_name[32];
    y1564_sla_t sla;
    uint32_t frame_size;
    uint8_t cos;
    bool enabled;
} y1564_service_t;

typedef struct {
    uint32_t step;
    double offered_rate_pct;
    double achieved_rate_mbps;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double flr_pct;
    double fd_avg_ms;
    double fd_min_ms;
    double fd_max_ms;
    double fdv_ms;
    bool flr_pass;
    bool fd_pass;
    bool fdv_pass;
    bool step_pass;
} y1564_step_result_t;

typedef struct {
    uint32_t service_id;
    y1564_step_result_t steps[4];
    bool service_pass;
} y1564_config_result_t;

typedef struct {
    uint32_t service_id;
    uint32_t duration_sec;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double flr_pct;
    double fd_avg_ms;
    double fd_min_ms;
    double fd_max_ms;
    double fdv_ms;
    bool flr_pass;
    bool fd_pass;
    bool fdv_pass;
    bool service_pass;
} y1564_perf_result_t;

extern int y1564_config_test(rfc2544_ctx_t *ctx, const y1564_service_t *service,
                             y1564_config_result_t *result);
extern int y1564_perf_test(rfc2544_ctx_t *ctx, const y1564_service_t *service,
                           uint32_t duration_sec, y1564_perf_result_t *result);
*/
import "C"
import "fmt"

// RunY1564ConfigTest executes ITU-T Y.1564 Service Configuration Test
func (c *Context) RunY1564ConfigTest(service *Y1564Service) (*Y1564ConfigResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Convert Go service to C service
	var cService C.y1564_service_t
	cService.service_id = C.uint32_t(service.ServiceID)
	cService.sla.cir_mbps = C.double(service.SLA.CIRMbps)
	cService.sla.eir_mbps = C.double(service.SLA.EIRMbps)
	cService.sla.cbs_bytes = C.uint32_t(service.SLA.CBSBytes)
	cService.sla.ebs_bytes = C.uint32_t(service.SLA.EBSBytes)
	cService.sla.fd_threshold_ms = C.double(service.SLA.FDThresholdMs)
	cService.sla.fdv_threshold_ms = C.double(service.SLA.FDVThresholdMs)
	cService.sla.flr_threshold_pct = C.double(service.SLA.FLRThresholdPct)
	cService.frame_size = C.uint32_t(service.FrameSize)
	cService.cos = C.uint8_t(service.CoS)
	cService.enabled = C.bool(service.Enabled)

	// Copy service name (ensure null-termination)
	nameBytes := []byte(service.ServiceName)
	for i := 0; i < len(nameBytes) && i < 31; i++ {
		cService.service_name[i] = C.char(nameBytes[i])
	}
	cService.service_name[31] = 0 // Ensure null-termination

	var cResult C.y1564_config_result_t
	ret := C.y1564_config_test(c.ctx, &cService, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1564 config test failed: %d", ret)
	}

	result := &Y1564ConfigResult{
		ServiceID:   uint32(cResult.service_id),
		ServicePass: bool(cResult.service_pass),
	}

	for i := 0; i < 4; i++ {
		result.Steps[i] = Y1564StepResult{
			Step:             uint32(cResult.steps[i].step),
			OfferedRatePct:   float64(cResult.steps[i].offered_rate_pct),
			AchievedRateMbps: float64(cResult.steps[i].achieved_rate_mbps),
			FramesTx:         uint64(cResult.steps[i].frames_tx),
			FramesRx:         uint64(cResult.steps[i].frames_rx),
			FLRPct:           float64(cResult.steps[i].flr_pct),
			FDAvgMs:          float64(cResult.steps[i].fd_avg_ms),
			FDMinMs:          float64(cResult.steps[i].fd_min_ms),
			FDMaxMs:          float64(cResult.steps[i].fd_max_ms),
			FDVMs:            float64(cResult.steps[i].fdv_ms),
			FLRPass:          bool(cResult.steps[i].flr_pass),
			FDPass:           bool(cResult.steps[i].fd_pass),
			FDVPass:          bool(cResult.steps[i].fdv_pass),
			StepPass:         bool(cResult.steps[i].step_pass),
		}
	}

	return result, nil
}

// RunY1564PerfTest executes ITU-T Y.1564 Service Performance Test
func (c *Context) RunY1564PerfTest(service *Y1564Service, durationSec uint32) (*Y1564PerfResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Convert Go service to C service
	var cService C.y1564_service_t
	cService.service_id = C.uint32_t(service.ServiceID)
	cService.sla.cir_mbps = C.double(service.SLA.CIRMbps)
	cService.sla.eir_mbps = C.double(service.SLA.EIRMbps)
	cService.sla.cbs_bytes = C.uint32_t(service.SLA.CBSBytes)
	cService.sla.ebs_bytes = C.uint32_t(service.SLA.EBSBytes)
	cService.sla.fd_threshold_ms = C.double(service.SLA.FDThresholdMs)
	cService.sla.fdv_threshold_ms = C.double(service.SLA.FDVThresholdMs)
	cService.sla.flr_threshold_pct = C.double(service.SLA.FLRThresholdPct)
	cService.frame_size = C.uint32_t(service.FrameSize)
	cService.cos = C.uint8_t(service.CoS)
	cService.enabled = C.bool(service.Enabled)

	// Copy service name (ensure null-termination)
	nameBytes := []byte(service.ServiceName)
	for i := 0; i < len(nameBytes) && i < 31; i++ {
		cService.service_name[i] = C.char(nameBytes[i])
	}
	cService.service_name[31] = 0 // Ensure null-termination

	var cResult C.y1564_perf_result_t
	ret := C.y1564_perf_test(c.ctx, &cService, C.uint32_t(durationSec), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1564 perf test failed: %d", ret)
	}

	return &Y1564PerfResult{
		ServiceID:   uint32(cResult.service_id),
		DurationSec: uint32(cResult.duration_sec),
		FramesTx:    uint64(cResult.frames_tx),
		FramesRx:    uint64(cResult.frames_rx),
		FLRPct:      float64(cResult.flr_pct),
		FDAvgMs:     float64(cResult.fd_avg_ms),
		FDMinMs:     float64(cResult.fd_min_ms),
		FDMaxMs:     float64(cResult.fd_max_ms),
		FDVMs:       float64(cResult.fdv_ms),
		FLRPass:     bool(cResult.flr_pass),
		FDPass:      bool(cResult.fd_pass),
		FDVPass:     bool(cResult.fdv_pass),
		ServicePass: bool(cResult.service_pass),
	}, nil
}
