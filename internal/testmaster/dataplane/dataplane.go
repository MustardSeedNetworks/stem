//go:build cgo && linux

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This package wraps the high-performance C library for test execution,
// handling packet generation, timing, and result collection.
package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../build -lreflector -lpthread -lm
#cgo linux LDFLAGS: -lxdp -lbpf

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>

// Forward declarations for C types
typedef struct rfc2544_ctx rfc2544_ctx_t;

// Test types
typedef enum {
    TEST_THROUGHPUT = 0,
    TEST_LATENCY = 1,
    TEST_FRAME_LOSS = 2,
    TEST_BACK_TO_BACK = 3,
    TEST_SYSTEM_RECOVERY = 4,
    TEST_RESET = 5,
    TEST_Y1564_CONFIG = 6,
    TEST_Y1564_PERF = 7,
    TEST_Y1564_FULL = 8
} test_type_t;

// Test state
typedef enum {
    STATE_IDLE = 0,
    STATE_RUNNING = 1,
    STATE_COMPLETED = 2,
    STATE_FAILED = 3,
    STATE_CANCELLED = 4
} test_state_t;

// Stats format
typedef enum {
    STATS_FORMAT_TEXT = 0,
    STATS_FORMAT_JSON = 1,
    STATS_FORMAT_CSV = 2
} stats_format_t;

// Latency stats
typedef struct {
    uint64_t count;
    double min_ns;
    double max_ns;
    double avg_ns;
    double jitter_ns;
    double p50_ns;
    double p95_ns;
    double p99_ns;
} latency_stats_t;

// Throughput result
typedef struct {
    uint32_t frame_size;
    double max_rate_pct;
    double max_rate_mbps;
    double max_rate_pps;
    uint64_t frames_tested;
    uint32_t iterations;
    latency_stats_t latency;
} throughput_result_t;

// Frame loss point
typedef struct {
    double offered_rate_pct;
    double actual_rate_mbps;
    uint64_t frames_sent;
    uint64_t frames_recv;
    double loss_pct;
} frame_loss_point_t;

// Latency result
typedef struct {
    uint32_t frame_size;
    double offered_rate_pct;
    latency_stats_t latency;
} latency_result_t;

// Burst result
typedef struct {
    uint32_t frame_size;
    uint64_t max_burst;
    double burst_duration;
    uint32_t trials;
} burst_result_t;

// System recovery result (Section 26.5)
typedef struct {
    uint32_t frame_size;
    double overload_rate_pct;
    double recovery_rate_pct;
    uint32_t overload_sec;
    double recovery_time_ms;
    uint64_t frames_lost;
    uint32_t trials;
} recovery_result_t;

// Reset result (Section 26.6)
typedef struct {
    uint32_t frame_size;
    double reset_time_ms;
    uint64_t frames_lost;
    uint32_t trials;
    bool manual_reset;
} reset_result_t;

// Y.1564 SLA parameters
typedef struct {
    double cir_mbps;
    double eir_mbps;
    uint32_t cbs_bytes;
    uint32_t ebs_bytes;
    double fd_threshold_ms;
    double fdv_threshold_ms;
    double flr_threshold_pct;
} y1564_sla_t;

// Y.1564 Service configuration
typedef struct {
    uint32_t service_id;
    char service_name[32];
    y1564_sla_t sla;
    uint32_t frame_size;
    uint8_t cos;
    bool enabled;
} y1564_service_t;

// Y.1564 Step result
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

// Y.1564 Configuration test result
typedef struct {
    uint32_t service_id;
    y1564_step_result_t steps[4];
    bool service_pass;
} y1564_config_result_t;

// Y.1564 Performance test result
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


// RFC 2889 - LAN Switch Benchmarking Types
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

// RFC 6349 - TCP Throughput Testing Types

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

// ITU-T Y.1731 - Ethernet OAM Performance Monitoring Types

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
    ccm_interval_t interval;
    uint64_t ccm_sent;
    uint64_t ccm_received;
    uint64_t ccm_errors;
    bool rdi_received;
    bool connectivity_ok;
    double uptime_pct;
} y1731_ccm_result_t;

typedef struct {
    uint32_t mep_id;
    meg_level_t meg_level;
    char meg_id[32];
    ccm_interval_t ccm_interval;
    uint8_t priority;
    bool enabled;
} y1731_mep_config_t;

typedef struct {
    y1731_mep_config_t mep;
    y1731_opcode_t test_type;
    uint32_t duration_sec;
    uint32_t measurement_interval_ms;
    uint32_t frame_size;
    bool priority_tagged;
    uint8_t priority;
} y1731_config_t;

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


// MEF 48/49 - Carrier Ethernet Performance Testing Types

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

// IEEE 802.1Qbv - Time-Sensitive Networking (TSN) Types
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

// Trial result (used for custom traffic)
typedef struct {
    uint64_t packets_sent;
    uint64_t packets_recv;
    uint64_t bytes_sent;
    double loss_pct;
    double elapsed_sec;
    double achieved_pps;
    double achieved_mbps;
    latency_stats_t latency;
} trial_result_t;

