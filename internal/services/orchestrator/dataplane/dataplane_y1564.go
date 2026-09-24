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

// RunY1564ConfigTest executes ITU-T Y.1564 Service Configuration Test.
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

	for i := range len(result.Steps) {
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

// RunY1564PerfTest executes ITU-T Y.1564 Service Performance Test.
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
