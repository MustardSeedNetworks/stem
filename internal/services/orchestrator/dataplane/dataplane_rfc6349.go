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

typedef enum {
    TCP_THROUGHPUT = 0,
    TCP_SINGLE_STREAM = 0,
    TCP_MULTI_STREAM = 1,
    TCP_BIDIRECTIONAL = 2
} tcp_test_mode_t;

typedef struct {
    double achieved_rate_mbps;
    double theoretical_rate_mbps;
    double rtt_min_ms;
    double rtt_avg_ms;
    double rtt_max_ms;
    uint64_t bdp_bytes;
    uint32_t rwnd_used;
    uint64_t bytes_transferred;
    uint64_t retransmissions;
    uint32_t test_duration_ms;
    double tcp_efficiency;
    double buffer_delay_pct;
    double transfer_time_ratio;
    bool passed;
} rfc6349_result_t;

typedef struct {
    uint32_t path_mtu;
    uint32_t mss;
    double rtt_min_ms;
    double rtt_avg_ms;
    double rtt_max_ms;
    uint64_t bdp_bytes;
    uint32_t ideal_rwnd;
    double bottleneck_bw_mbps;
} tcp_path_info_t;

typedef struct {
    double target_rate_mbps;
    double min_rtt_ms;
    double max_rtt_ms;
    uint32_t rwnd_size;
    uint32_t test_duration_sec;
    uint32_t parallel_streams;
    uint32_t mss;
    tcp_test_mode_t mode;
} rfc6349_config_t;

extern int rfc6349_path_test(rfc2544_ctx_t *ctx, const rfc6349_config_t *config, tcp_path_info_t *path);
extern int rfc6349_throughput_test(rfc2544_ctx_t *ctx, const rfc6349_config_t *config,
                                 rfc6349_result_t *result);
extern void rfc6349_default_config(rfc6349_config_t *config);
*/
import "C"
import "fmt"

// RunRFC6349PathTest executes RFC 6349 path analysis.
func (c *Context) RunRFC6349PathTest(cfg *RFC6349Config) (*TCPPathInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc6349_config_t
	C.rfc6349_default_config(&cCfg)
	fillRFC6349Config(&cCfg, cfg)

	var cPath C.tcp_path_info_t
	ret := C.rfc6349_path_test(c.ctx, &cCfg, &cPath)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 6349 path test failed: %d", ret)
	}

	return &TCPPathInfo{
		PathMTU:          uint32(cPath.path_mtu),
		MSS:              uint32(cPath.mss),
		RTTMinMs:         float64(cPath.rtt_min_ms),
		RTTAvgMs:         float64(cPath.rtt_avg_ms),
		RTTMaxMs:         float64(cPath.rtt_max_ms),
		BDPBytes:         uint64(cPath.bdp_bytes),
		IdealRWND:        uint32(cPath.ideal_rwnd),
		BottleneckBWMbps: float64(cPath.bottleneck_bw_mbps),
	}, nil
}

// RunRFC6349ThroughputTest executes RFC 6349 throughput test.
func (c *Context) RunRFC6349ThroughputTest(cfg *RFC6349Config) (*RFC6349Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cCfg C.rfc6349_config_t
	C.rfc6349_default_config(&cCfg)
	fillRFC6349Config(&cCfg, cfg)

	var cResult C.rfc6349_result_t
	ret := C.rfc6349_throughput_test(c.ctx, &cCfg, &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("RFC 6349 throughput test failed: %d", ret)
	}

	return &RFC6349Result{
		AchievedRateMbps:    float64(cResult.achieved_rate_mbps),
		TheoreticalRateMbps: float64(cResult.theoretical_rate_mbps),
		RTTMinMs:            float64(cResult.rtt_min_ms),
		RTTAvgMs:            float64(cResult.rtt_avg_ms),
		RTTMaxMs:            float64(cResult.rtt_max_ms),
		BDPBytes:            uint64(cResult.bdp_bytes),
		RWNDUsed:            uint32(cResult.rwnd_used),
		BytesTransferred:    uint64(cResult.bytes_transferred),
		Retransmissions:     uint64(cResult.retransmissions),
		TestDurationMs:      uint32(cResult.test_duration_ms),
		TCPEfficiency:       float64(cResult.tcp_efficiency),
		BufferDelayPct:      float64(cResult.buffer_delay_pct),
		TransferTimeRatio:   float64(cResult.transfer_time_ratio),
		Passed:              bool(cResult.passed),
	}, nil
}

func fillRFC6349Config(cCfg *C.rfc6349_config_t, cfg *RFC6349Config) {
	if cfg == nil {
		return
	}
	if cfg.TargetRateMbps > 0 {
		cCfg.target_rate_mbps = C.double(cfg.TargetRateMbps)
	}
	if cfg.MinRTTMs > 0 {
		cCfg.min_rtt_ms = C.double(cfg.MinRTTMs)
	}
	if cfg.MaxRTTMs > 0 {
		cCfg.max_rtt_ms = C.double(cfg.MaxRTTMs)
	}
	if cfg.RWNDSize > 0 {
		cCfg.rwnd_size = C.uint32_t(cfg.RWNDSize)
	}
	if cfg.DurationSec > 0 {
		cCfg.test_duration_sec = C.uint32_t(cfg.DurationSec)
	}
	if cfg.ParallelStreams > 0 {
		cCfg.parallel_streams = C.uint32_t(cfg.ParallelStreams)
	}
	if cfg.MSS > 0 {
		cCfg.mss = C.uint32_t(cfg.MSS)
	}
	if cfg.Mode > 0 {
		cCfg.mode = C.tcp_test_mode_t(cfg.Mode)
	}
}