// Config structure
typedef struct {
    char interface[64];
    uint64_t line_rate;
    bool auto_detect_nic;

    test_type_t test_type;
    uint32_t frame_size;
    bool include_jumbo;
    uint32_t trial_duration_sec;
    uint32_t warmup_sec;

    double initial_rate_pct;
    double resolution_pct;
    uint32_t max_iterations;
    double acceptable_loss;

    uint32_t latency_samples;
    double latency_load_pct[10];
    uint32_t latency_load_count;

    double loss_start_pct;
    double loss_end_pct;
    double loss_step_pct;

    uint64_t initial_burst;
    uint32_t burst_trials;

    bool hw_timestamp;
    bool measure_latency;

    stats_format_t output_format;
    bool verbose;

    bool use_pacing;
    uint32_t batch_size;

    bool use_dpdk;
    char *dpdk_args;
} rfc2544_config_t;

// External C functions
extern int rfc2544_init(rfc2544_ctx_t **ctx, const char *interface);
extern int rfc2544_configure(rfc2544_ctx_t *ctx, const rfc2544_config_t *config);
extern int rfc2544_run(rfc2544_ctx_t *ctx);
extern void rfc2544_cancel(rfc2544_ctx_t *ctx);
extern test_state_t rfc2544_get_state(const rfc2544_ctx_t *ctx);
extern void rfc2544_cleanup(rfc2544_ctx_t *ctx);

extern int rfc2544_throughput_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                                   throughput_result_t *result, uint32_t *result_count);
extern int rfc2544_latency_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                                double load_pct, latency_result_t *result);
extern int rfc2544_frame_loss_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                                   frame_loss_point_t *results, uint32_t *result_count);
extern int rfc2544_back_to_back_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                                     burst_result_t *result);
extern int rfc2544_system_recovery_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                                        double throughput_pct, uint32_t overload_sec,
                                        recovery_result_t *result);
extern int rfc2544_reset_test(rfc2544_ctx_t *ctx, uint32_t frame_size,
                              reset_result_t *result);

extern uint64_t rfc2544_get_line_rate(const char *interface);
extern uint64_t rfc2544_calc_pps(uint64_t line_rate, uint32_t frame_size);
extern void rfc2544_default_config(rfc2544_config_t *config);

// Y.1564 functions
extern int y1564_config_test(rfc2544_ctx_t *ctx, const y1564_service_t *service,
                             y1564_config_result_t *result);
extern int y1564_perf_test(rfc2544_ctx_t *ctx, const y1564_service_t *service,
                           uint32_t duration_sec, y1564_perf_result_t *result);
extern int y1564_multi_service_test(rfc2544_ctx_t *ctx, const y1564_service_t *services,
                                    uint32_t service_count, y1564_config_result_t *config_results,
                                    y1564_perf_result_t *perf_results);

// RFC 2889 functions
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

// RFC 6349 functions
extern int rfc6349_path_test(rfc2544_ctx_t *ctx, const rfc6349_config_t *config, tcp_path_info_t *path);
extern int rfc6349_throughput_test(rfc2544_ctx_t *ctx, const rfc6349_config_t *config,
                                 rfc6349_result_t *result);
extern void rfc6349_default_config(rfc6349_config_t *config);

// Y.1731 functions
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

// MEF functions
extern int mef_config_test(rfc2544_ctx_t *ctx, const mef_config_t *config, mef_config_result_t *result);
extern int mef_perf_test(rfc2544_ctx_t *ctx, const mef_config_t *config, mef_perf_result_t *result);
extern int mef_full_test(rfc2544_ctx_t *ctx, const mef_config_t *config,
                        mef_config_result_t *config_result, mef_perf_result_t *perf_result);
extern void mef_default_config(mef_config_t *config);

// TSN functions
extern int tsn_gate_timing_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                              tsn_timing_result_t_v2 *result);
extern int tsn_isolation_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                            tsn_isolation_result_t *result);
extern int tsn_scheduled_latency_test(rfc2544_ctx_t *ctx, const tsn_config_t *config,
                                    uint32_t traffic_class, tsn_latency_result_t *result);
extern int tsn_full_test(rfc2544_ctx_t *ctx, const tsn_config_t *config, tsn_full_result_t *result);
extern void tsn_default_config(tsn_config_t *config);

// Custom trial helper
extern int run_trial_custom(rfc2544_ctx_t *ctx, uint32_t frame_size, double rate_pct,
                          uint32_t duration_sec, uint32_t warmup_sec, const char *signature,
                          uint32_t stream_id, trial_result_t *result);
*/
import "C"

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// TestType mirrors C test_type_t
type TestType int

const (
	TestThroughput TestType = iota
	TestLatency
	TestFrameLoss
	TestBackToBack
	TestSystemRecovery
	TestReset
	TestY1564Config
	TestY1564Perf
	TestY1564Full
)

// TestState mirrors C test_state_t
type TestState int

const (
	StateIdle TestState = iota
	StateRunning
	StateCompleted
	StateFailed
	StateCancelled
)

// LatencyStats contains latency measurements
type LatencyStats struct {
	Count    uint64
	MinNs    float64
	MaxNs    float64
	AvgNs    float64
	JitterNs float64
	P50Ns    float64
	P95Ns    float64
	P99Ns    float64
}

