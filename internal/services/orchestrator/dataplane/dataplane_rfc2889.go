//go:build cgo && linux

package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../../build -lreflector -lpthread -lm
#cgo linux LDFLAGS: -lxdp -lbpf

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

typedef struct rfc2544_ctx rfc2544_ctx_t;

#define RFC2889_MAX_PORTS 64

typedef enum {
    RFC2889_FORWARDING_RATE = 0,
    RFC2889_ADDRESS_CACHING = 1,
    RFC2889_ADDRESS_LEARNING = 2,
    RFC2889_BROADCAST_FORWARDING = 3,
    RFC2889_BROADCAST_LATENCY = 4,
    RFC2889_CONGESTION_CONTROL = 5,
    RFC2889_FORWARD_PRESSURE = 6,
    RFC2889_ERROR_FILTERING = 7,
    RFC2889_TEST_COUNT = 8
} rfc2889_test_type_t;

typedef enum {
    TRAFFIC_FULLY_MESHED = 0,
    TRAFFIC_PARTIALLY_MESHED = 1,
    TRAFFIC_PAIR_WISE = 2,
    TRAFFIC_ONE_TO_MANY = 3,
    TRAFFIC_MANY_TO_ONE = 4
} traffic_pattern_t;

typedef struct {
    uint32_t frame_size;
    uint32_t port_count;
    traffic_pattern_t pattern;
    double max_rate_pct;
    double max_rate_fps;
    double aggregate_rate_mbps;
    uint64_t frames_tx;
    uint64_t frames_rx;
} rfc2889_fwd_result_t;

typedef struct {
    uint32_t address_count;
    uint32_t frame_size;
    uint32_t port_count;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double loss_pct;
    bool passed;
} rfc2889_cache_result_t;

typedef struct {
    uint32_t frame_size;
    uint32_t port_count;
    double learning_rate_fps;
    uint32_t addresses_learned;
    double learning_time_ms;
    uint32_t verification_frames;
    double verification_loss_pct;
} rfc2889_learning_result_t;

typedef struct {
    uint32_t frame_size;
    uint32_t ingress_ports;
    uint32_t egress_ports;
    double broadcast_rate_fps;
    double broadcast_rate_mbps;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double replication_factor;
} rfc2889_broadcast_result_t;

typedef struct {
    uint32_t frame_size;
    double overload_rate_pct;
    uint64_t frames_tx;
    uint64_t frames_rx;
    uint64_t frames_dropped;
    double head_of_line_blocking;
    bool backpressure_observed;
    uint64_t pause_frames_rx;
} rfc2889_congestion_result_t;

typedef struct {
    char interface[64];
    uint8_t mac_base[6];
    uint32_t mac_count;
    bool is_ingress;
    bool is_egress;
} rfc2889_port_t;

typedef struct {
    rfc2889_test_type_t test_type;
    traffic_pattern_t pattern;
    uint32_t port_count;
    rfc2889_port_t ports[RFC2889_MAX_PORTS];
    uint32_t frame_size;
    uint32_t trial_duration_sec;
    uint32_t warmup_sec;
    uint32_t address_count;
    double acceptable_loss_pct;
} rfc2889_config_t;

extern int rfc2889_forwarding_test(rfc2544_ctx_t *ctx, const rfc2889_config_t *config,
                                 rfc2889_fwd_result_t *result);
extern int rfc2889_caching_test(rfc2544_ctx_t *ctx, const rfc2889_config_t *config,
                              rfc2889_cache_result_t *result);
extern int rfc2889_learning_test(rfc2544_ctx_t *ctx, const rfc2889_config_t *config,
                               rfc2889_learning_result_t *result);
extern int rfc2889_broadcast_test(rfc2544_ctx_t *ctx, const rfc2889_config_t *config,
                                rfc2889_broadcast_result_t *result);
extern int rfc2889_congestion_test(rfc2544_ctx_t *ctx, const rfc2889_config_t *config,
                                 rfc2889_congestion_result_t *result);
extern void rfc2889_default_config(rfc2889_config_t *config);
*/
import "C"
import "fmt"

// RunRFC2889ForwardingTest executes RFC 2889 forwarding rate test.
func (c *Context) RunRFC2889ForwardingTest(cfg *RFC2889Config) (*RFC2889ForwardingResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc2889_config_t
	C.rfc2889_default_config(&cCfg)
	fillRFC2889Config(&cCfg, cfg)

	var cResult C.rfc2889_fwd_result_t
	ret := C.rfc2889_forwarding_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 2889 forwarding test failed: %d", ret)
	}

	return &RFC2889ForwardingResult{
		FrameSize:         uint32(cResult.frame_size),
		PortCount:         uint32(cResult.port_count),
		Pattern:           uint32(cResult.pattern),
		MaxRatePct:        float64(cResult.max_rate_pct),
		MaxRateFps:        float64(cResult.max_rate_fps),
		AggregateRateMbps: float64(cResult.aggregate_rate_mbps),
		FramesTx:          uint64(cResult.frames_tx),
		FramesRx:          uint64(cResult.frames_rx),
	}, nil
}

