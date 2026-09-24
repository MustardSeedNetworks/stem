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
	defaultY1731Count      = 10
	defaultY1731IntervalMs = 1000
	defaultY1731Duration   = 60
	y1731MEGIDMaxLength    = 31
)

// RunY1731DelayTest executes Y.1731 delay measurement.
func (c *Context) RunY1731DelayTest(cfg *Y1731Config) (*Y1731DelayResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := c.newY1731Session(cfg)
	if err != nil {
		return nil, err
	}

	count, interval := y1731CountInterval(cfg)

	var cResult C.y1731_delay_result_t
	ret := C.y1731_delay_measurement(c.ctx, &session, C.uint32_t(count), C.uint32_t(interval), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1731 delay test failed: %d", ret)
	}

	return y1731DelayResult(&cResult), nil
}

func y1731DelayResult(result *C.y1731_delay_result_t) *Y1731DelayResult {
	return &Y1731DelayResult{
		FramesSent:       uint32(result.frames_sent),
		FramesReceived:   uint32(result.frames_received),
		FramesLost:       uint32(result.frames_lost),
		DelayMinUs:       float64(result.delay_min_us),
		DelayAvgUs:       float64(result.delay_avg_us),
		DelayMaxUs:       float64(result.delay_max_us),
		DelayVariationUs: float64(result.delay_variation_us),
	}
}

// RunY1731LossTest executes Y.1731 loss measurement.
func (c *Context) RunY1731LossTest(cfg *Y1731Config) (*Y1731LossResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := c.newY1731Session(cfg)
	if err != nil {
		return nil, err
	}

	duration := y1731Duration(cfg)

	var cResult C.y1731_loss_result_t
	ret := C.y1731_loss_measurement(c.ctx, &session, C.uint32_t(duration), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1731 loss test failed: %d", ret)
	}

	return y1731LossResult(&cResult), nil
}

// RunY1731SyntheticLossTest executes Y.1731 synthetic loss measurement.
func (c *Context) RunY1731SyntheticLossTest(cfg *Y1731Config) (*Y1731LossResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := c.newY1731Session(cfg)
	if err != nil {
		return nil, err
	}

	count, interval := y1731CountInterval(cfg)

	var cResult C.y1731_loss_result_t
	ret := C.y1731_synthetic_loss(c.ctx, &session, C.uint32_t(count), C.uint32_t(interval), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1731 synthetic loss test failed: %d", ret)
	}

	return y1731LossResult(&cResult), nil
}

func y1731LossResult(result *C.y1731_loss_result_t) *Y1731LossResult {
	return &Y1731LossResult{
		FramesTx:         uint64(result.frames_tx),
		FramesRx:         uint64(result.frames_rx),
		NearEndLoss:      uint64(result.near_end_loss),
		FarEndLoss:       uint64(result.far_end_loss),
		NearEndLossRatio: float64(result.near_end_loss_ratio),
		FarEndLossRatio:  float64(result.far_end_loss_ratio),
		AvailabilityPct:  float64(result.availability_pct),
	}
}

// RunY1731LoopbackTest executes Y.1731 loopback test.
func (c *Context) RunY1731LoopbackTest(cfg *Y1731Config) (*Y1731LoopbackResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := c.newY1731Session(cfg)
	if err != nil {
		return nil, err
	}

	count := y1731Count(cfg)

	var cResult C.y1731_loopback_result_t
	ret := C.y1731_loopback(c.ctx, &session, nil, C.uint32_t(count), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("Y.1731 loopback test failed: %d", ret)
	}

	return &Y1731LoopbackResult{
		LBMSent:     uint64(cResult.lbm_sent),
		LBRReceived: uint64(cResult.lbr_received),
		RTTMinMs:    float64(cResult.rtt_min_ms),
		RTTAvgMs:    float64(cResult.rtt_avg_ms),
		RTTMaxMs:    float64(cResult.rtt_max_ms),
	}, nil
}

func (c *Context) newY1731Session(cfg *Y1731Config) (C.y1731_session_t, error) {
	var mep C.y1731_mep_config_t
	C.y1731_default_mep_config(&mep)
	fillY1731MEPConfig(&mep, cfg)

	var session C.y1731_session_t
	ret := C.y1731_session_init(c.ctx, &mep, &session)
	if ret < 0 {
		return session, fmt.Errorf("Y.1731 session init failed: %d", ret)
	}
	return session, nil
}

func fillY1731MEPConfig(mep *C.y1731_mep_config_t, cfg *Y1731Config) {
	if cfg == nil {
		return
	}
	if cfg.MEPID > 0 {
		mep.mep_id = C.uint32_t(cfg.MEPID)
	}
	if cfg.MEGLevel > 0 {
		mep.meg_level = C.meg_level_t(cfg.MEGLevel)
	}
	if cfg.MEGID != "" {
		megBytes := []byte(cfg.MEGID)
		for i := range min(len(megBytes), y1731MEGIDMaxLength) {
			mep.meg_id[i] = C.char(megBytes[i])
		}
		mep.meg_id[y1731MEGIDMaxLength] = 0
	}
	if cfg.CCMInterval > 0 {
		mep.ccm_interval = C.ccm_interval_t(cfg.CCMInterval)
	}
	if cfg.Priority > 0 {
		mep.priority = C.uint8_t(cfg.Priority)
	}
	mep.enabled = C.bool(true)
}

func y1731CountInterval(cfg *Y1731Config) (uint32, uint32) {
	count := uint32(defaultY1731Count)
	interval := uint32(defaultY1731IntervalMs)
	if cfg != nil {
		if cfg.Count > 0 {
			count = cfg.Count
		}
		if cfg.IntervalMs > 0 {
			interval = cfg.IntervalMs
		}
	}
	return count, interval
}

func y1731Count(cfg *Y1731Config) uint32 {
	count := uint32(defaultY1731Count)
	if cfg != nil && cfg.Count > 0 {
		count = cfg.Count
	}
	return count
}

func y1731Duration(cfg *Y1731Config) uint32 {
	duration := uint32(defaultY1731Duration)
	if cfg != nil && cfg.DurationSec > 0 {
		duration = cfg.DurationSec
	}
	return duration
}