// ThroughputResult from binary search test
type ThroughputResult struct {
	FrameSize    uint32
	MaxRatePct   float64
	MaxRateMbps  float64
	MaxRatePps   float64
	FramesTested uint64
	Iterations   uint32
	Latency      LatencyStats
}

// FrameLossPoint for a single load level
type FrameLossPoint struct {
	OfferedRatePct float64
	ActualRateMbps float64
	FramesSent     uint64
	FramesRecv     uint64
	LossPct        float64
}

// LatencyResult from latency test
type LatencyResult struct {
	FrameSize      uint32
	OfferedRatePct float64
	Latency        LatencyStats
}

// BurstResult from back-to-back test
type BurstResult struct {
	FrameSize     uint32
	MaxBurst      uint64
	BurstDuration float64
	Trials        uint32
}

// RecoveryResult from RFC 2544 Section 26.5 System Recovery test
type RecoveryResult struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResult from RFC 2544 Section 26.6 Reset test
type ResetResult struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// Y1564SLA contains SLA parameters for Y.1564 testing
type Y1564SLA struct {
	CIRMbps         float64
	EIRMbps         float64
	CBSBytes        uint32
	EBSBytes        uint32
	FDThresholdMs   float64
	FDVThresholdMs  float64
	FLRThresholdPct float64
}

// Y1564Service represents a service configuration for Y.1564 testing
type Y1564Service struct {
	ServiceID   uint32
	ServiceName string
	SLA         Y1564SLA
	FrameSize   uint32
	CoS         uint8
	Enabled     bool
}

// Y1564StepResult from a Y.1564 configuration test step
type Y1564StepResult struct {
	Step             uint32
	OfferedRatePct   float64
	AchievedRateMbps float64
	FramesTx         uint64
	FramesRx         uint64
	FLRPct           float64
	FDAvgMs          float64
	FDMinMs          float64
	FDMaxMs          float64
	FDVMs            float64
	FLRPass          bool
	FDPass           bool
	FDVPass          bool
	StepPass         bool
}

// Y1564ConfigResult from Y.1564 service configuration test
type Y1564ConfigResult struct {
	ServiceID   uint32
	Steps       [4]Y1564StepResult
	ServicePass bool
}

// Y1564PerfResult from Y.1564 service performance test
type Y1564PerfResult struct {
	ServiceID   uint32
	DurationSec uint32
	FramesTx    uint64
	FramesRx    uint64
	FLRPct      float64
	FDAvgMs     float64
	FDMinMs     float64
	FDMaxMs     float64
	FDVMs       float64
	FLRPass     bool
	FDPass      bool
	FDVPass     bool
	ServicePass bool
}

// RFC 2889 configuration and results

type RFC2889Config struct {
	FrameSize         uint32
	DurationSec       uint32
	WarmupSec         uint32
	AddressCount      uint32
	AcceptableLossPct float64
	PortCount         uint32
	Pattern           uint32
}

type RFC2889ForwardingResult struct {
	FrameSize         uint32
	PortCount         uint32
	Pattern           uint32
	MaxRatePct        float64
	MaxRateFps        float64
	AggregateRateMbps float64
	FramesTx          uint64
	FramesRx          uint64
}

type RFC2889CachingResult struct {
	AddressCount uint32
	FrameSize    uint32
	PortCount    uint32
	FramesTx     uint64
	FramesRx     uint64
	LossPct      float64
	Passed       bool
}

type RFC2889LearningResult struct {
	FrameSize           uint32
	PortCount           uint32
	LearningRateFps     float64
	AddressesLearned    uint32
	LearningTimeMs      float64
	VerificationFrames  uint32
	VerificationLossPct float64
}

type RFC2889BroadcastResult struct {
	FrameSize         uint32
	IngressPorts      uint32
	EgressPorts       uint32
	BroadcastRateFps  float64
	BroadcastRateMbps float64
	FramesTx          uint64
	FramesRx          uint64
	ReplicationFactor float64
}

type RFC2889CongestionResult struct {
	FrameSize            uint32
	OverloadRatePct      float64
	FramesTx             uint64
	FramesRx             uint64
	FramesDropped        uint64
	HeadOfLineBlocking   float64
	BackpressureObserved bool
	PauseFramesRx        uint64
}

// RFC 6349 configuration and results

type RFC6349Config struct {
	TargetRateMbps  float64
	MinRTTMs        float64
	MaxRTTMs        float64
	RWNDSize        uint32
	DurationSec     uint32
	ParallelStreams uint32
	MSS             uint32
	Mode            uint32
}

type RFC6349Result struct {
	AchievedRateMbps    float64
	TheoreticalRateMbps float64
	RTTMinMs            float64
	RTTAvgMs            float64
	RTTMaxMs            float64
	BDPBytes            uint64
	RWNDUsed            uint32
	BytesTransferred    uint64
	Retransmissions     uint64
	TestDurationMs      uint32
	TCPEfficiency       float64
	BufferDelayPct      float64
	TransferTimeRatio   float64
	Passed              bool
}

