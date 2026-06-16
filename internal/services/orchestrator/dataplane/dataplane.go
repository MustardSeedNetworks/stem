//go:build cgo && linux

// Package dataplane provides CGO bindings to the C test master dataplane.
//
// This package wraps the high-performance C library for test execution,
// handling packet generation, timing, and result collection.
package dataplane

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../../build -lreflector -lpthread -lm
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
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// ErrNotSupported is defined for interface parity across build targets.
// in the CGO build since the dataplane is available.
var ErrNotSupported = errors.New("CGO dataplane not available on this platform")

// Context wraps the C rfc2544_ctx_t
type Context struct {
	ctx       *C.rfc2544_ctx_t
	mu        sync.Mutex
	stats     Stats
	config    Config
	frameSize uint32
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

// NewTestContext creates a test context for unit tests that need a non-nil
// context but do not execute dataplane operations.
func NewTestContext() *Context {
	return &Context{}
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
