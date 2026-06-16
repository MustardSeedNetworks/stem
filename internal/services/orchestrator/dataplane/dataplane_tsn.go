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

#define TSN_MAX_GCL_ENTRIES 256

typedef uint8_t tsn_priority_t;

typedef enum {
    GATE_CLOSED = 0,
    GATE_OPEN = 1
} gate_state_t;

typedef struct {
    uint8_t gate_states;
    uint32_t time_interval_ns;
} gcl_entry_t;

typedef struct {
    uint32_t entry_count;
    gcl_entry_t entries[TSN_MAX_GCL_ENTRIES];
    uint64_t base_time_ns;
    uint32_t cycle_time_ns;
    uint32_t cycle_time_extension_ns;
} gate_control_list_t;

typedef struct {
    uint8_t dst_mac[6];
    uint16_t vlan_id;
    uint8_t priority;
    uint32_t stream_id;
} tsn_stream_id_t;

typedef struct {
    tsn_stream_id_t stream;
    double bandwidth_mbps;
    uint32_t max_frame_size;
    uint32_t max_interval_frames;
    uint32_t interval_ns;
    uint32_t max_latency_ns;
} tsn_reservation_t;

typedef struct {
    double latency_min_ns;
    double latency_avg_ns;
    double latency_max_ns;
    double jitter_ns;
    bool deadline_met;
    uint64_t frames_on_time;
    uint64_t frames_late;
    double on_time_pct;
} tsn_timing_result_t;

typedef struct {
    uint8_t gate_id;
    uint64_t frames_tx;
    uint64_t frames_rx;
    uint64_t frames_blocked;
    double gate_efficiency_pct;
    double guard_band_violation_pct;
    tsn_timing_result_t timing;
} tsn_gate_result_t;

typedef struct {
    tsn_stream_id_t stream;
    uint64_t frames_tx;
    uint64_t frames_rx;
    double throughput_mbps;
    double loss_pct;
    tsn_timing_result_t timing;
    bool reservation_met;
    bool deadline_met;
} tsn_stream_result_t;

typedef struct {
    double offset_ns;
    double offset_max_ns;
    double path_delay_ns;
    double freq_offset_ppb;
    bool sync_locked;
    uint32_t sync_steps;
} tsn_sync_result_t;

typedef struct {
    gate_control_list_t gcl;
    bool verify_gcl;
    uint32_t stream_count;
    tsn_reservation_t streams[8];
    uint32_t duration_sec;
    uint32_t warmup_sec;
    uint32_t frame_size;
    uint32_t max_latency_ns;
    uint32_t max_jitter_ns;
    bool require_ptp_sync;
    uint32_t max_sync_offset_ns;
    bool ptp_enabled;
    bool preemption_enabled;
    uint32_t num_traffic_classes;
    uint64_t base_time_ns;
    uint32_t cycle_time_ns;
} tsn_config_t;

typedef struct {
    uint32_t cycles_tested;
    uint32_t timing_errors;
    double max_gate_deviation_ns;
    double avg_gate_deviation_ns;
    bool gate_timing_passed;
} tsn_timing_result_t_v2;

typedef struct {
    uint64_t frames_tx;
    uint64_t frames_rx;
    uint64_t frames_interfered;
    double isolation_pct;
    double latency_avg_ns;
    double latency_max_ns;
    bool passed;
} tsn_class_result_t;

typedef struct {
    uint32_t num_classes;
    tsn_class_result_t class_results[8];
    bool overall_passed;
} tsn_isolation_result_t;

typedef struct {
    uint32_t traffic_class;
    uint32_t samples;
    double latency_min_ns;
    double latency_avg_ns;
    double latency_max_ns;
    double latency_99_ns;
    double latency_999_ns;
    double jitter_ns;
    bool latency_passed;
    bool jitter_passed;
    bool overall_passed;
} tsn_latency_result_t;

typedef struct {
    uint32_t samples;
    double offset_avg_ns;
    double offset_max_ns;
    double offset_stddev_ns;
    bool sync_achieved;
} tsn_ptp_result_t;

typedef struct {
    tsn_timing_result_t_v2 timing_result;
    tsn_isolation_result_t isolation_result;
    tsn_latency_result_t latency_results[8];
    tsn_ptp_result_t ptp_result;
    bool overall_passed;
} tsn_full_result_t;

extern int tsn_gate_timing_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                              tsn_timing_result_t_v2 *result);
extern int tsn_isolation_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                            tsn_isolation_result_t *result);
extern int tsn_scheduled_latency_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                                    uint32_t traffic_class, tsn_latency_result_t *result);
extern int tsn_full_test(rfc2544_ctx_t *ctx, const tsn_config_t *config, tsn_full_result_t *result);
extern void tsn_default_config(tsn_config_t *config);
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

	var cResult C.tsn_timing_result_t_v2
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
	for i := 0; i < len(result.ClassResults); i++ {
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

	for i := 0; i < len(result.IsolationResult.ClassResults); i++ {
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

	for i := 0; i < len(result.LatencyResults); i++ {
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