type TCPPathInfo struct {
	PathMTU          uint32
	MSS              uint32
	RTTMinMs         float64
	RTTAvgMs         float64
	RTTMaxMs         float64
	BDPBytes         uint64
	IdealRWND        uint32
	BottleneckBWMbps float64
}

// Y.1731 configuration and results

type Y1731Config struct {
	MEPID          uint32
	MEGLevel       uint32
	MEGID          string
	CCMInterval    uint32
	Priority       uint8
	DurationSec    uint32
	IntervalMs     uint32
	Count          uint32
	FrameSize      uint32
	PriorityTagged bool
}

type Y1731DelayResult struct {
	FramesSent       uint32
	FramesReceived   uint32
	FramesLost       uint32
	DelayMinUs       float64
	DelayAvgUs       float64
	DelayMaxUs       float64
	DelayVariationUs float64
}

type Y1731LossResult struct {
	FramesTx         uint64
	FramesRx         uint64
	NearEndLoss      uint64
	FarEndLoss       uint64
	NearEndLossRatio float64
	FarEndLossRatio  float64
	AvailabilityPct  float64
}

type Y1731LoopbackResult struct {
	LBMSent     uint64
	LBRReceived uint64
	RTTMinMs    float64
	RTTAvgMs    float64
	RTTMaxMs    float64
}

// MEF configuration and results

type MEFConfig struct {
	ServiceID         string
	CoS               uint32
	CIRMbps           float64
	EIRMbps           float64
	CBSBytes          uint32
	EBSBytes          uint32
	FDThresholdUs     float64
	FDVThresholdUs    float64
	FLRThresholdPct   float64
	AvailabilityPct   float64
	ConfigDurationSec uint32
	PerfDurationMin   uint32
	FrameSizes        []uint32
}

type MEFStepResult struct {
	StepPct          uint32
	OfferedRateKbps  uint32
	AchievedRateKbps uint32
	FramesTx         uint64
	FramesRx         uint64
	FDUs             float64
	FDMinUs          float64
	FDMaxUs          float64
	FDVUs            float64
	FLRPct           float64
	Passed           bool
}

type MEFConfigResult struct {
	ServiceID     string
	Steps         [4]MEFStepResult
	NumSteps      uint32
	OverallPassed bool
}

type MEFPerfResult struct {
	ServiceID       string
	DurationSec     uint32
	FramesTx        uint64
	FramesRx        uint64
	ThroughputKbps  uint32
	FDMinUs         float64
	FDAvgUs         float64
	FDMaxUs         float64
	FDVUs           float64
	FLRPct          float64
	AvailabilityPct float64
	FDPassed        bool
	FDVPassed       bool
	FLRPassed       bool
	AvailPassed     bool
	OverallPassed   bool
}

// TSN configuration and results

type TSNConfig struct {
	DurationSec       uint32
	WarmupSec         uint32
	FrameSize         uint32
	MaxLatencyNs      uint32
	MaxJitterNs       uint32
	RequirePTPSync    bool
	MaxSyncOffsetNs   uint32
	PTPEnabled        bool
	PreemptionEnabled bool
	NumTrafficClasses uint32
	BaseTimeNs        uint64
	CycleTimeNs       uint32
	TrafficClass      uint32
}

type TSNTimingResult struct {
	CyclesTested       uint32
	TimingErrors       uint32
	MaxGateDeviationNs float64
	AvgGateDeviationNs float64
	GateTimingPassed   bool
}

type TSNClassResult struct {
	FramesTx         uint64
	FramesRx         uint64
	FramesInterfered uint64
	IsolationPct     float64
	LatencyAvgNs     float64
	LatencyMaxNs     float64
	Passed           bool
}

type TSNIsolationResult struct {
	NumClasses    uint32
	ClassResults  [8]TSNClassResult
	OverallPassed bool
}

type TSNLatencyResult struct {
	TrafficClass  uint32
	Samples       uint32
	LatencyMinNs  float64
	LatencyAvgNs  float64
	LatencyMaxNs  float64
	Latency99Ns   float64
	Latency999Ns  float64
	JitterNs      float64
	LatencyPassed bool
	JitterPassed  bool
	OverallPassed bool
}

type TSNPTPResult struct {
	Samples        uint32
	OffsetAvgNs    float64
	OffsetMaxNs    float64
	OffsetStddevNs float64
	SyncAchieved   bool
}

type TSNFullResult struct {
	TimingResult    TSNTimingResult
	IsolationResult TSNIsolationResult
	LatencyResults  [8]TSNLatencyResult
	PTPResult       TSNPTPResult
	OverallPassed   bool
}

// Traffic generation configuration

type TrafficGenConfig struct {
	FrameSize       uint32
	RatePct         float64
	DurationSec     uint32
	WarmupSec       uint32
	StreamID        uint32
	BurstMode       bool
	BurstSize       uint32
	InterBurstGapUs uint32
	SrcMac          string
	DstMac          string
	VlanID          uint16
	VlanPriority    uint8
}

// Traffic generation result

type TrafficGenResult struct {
	PacketsSent  uint64
	PacketsRecv  uint64
	BytesSent    uint64
	LossPct      float64
	ElapsedSec   float64
	AchievedPPS  float64
	AchievedMbps float64
	Latency      LatencyStats
}