// RunRFC2889CachingTest executes RFC 2889 address caching test.
func (c *Context) RunRFC2889CachingTest(cfg *RFC2889Config) (*RFC2889CachingResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc2889_config_t
	C.rfc2889_default_config(&cCfg)
	fillRFC2889Config(&cCfg, cfg)

	var cResult C.rfc2889_cache_result_t
	ret := C.rfc2889_caching_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 2889 caching test failed: %d", ret)
	}

	return &RFC2889CachingResult{
		AddressCount: uint32(cResult.address_count),
		FrameSize:    uint32(cResult.frame_size),
		PortCount:    uint32(cResult.port_count),
		FramesTx:     uint64(cResult.frames_tx),
		FramesRx:     uint64(cResult.frames_rx),
		LossPct:      float64(cResult.loss_pct),
		Passed:       bool(cResult.passed),
	}, nil
}

// RunRFC2889LearningTest executes RFC 2889 address learning test.
func (c *Context) RunRFC2889LearningTest(cfg *RFC2889Config) (*RFC2889LearningResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc2889_config_t
	C.rfc2889_default_config(&cCfg)
	fillRFC2889Config(&cCfg, cfg)

	var cResult C.rfc2889_learning_result_t
	ret := C.rfc2889_learning_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 2889 learning test failed: %d", ret)
	}

	return &RFC2889LearningResult{
		FrameSize:           uint32(cResult.frame_size),
		PortCount:           uint32(cResult.port_count),
		LearningRateFps:     float64(cResult.learning_rate_fps),
		AddressesLearned:    uint32(cResult.addresses_learned),
		LearningTimeMs:      float64(cResult.learning_time_ms),
		VerificationFrames:  uint32(cResult.verification_frames),
		VerificationLossPct: float64(cResult.verification_loss_pct),
	}, nil
}

// RunRFC2889BroadcastTest executes RFC 2889 broadcast forwarding test.
func (c *Context) RunRFC2889BroadcastTest(cfg *RFC2889Config) (*RFC2889BroadcastResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc2889_config_t
	C.rfc2889_default_config(&cCfg)
	fillRFC2889Config(&cCfg, cfg)

	var cResult C.rfc2889_broadcast_result_t
	ret := C.rfc2889_broadcast_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 2889 broadcast test failed: %d", ret)
	}

	return &RFC2889BroadcastResult{
		FrameSize:         uint32(cResult.frame_size),
		IngressPorts:      uint32(cResult.ingress_ports),
		EgressPorts:       uint32(cResult.egress_ports),
		BroadcastRateFps:  float64(cResult.broadcast_rate_fps),
		BroadcastRateMbps: float64(cResult.broadcast_rate_mbps),
		FramesTx:          uint64(cResult.frames_tx),
		FramesRx:          uint64(cResult.frames_rx),
		ReplicationFactor: float64(cResult.replication_factor),
	}, nil
}

// RunRFC2889CongestionTest executes RFC 2889 congestion control test.
func (c *Context) RunRFC2889CongestionTest(cfg *RFC2889Config) (*RFC2889CongestionResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc2889_config_t
	C.rfc2889_default_config(&cCfg)
	fillRFC2889Config(&cCfg, cfg)

	var cResult C.rfc2889_congestion_result_t
	ret := C.rfc2889_congestion_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 2889 congestion test failed: %d", ret)
	}

	return &RFC2889CongestionResult{
		FrameSize:            uint32(cResult.frame_size),
		OverloadRatePct:      float64(cResult.overload_rate_pct),
		FramesTx:             uint64(cResult.frames_tx),
		FramesRx:             uint64(cResult.frames_rx),
		FramesDropped:        uint64(cResult.frames_dropped),
		HeadOfLineBlocking:   float64(cResult.head_of_line_blocking),
		BackpressureObserved: bool(cResult.backpressure_observed),
		PauseFramesRx:        uint64(cResult.pause_frames_rx),
	}, nil
}

func fillRFC2889Config(cCfg *C.rfc2889_config_t, cfg *RFC2889Config) {
	if cfg == nil {
		return
	}
	if cfg.FrameSize > 0 {
		cCfg.frame_size = C.uint32_t(cfg.FrameSize)
	}
	if cfg.DurationSec > 0 {
		cCfg.trial_duration_sec = C.uint32_t(cfg.DurationSec)
	}
	if cfg.WarmupSec > 0 {
		cCfg.warmup_sec = C.uint32_t(cfg.WarmupSec)
	}
	if cfg.AddressCount > 0 {
		cCfg.address_count = C.uint32_t(cfg.AddressCount)
	}
	if cfg.AcceptableLossPct > 0 {
		cCfg.acceptable_loss_pct = C.double(cfg.AcceptableLossPct)
	}
	if cfg.PortCount > 0 {
		cCfg.port_count = C.uint32_t(cfg.PortCount)
	}
	if cfg.Pattern > 0 {
		cCfg.pattern = C.traffic_pattern_t(cfg.Pattern)
	}
}
