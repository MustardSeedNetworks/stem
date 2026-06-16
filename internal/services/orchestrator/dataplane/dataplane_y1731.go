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
    Y1731_CCM = 1,
    Y1731_LBR = 2,
    Y1731_LBM = 3,
    Y1731_LTR = 4,
    Y1731_LTM = 5,
    Y1731_AIS = 33,
    Y1731_LCK = 35,
    Y1731_TST = 37,
    Y1731_APS = 39,
    Y1731_MCC = 41,
    Y1731_LMR = 42,
    Y1731_LMM = 43,
    Y1731_1DM = 45,
    Y1731_DMR = 46,
    Y1731_DMM = 47,
    Y1731_EXR = 48,
    Y1731_EXM = 49,
    Y1731_VSR = 50,
    Y1731_VSM = 51,
    Y1731_SLR = 54,
    Y1731_SLM = 55
} y1731_opcode_t;

typedef enum {
    MEG_LEVEL_CUSTOMER = 0,
    MEG_LEVEL_1 = 1,
    MEG_LEVEL_2 = 2,
    MEG_LEVEL_PROVIDER = 3,
    MEG_LEVEL_4 = 4,
    MEG_LEVEL_5 = 5,
    MEG_LEVEL_6 = 6,
    MEG_LEVEL_OPERATOR = 7
} meg_level_t;

typedef enum {
    CCM_INVALID = 0,
    CCM_3_33MS = 1,
    CCM_10MS = 2,
    CCM_100MS = 3,
    CCM_1S = 4,
    CCM_10S = 5,
    CCM_1MIN = 6,
    CCM_10MIN = 7
} ccm_interval_t;

typedef struct {
    uint32_t frames_sent;
    uint32_t frames_received;
    uint32_t frames_lost;
    double delay_min_us;
    double delay_avg_us;
    double delay_max_us;
    double delay_variation_us;
} y1731_delay_result_t;

typedef struct {
    uint64_t frames_tx;
    uint64_t frames_rx;
    uint64_t near_end_loss;
    uint64_t far_end_loss;
    double near_end_loss_ratio;
    double far_end_loss_ratio;
    double availability_pct;
} y1731_loss_result_t;

typedef struct {
    uint64_t lbm_sent;
    uint64_t lbr_received;
    double rtt_min_ms;
    double rtt_avg_ms;
    double rtt_max_ms;
} y1731_loopback_result_t;

typedef struct {
    uint32_t mep_id;
    meg_level_t meg_level;
    char meg_id[32];
    ccm_interval_t ccm_interval;
    uint8_t priority;
    bool enabled;
} y1731_mep_config_t;

typedef enum {
    Y1731_STATE_INIT = 0,
    Y1731_STATE_RUNNING = 1,
    Y1731_STATE_STOPPED = 2,
    Y1731_STATE_ERROR = 3
} y1731_state_t;

typedef struct {
    y1731_mep_config_t local_mep;
    y1731_mep_config_t remote_mep;
    y1731_state_t state;
    uint64_t ccm_tx_count;
    uint64_t ccm_rx_count;
    bool rdi_received;
    uint64_t last_ccm_time;
} y1731_session_t;

extern int y1731_delay_measurement(rfc2544_ctx_t *ctx, y1731_session_t *session, uint32_t count,
                                 uint32_t interval_ms, y1731_delay_result_t *result);
extern int y1731_loss_measurement(rfc2544_ctx_t *ctx, y1731_session_t *session, uint32_t duration_sec,
                                y1731_loss_result_t *result);
extern int y1731_synthetic_loss(rfc2544_ctx_t *ctx, y1731_session_t *session, uint32_t count,
                              uint32_t interval_ms, y1731_loss_result_t *result);
extern int y1731_loopback(rfc2544_ctx_t *ctx, y1731_session_t *session, const uint8_t *target_mac,
                        uint32_t count, y1731_loopback_result_t *result);
extern int y1731_session_init(rfc2544_ctx_t *ctx, const y1731_mep_config_t *config,
                            y1731_session_t *session);
extern void y1731_default_mep_config(y1731_mep_config_t *config);
*/
import "C"
import "fmt"

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

	return &Y1731DelayResult{
		FramesSent:       uint32(cResult.frames_sent),
		FramesReceived:   uint32(cResult.frames_received),
		FramesLost:       uint32(cResult.frames_lost),
		DelayMinUs:       float64(cResult.delay_min_us),
		DelayAvgUs:       float64(cResult.delay_avg_us),
		DelayMaxUs:       float64(cResult.delay_max_us),
		DelayVariationUs: float64(cResult.delay_variation_us),
	}, nil
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

	return &Y1731LossResult{
		FramesTx:         uint64(cResult.frames_tx),
		FramesRx:         uint64(cResult.frames_rx),
		NearEndLoss:      uint64(cResult.near_end_loss),
		FarEndLoss:       uint64(cResult.far_end_loss),
		NearEndLossRatio: float64(cResult.near_end_loss_ratio),
		FarEndLossRatio:  float64(cResult.far_end_loss_ratio),
		AvailabilityPct:  float64(cResult.availability_pct),
	}, nil
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

	return &Y1731LossResult{
		FramesTx:         uint64(cResult.frames_tx),
		FramesRx:         uint64(cResult.frames_rx),
		NearEndLoss:      uint64(cResult.near_end_loss),
		FarEndLoss:       uint64(cResult.far_end_loss),
		NearEndLossRatio: float64(cResult.near_end_loss_ratio),
		FarEndLossRatio:  float64(cResult.far_end_loss_ratio),
		AvailabilityPct:  float64(cResult.availability_pct),
	}, nil
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

	if cfg != nil {
		if cfg.MEPID > 0 {
			mep.mep_id = C.uint32_t(cfg.MEPID)
		}
		if cfg.MEGLevel > 0 {
			mep.meg_level = C.meg_level_t(cfg.MEGLevel)
		}
		if cfg.MEGID != "" {
			megBytes := []byte(cfg.MEGID)
			for i := 0; i < len(megBytes) && i < 31; i++ {
				mep.meg_id[i] = C.char(megBytes[i])
			}
			mep.meg_id[31] = 0
		}
		if cfg.CCMInterval > 0 {
			mep.ccm_interval = C.ccm_interval_t(cfg.CCMInterval)
		}
		if cfg.Priority > 0 {
			mep.priority = C.uint8_t(cfg.Priority)
		}
		mep.enabled = C.bool(true)
	}

	var session C.y1731_session_t
	ret := C.y1731_session_init(c.ctx, &mep, &session)
	if ret < 0 {
		return session, fmt.Errorf("Y.1731 session init failed: %d", ret)
	}
	return session, nil
}

func y1731CountInterval(cfg *Y1731Config) (uint32, uint32) {
	count := uint32(10)
	interval := uint32(1000)
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
	count := uint32(10)
	if cfg != nil && cfg.Count > 0 {
		count = cfg.Count
	}
	return count
}

func y1731Duration(cfg *Y1731Config) uint32 {
	duration := uint32(60)
	if cfg != nil && cfg.DurationSec > 0 {
		duration = cfg.DurationSec
	}
	return duration
}