// Config for RFC2544 tests
type Config struct {
	Interface      string
	LineRate       uint64
	AutoDetect     bool
	TestType       TestType
	FrameSize      uint32
	IncludeJumbo   bool
	TrialDuration  time.Duration
	WarmupPeriod   time.Duration
	InitialRatePct float64
	ResolutionPct  float64
	MaxIterations  uint32
	AcceptableLoss float64
	HWTimestamp    bool
	MeasureLatency bool
	UsePacing      bool
	BatchSize      uint32
	UseDPDK        bool
	DPDKArgs       string
}

// Context wraps the C rfc2544_ctx_t
type Context struct {
	ctx       *C.rfc2544_ctx_t
	mu        sync.Mutex
	stats     Stats
	config    Config
	frameSize uint32
}

// Stats for real-time monitoring
type Stats struct {
	TxPackets   uint64
	TxBytes     uint64
	RxPackets   uint64
	RxBytes     uint64
	CurrentRate float64
	Progress    float64
	Timestamp   time.Time
}

// NewContext creates a new RFC2544 test context
func NewContext(iface string) (*Context, error) {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))

	var cctx *C.rfc2544_ctx_t
	ret := C.rfc2544_init(&cctx, cIface)
	if ret < 0 {
		return nil, fmt.Errorf("init failed: %d", ret)
	}

	return &Context{ctx: cctx}, nil
}

// Configure applies test configuration
func (c *Context) Configure(cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ccfg C.rfc2544_config_t
	C.rfc2544_default_config(&ccfg)

	// Copy interface name
	cIface := C.CString(cfg.Interface)
	defer C.free(unsafe.Pointer(cIface))
	C.strncpy(&ccfg._interface[0], cIface, 63)

	ccfg.line_rate = C.uint64_t(cfg.LineRate)
	ccfg.auto_detect_nic = C.bool(cfg.AutoDetect)
	ccfg.test_type = C.test_type_t(cfg.TestType)
	ccfg.frame_size = C.uint32_t(cfg.FrameSize)
	ccfg.include_jumbo = C.bool(cfg.IncludeJumbo)
	ccfg.trial_duration_sec = C.uint32_t(cfg.TrialDuration.Seconds())
	ccfg.warmup_sec = C.uint32_t(cfg.WarmupPeriod.Seconds())
	ccfg.initial_rate_pct = C.double(cfg.InitialRatePct)
	ccfg.resolution_pct = C.double(cfg.ResolutionPct)
	ccfg.max_iterations = C.uint32_t(cfg.MaxIterations)
	ccfg.acceptable_loss = C.double(cfg.AcceptableLoss)
	ccfg.hw_timestamp = C.bool(cfg.HWTimestamp)
	ccfg.measure_latency = C.bool(cfg.MeasureLatency)
	ccfg.use_pacing = C.bool(cfg.UsePacing)
	ccfg.batch_size = C.uint32_t(cfg.BatchSize)
	ccfg.use_dpdk = C.bool(cfg.UseDPDK)

	var dpdkArgsPtr *C.char
	if cfg.DPDKArgs != "" {
		dpdkArgsPtr = C.CString(cfg.DPDKArgs)
		ccfg.dpdk_args = dpdkArgsPtr
	}

	ret := C.rfc2544_configure(c.ctx, &ccfg)

	// Free DPDK args string after configure copies it
	if dpdkArgsPtr != nil {
		C.free(unsafe.Pointer(dpdkArgsPtr))
	}

	if ret < 0 {
		return fmt.Errorf("configure failed: %d", ret)
	}

	return nil
}

// Run starts the configured test
func (c *Context) Run() error {
	ret := C.rfc2544_run(c.ctx)
	if ret < 0 {
		return fmt.Errorf("run failed: %d", ret)
	}
	return nil
}

// Cancel stops a running test
func (c *Context) Cancel() {
	C.rfc2544_cancel(c.ctx)
}

// State returns the current test state
func (c *Context) State() TestState {
	return TestState(C.rfc2544_get_state(c.ctx))
}

// Close cleans up resources
func (c *Context) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ctx != nil {
		C.rfc2544_cleanup(c.ctx)
		c.ctx = nil
	}
}

// runThroughputTestOld executes RFC 2544 Section 26.1 throughput test (deprecated, use RunThroughputTest)
func (c *Context) runThroughputTestOld(frameSize uint32) ([]ThroughputResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 8 // 7 standard + 1 jumbo
	results := make([]C.throughput_result_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_throughput_test(c.ctx, C.uint32_t(frameSize),
		&results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("throughput test failed: %d", ret)
	}

	goResults := make([]ThroughputResult, count)
	for i := 0; i < int(count); i++ {
		goResults[i] = ThroughputResult{
			FrameSize:    uint32(results[i].frame_size),
			MaxRatePct:   float64(results[i].max_rate_pct),
			MaxRateMbps:  float64(results[i].max_rate_mbps),
			MaxRatePps:   float64(results[i].max_rate_pps),
			FramesTested: uint64(results[i].frames_tested),
			Iterations:   uint32(results[i].iterations),
			Latency: LatencyStats{
				Count:    uint64(results[i].latency.count),
				MinNs:    float64(results[i].latency.min_ns),
				MaxNs:    float64(results[i].latency.max_ns),
				AvgNs:    float64(results[i].latency.avg_ns),
				JitterNs: float64(results[i].latency.jitter_ns),
				P50Ns:    float64(results[i].latency.p50_ns),
				P95Ns:    float64(results[i].latency.p95_ns),
				P99Ns:    float64(results[i].latency.p99_ns),
			},
		}
	}

	return goResults, nil
}

// runLatencyTestOld executes RFC 2544 Section 26.2 latency test (deprecated)
func (c *Context) runLatencyTestOld(frameSize uint32, loadPct float64) (*LatencyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.latency_result_t
	ret := C.rfc2544_latency_test(c.ctx, C.uint32_t(frameSize),
		C.double(loadPct), &result)
	if ret < 0 {
		return nil, fmt.Errorf("latency test failed: %d", ret)
	}

	return &LatencyResult{
		FrameSize:      uint32(result.frame_size),
		OfferedRatePct: float64(result.offered_rate_pct),
		Latency: LatencyStats{
			Count:    uint64(result.latency.count),
			MinNs:    float64(result.latency.min_ns),
			MaxNs:    float64(result.latency.max_ns),
			AvgNs:    float64(result.latency.avg_ns),
			JitterNs: float64(result.latency.jitter_ns),
			P50Ns:    float64(result.latency.p50_ns),
			P95Ns:    float64(result.latency.p95_ns),
			P99Ns:    float64(result.latency.p99_ns),
		},
	}, nil
}

// runFrameLossTestOld executes RFC 2544 Section 26.3 frame loss test (deprecated)
func (c *Context) runFrameLossTestOld(frameSize uint32) ([]FrameLossPoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 20 // Up to 20 load levels
	results := make([]C.frame_loss_point_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_frame_loss_test(c.ctx, C.uint32_t(frameSize),
		&results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("frame loss test failed: %d", ret)
	}

	goResults := make([]FrameLossPoint, count)
	for i := 0; i < int(count); i++ {
		goResults[i] = FrameLossPoint{
			OfferedRatePct: float64(results[i].offered_rate_pct),
			ActualRateMbps: float64(results[i].actual_rate_mbps),
			FramesSent:     uint64(results[i].frames_sent),
			FramesRecv:     uint64(results[i].frames_recv),
			LossPct:        float64(results[i].loss_pct),
		}
	}

	return goResults, nil
}

// runBackToBackTestOld executes RFC 2544 Section 26.4 burst test (deprecated)
func (c *Context) runBackToBackTestOld(frameSize uint32) (*BurstResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.burst_result_t
	ret := C.rfc2544_back_to_back_test(c.ctx, C.uint32_t(frameSize), &result)
	if ret < 0 {
		return nil, fmt.Errorf("back-to-back test failed: %d", ret)
	}

	return &BurstResult{
		FrameSize:     uint32(result.frame_size),
		MaxBurst:      uint64(result.max_burst),
		BurstDuration: float64(result.burst_duration),
		Trials:        uint32(result.trials),
	}, nil
}

// GetLineRate returns the interface line rate in bits/sec
func GetLineRate(iface string) uint64 {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))
	return uint64(C.rfc2544_get_line_rate(cIface))
}

// CalcPPS calculates packets per second for given rate and frame size
func CalcPPS(lineRate uint64, frameSize uint32) uint64 {
	return uint64(C.rfc2544_calc_pps(C.uint64_t(lineRate), C.uint32_t(frameSize)))
}

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

// RunCustomStreamTest executes a custom traffic stream.
func (c *Context) RunCustomStreamTest(cfg *TrafficGenConfig) (*TrafficGenResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	frameSize := uint32(1518)
	ratePct := 10.0
	durationSec := uint32(10)
	warmupSec := uint32(1)
	streamID := uint32(0)

	if cfg != nil {
		if cfg.FrameSize > 0 {
			frameSize = cfg.FrameSize
		}
		if cfg.RatePct > 0 {
			ratePct = cfg.RatePct
		}
		if cfg.DurationSec > 0 {
			durationSec = cfg.DurationSec
		}
		if cfg.WarmupSec > 0 {
			warmupSec = cfg.WarmupSec
		}
		if cfg.StreamID > 0 {
			streamID = cfg.StreamID
		}
	}

	signature := C.CString("CUSTOM ")
	defer C.free(unsafe.Pointer(signature))

	var cResult C.trial_result_t
	ret := C.run_trial_custom(c.ctx, C.uint32_t(frameSize), C.double(ratePct), C.uint32_t(durationSec),
		C.uint32_t(warmupSec), signature, C.uint32_t(streamID), &cResult)
	if ret < 0 {
		return nil, fmt.Errorf("custom stream test failed: %d", ret)
	}

	return &TrafficGenResult{
		PacketsSent:  uint64(cResult.packets_sent),
		PacketsRecv:  uint64(cResult.packets_recv),
		BytesSent:    uint64(cResult.bytes_sent),
		LossPct:      float64(cResult.loss_pct),
		ElapsedSec:   float64(cResult.elapsed_sec),
		AchievedPPS:  float64(cResult.achieved_pps),
		AchievedMbps: float64(cResult.achieved_mbps),
		Latency: LatencyStats{
			Count:    uint64(cResult.latency.count),
			MinNs:    float64(cResult.latency.min_ns),
			MaxNs:    float64(cResult.latency.max_ns),
			AvgNs:    float64(cResult.latency.avg_ns),
			JitterNs: float64(cResult.latency.jitter_ns),
			P50Ns:    float64(cResult.latency.p50_ns),
			P95Ns:    float64(cResult.latency.p95_ns),
			P99Ns:    float64(cResult.latency.p99_ns),
		},
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

// =============================================================================
// Wrapper types and functions for CLI integration
// =============================================================================

// ThroughputResult wraps the throughput test result for CLI
type ThroughputResultCLI struct {
	FrameSize   uint32
	MaxRatePct  float64
	MaxRateMbps float64
	MaxRatePPS  float64
	Iterations  uint32
	Latency     LatencyStats
}

// LatencyResultCLI wraps the latency test result for CLI
type LatencyResultCLI struct {
	FrameSize uint32
	LoadPct   float64
	Latency   LatencyStats
}

// FrameLossResultCLI wraps the frame loss test result for CLI
type FrameLossResultCLI struct {
	FrameSize  uint32
	OfferedPct float64
	FramesTx   uint64
	FramesRx   uint64
	LossPct    float64
}

// BackToBackResultCLI wraps the back-to-back test result for CLI
type BackToBackResultCLI struct {
	FrameSize       uint32
	MaxBurstFrames  uint64
	BurstDurationUs uint64
	Trials          uint32
}

// RecoveryResultCLI wraps the system recovery test result for CLI
type RecoveryResultCLI struct {
	FrameSize       uint32
	OverloadRatePct float64
	RecoveryRatePct float64
	OverloadSec     uint32
	RecoveryTimeMs  float64
	FramesLost      uint64
	Trials          uint32
}

// ResetResultCLI wraps the reset test result for CLI
type ResetResultCLI struct {
	FrameSize   uint32
	ResetTimeMs float64
	FramesLost  uint64
	Trials      uint32
	ManualReset bool
}

// New creates a new RFC2544 context with configuration
func New(cfg Config) (*Context, error) {
	ctx, err := NewContext(cfg.Interface)
	if err != nil {
		return nil, err
	}

	if err := ctx.Configure(&cfg); err != nil {
		ctx.Close()
		return nil, err
	}

	// Store config in context for later use
	ctx.config = cfg

	return ctx, nil
}

// SetFrameSize sets the frame size for subsequent tests
func (c *Context) SetFrameSize(frameSize uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frameSize = frameSize
}

// RunThroughputTestCLI runs throughput test and returns CLI-friendly result
func (c *Context) RunThroughputTest() (*ThroughputResultCLI, error) {
	results, err := c.runThroughputTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results")
	}

	r := results[0]
	return &ThroughputResultCLI{
		FrameSize:   r.FrameSize,
		MaxRatePct:  r.MaxRatePct,
		MaxRateMbps: r.MaxRateMbps,
		MaxRatePPS:  r.MaxRatePps,
		Iterations:  r.Iterations,
		Latency:     r.Latency,
	}, nil
}

// RunLatencyTestCLI runs latency test at multiple load levels
func (c *Context) RunLatencyTest(loadLevels []float64) ([]LatencyResultCLI, error) {
	var results []LatencyResultCLI

	for _, load := range loadLevels {
		result, err := c.runLatencyTestInternal(c.frameSize, load)
		if err != nil {
			continue
		}
		results = append(results, LatencyResultCLI{
			FrameSize: c.frameSize,
			LoadPct:   load,
			Latency:   result.Latency,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no latency results")
	}

	return results, nil
}

// RunFrameLossTestCLI runs frame loss test with stepped load
func (c *Context) RunFrameLossTest(startPct, endPct, stepPct float64) ([]FrameLossResultCLI, error) {
	results, err := c.runFrameLossTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}

	var cliResults []FrameLossResultCLI
	for _, r := range results {
		cliResults = append(cliResults, FrameLossResultCLI{
			FrameSize:  c.frameSize,
			OfferedPct: r.OfferedRatePct,
			FramesTx:   r.FramesSent,
			FramesRx:   r.FramesRecv,
			LossPct:    r.LossPct,
		})
	}

	return cliResults, nil
}

// RunBackToBackTestCLI runs back-to-back burst test
func (c *Context) RunBackToBackTest(initialBurst uint64, trials uint32) (*BackToBackResultCLI, error) {
	result, err := c.runBackToBackTestInternal(c.frameSize)
	if err != nil {
		return nil, err
	}

	return &BackToBackResultCLI{
		FrameSize:       c.frameSize,
		MaxBurstFrames:  result.MaxBurst,
		BurstDurationUs: uint64(result.BurstDuration),
		Trials:          result.Trials,
	}, nil
}

// RunSystemRecoveryTest runs RFC 2544 Section 26.5 System Recovery test
func (c *Context) RunSystemRecoveryTest(throughputPct float64, overloadSec uint32) (*RecoveryResultCLI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.recovery_result_t

	ret := C.rfc2544_system_recovery_test(c.ctx, C.uint32_t(c.frameSize),
		C.double(throughputPct), C.uint32_t(overloadSec), &result)
	if ret < 0 {
		return nil, fmt.Errorf("system recovery test failed: %d", ret)
	}

	return &RecoveryResultCLI{
		FrameSize:       uint32(result.frame_size),
		OverloadRatePct: float64(result.overload_rate_pct),
		RecoveryRatePct: float64(result.recovery_rate_pct),
		OverloadSec:     uint32(result.overload_sec),
		RecoveryTimeMs:  float64(result.recovery_time_ms),
		FramesLost:      uint64(result.frames_lost),
		Trials:          uint32(result.trials),
	}, nil
}

// RunResetTest runs RFC 2544 Section 26.6 Reset test
func (c *Context) RunResetTest() (*ResetResultCLI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.reset_result_t

	ret := C.rfc2544_reset_test(c.ctx, C.uint32_t(c.frameSize), &result)
	if ret < 0 {
		return nil, fmt.Errorf("reset test failed: %d", ret)
	}

	return &ResetResultCLI{
		FrameSize:   uint32(result.frame_size),
		ResetTimeMs: float64(result.reset_time_ms),
		FramesLost:  uint64(result.frames_lost),
		Trials:      uint32(result.trials),
		ManualReset: bool(result.manual_reset),
	}, nil
}

// Internal wrappers for the existing methods
func (c *Context) runThroughputTestInternal(frameSize uint32) ([]ThroughputResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 8
	results := make([]C.throughput_result_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_throughput_test(c.ctx, C.uint32_t(frameSize), &results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("throughput test failed: %d", ret)
	}

	goResults := make([]ThroughputResult, count)
	for i := 0; i < int(count); i++ {
		goResults[i] = ThroughputResult{
			FrameSize:    uint32(results[i].frame_size),
			MaxRatePct:   float64(results[i].max_rate_pct),
			MaxRateMbps:  float64(results[i].max_rate_mbps),
			MaxRatePps:   float64(results[i].max_rate_pps),
			FramesTested: uint64(results[i].frames_tested),
			Iterations:   uint32(results[i].iterations),
			Latency: LatencyStats{
				Count:    uint64(results[i].latency.count),
				MinNs:    float64(results[i].latency.min_ns),
				MaxNs:    float64(results[i].latency.max_ns),
				AvgNs:    float64(results[i].latency.avg_ns),
				JitterNs: float64(results[i].latency.jitter_ns),
				P50Ns:    float64(results[i].latency.p50_ns),
				P95Ns:    float64(results[i].latency.p95_ns),
				P99Ns:    float64(results[i].latency.p99_ns),
			},
		}
	}

	return goResults, nil
}

func (c *Context) runLatencyTestInternal(frameSize uint32, loadPct float64) (*LatencyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.latency_result_t
	ret := C.rfc2544_latency_test(c.ctx, C.uint32_t(frameSize), C.double(loadPct), &result)
	if ret < 0 {
		return nil, fmt.Errorf("latency test failed: %d", ret)
	}

	return &LatencyResult{
		FrameSize:      uint32(result.frame_size),
		OfferedRatePct: float64(result.offered_rate_pct),
		Latency: LatencyStats{
			Count:    uint64(result.latency.count),
			MinNs:    float64(result.latency.min_ns),
			MaxNs:    float64(result.latency.max_ns),
			AvgNs:    float64(result.latency.avg_ns),
			JitterNs: float64(result.latency.jitter_ns),
			P50Ns:    float64(result.latency.p50_ns),
			P95Ns:    float64(result.latency.p95_ns),
			P99Ns:    float64(result.latency.p99_ns),
		},
	}, nil
}

func (c *Context) runFrameLossTestInternal(frameSize uint32) ([]FrameLossPoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxResults := 20
	results := make([]C.frame_loss_point_t, maxResults)
	var count C.uint32_t

	ret := C.rfc2544_frame_loss_test(c.ctx, C.uint32_t(frameSize), &results[0], &count)
	if ret < 0 {
		return nil, fmt.Errorf("frame loss test failed: %d", ret)
	}

	goResults := make([]FrameLossPoint, count)
	for i := 0; i < int(count); i++ {
		goResults[i] = FrameLossPoint{
			OfferedRatePct: float64(results[i].offered_rate_pct),
			ActualRateMbps: float64(results[i].actual_rate_mbps),
			FramesSent:     uint64(results[i].frames_sent),
			FramesRecv:     uint64(results[i].frames_recv),
			LossPct:        float64(results[i].loss_pct),
		}
	}

	return goResults, nil
}

func (c *Context) runBackToBackTestInternal(frameSize uint32) (*BurstResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result C.burst_result_t
	ret := C.rfc2544_back_to_back_test(c.ctx, C.uint32_t(frameSize), &result)
	if ret < 0 {
		return nil, fmt.Errorf("back-to-back test failed: %d", ret)
	}

	return &BurstResult{
		FrameSize:     uint32(result.frame_size),
		MaxBurst:      uint64(result.max_burst),
		BurstDuration: float64(result.burst_duration),
		Trials:        uint32(result.trials),
	}, nil
}
