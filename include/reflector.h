/*
 * reflector.h - Core data structures and definitions for packet reflector
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 * High-performance packet reflector for network test tools
 */

#ifndef REFLECTOR_H
#define REFLECTOR_H

#include <stdatomic.h>
#include <stdbool.h>
#include <stdint.h>

#include <sys/types.h>

/* Threading support: GCD workers on macOS, pthread synchronization everywhere. */
#include <pthread.h>
#ifdef __APPLE__
#include <dispatch/dispatch.h>
#endif

/* Version information */
#include "version_generated.h"

/* Compiler hints for branch prediction */
#ifdef __GNUC__
#define likely(x)   __builtin_expect(!!(x), 1)
#define unlikely(x) __builtin_expect(!!(x), 0)
#else
#define likely(x)   (x)
#define unlikely(x) (x)
#endif

/* Force inline for hot path functions */
#ifdef __GNUC__
#define ALWAYS_INLINE __attribute__((always_inline)) inline
#else
#define ALWAYS_INLINE inline
#endif

/* Memory prefetch hints */
#ifdef __GNUC__
#define PREFETCH_READ(addr)  __builtin_prefetch(addr, 0, 3)
#define PREFETCH_WRITE(addr) __builtin_prefetch(addr, 1, 3)
#else
#define PREFETCH_READ(addr)  ((void)0)
#define PREFETCH_WRITE(addr) ((void)0)
#endif

/* Debug logging */
#ifdef ENABLE_HOT_PATH_DEBUG
#define DEBUG_LOG(fmt, ...) reflector_log(LOG_DEBUG, fmt, ##__VA_ARGS__)
#else
#define DEBUG_LOG(fmt, ...) ((void)0)
#endif

/* Configuration constants - NOLINT: C macros are correct for compile-time constants */
// NOLINTBEGIN(cppcoreguidelines-macro-to-enum,modernize-macro-to-enum)
#define MAX_IFNAME_LEN      16
#define MAX_WORKERS         16
#define BATCH_SIZE          64
#define STATS_FLUSH_BATCHES 8
#define FRAME_SIZE          4096
#define NUM_FRAMES          4096
#define UMEM_SIZE           (NUM_FRAMES * FRAME_SIZE)

/* ITO packet signatures (NetAlly/Fluke/NETSCOUT) */
#define ITO_SIG_PROBEOT         "PROBEOT"
#define ITO_SIG_DATAOT          "DATA:OT"
#define ITO_SIG_LATENCY         "LATENCY"
#define ITO_SIG_LEN             7
#define ITO_TOS_WIGGLE          0x01
#define PACKET_BLOCK_TIMEOUT_MS 1

/* Custom signatures (RFC2544/Y.1564 tester) */
#define CUSTOM_SIG_RFC2544 "RFC2544"
#define CUSTOM_SIG_Y1564   "Y.1564 "
#define CUSTOM_SIG_MSN     "MSNSEED"
#define CUSTOM_SIG_LEN     7

/* Ethernet frame offsets */
#define ETH_DST_OFFSET  0
#define ETH_SRC_OFFSET  6
#define ETH_TYPE_OFFSET 12
#define ETH_HDR_LEN     14

/* IPv4 header offsets */
#define IP_VER_IHL_OFFSET 0
#define IP_PROTO_OFFSET   9
#define IP_SRC_OFFSET     12
#define IP_DST_OFFSET     16
#define IP_HDR_MIN_LEN    20

/* UDP header offsets */
#define UDP_SRC_PORT_OFFSET 0
#define UDP_DST_PORT_OFFSET 2
#define UDP_HDR_LEN         8

/* ITO packet signature offset */
#define ITO_SIG_OFFSET 5

/* Minimum packet sizes */
#define MIN_ITO_PACKET_LEN      54
#define MIN_ITO_PACKET_LEN_IPV6 69
#define MIN_ITO_PACKET_LEN_VLAN 58

/* EtherType values */
#ifndef ETH_P_IP
#define ETH_P_IP 0x0800
#endif
#ifndef ETH_P_IPV6
#define ETH_P_IPV6 0x86DD
#endif
#ifndef ETH_P_8021Q
#define ETH_P_8021Q 0x8100
#endif
#ifndef ETH_P_8021AD
#define ETH_P_8021AD 0x88A8
#endif

/* VLAN header */
#define VLAN_HDR_LEN     4
#define VLAN_TPID_OFFSET 0
#define VLAN_TCI_OFFSET  2

/* IPv6 header offsets */
#define IPV6_HDR_LEN         40
#define IPV6_NEXT_HDR_OFFSET 6
#define IPV6_SRC_OFFSET      8
#define IPV6_DST_OFFSET      24
#define IPV6_ADDR_LEN        16

/* IP Protocol values */
#if !defined(__linux__) && !defined(IPPROTO_UDP)
#define IPPROTO_UDP 17
#endif

/* ITO test packet standard port */
#define ITO_UDP_PORT 3842

/* NetAlly OUI prefix */
#define NETALLY_OUI_BYTE0 0x00
#define NETALLY_OUI_BYTE1 0xc0
#define NETALLY_OUI_BYTE2 0x17

/* Minimum software checksum packet length */
#define MIN_CHECKSUM_PACKET_LEN 42
// NOLINTEND(cppcoreguidelines-macro-to-enum,modernize-macro-to-enum)

/* Reflection mode */
typedef enum {
    REFLECT_MODE_MAC    = 0,
    REFLECT_MODE_MAC_IP = 1,
    REFLECT_MODE_ALL    = 2
} reflect_mode_t;

/* Signature filter mode */
typedef enum {
    SIG_FILTER_ALL     = 0,
    SIG_FILTER_ITO     = 1,
    SIG_FILTER_RFC2544 = 2,
    SIG_FILTER_Y1564   = 3,
    SIG_FILTER_CUSTOM  = 4,
    SIG_FILTER_MSN     = 5,
    SIG_FILTER_PROBEOT = 6,
    SIG_FILTER_DATAOT  = 7,
    SIG_FILTER_LATENCY = 8
} sig_filter_t;

/* Packet signature types */
typedef enum {
    SIG_TYPE_PROBEOT = 0,
    SIG_TYPE_DATAOT  = 1,
    SIG_TYPE_LATENCY = 2,
    SIG_TYPE_RFC2544 = 3,
    SIG_TYPE_Y1564   = 4,
    SIG_TYPE_MSN     = 5,
    SIG_TYPE_UNKNOWN = 6,
    SIG_TYPE_COUNT   = 7
} sig_type_t;

/* Legacy alias */
typedef sig_type_t ito_sig_type_t;
#define ITO_SIG_TYPE_PROBEOT SIG_TYPE_PROBEOT
#define ITO_SIG_TYPE_DATAOT  SIG_TYPE_DATAOT
#define ITO_SIG_TYPE_LATENCY SIG_TYPE_LATENCY
#define ITO_SIG_TYPE_UNKNOWN SIG_TYPE_UNKNOWN
#define ITO_SIG_TYPE_COUNT   SIG_TYPE_COUNT

/* Error category types */
typedef enum {
    ERR_RX_INVALID_MAC = 0,
    ERR_RX_INVALID_ETHERTYPE,
    ERR_RX_INVALID_PROTOCOL,
    ERR_RX_INVALID_SIGNATURE,
    ERR_RX_TOO_SHORT,
    ERR_TX_FAILED,
    ERR_RX_NOMEM,
    ERR_CATEGORY_COUNT
} error_category_t;

/* Latency statistics */
typedef struct {
    uint64_t count;
    uint64_t total_ns;
    uint64_t min_ns;
    uint64_t max_ns;
    double   avg_ns;
} latency_stats_t;

/* Statistics structure */
typedef struct {
    uint64_t packets_received;
    uint64_t packets_reflected;
    uint64_t packets_dropped;
    uint64_t bytes_received;
    uint64_t bytes_reflected;

    uint64_t sig_probeot_count;
    uint64_t sig_dataot_count;
    uint64_t sig_latency_count;
    uint64_t sig_rfc2544_count;
    uint64_t sig_y1564_count;
    uint64_t sig_msn_count;
    uint64_t sig_unknown_count;

    uint64_t err_invalid_mac;
    uint64_t err_invalid_ethertype;
    uint64_t err_invalid_protocol;
    uint64_t err_invalid_signature;
    uint64_t err_too_short;
    uint64_t err_tx_failed;
    uint64_t err_nomem;

    uint64_t rx_invalid;
    uint64_t rx_nomem;
    uint64_t tx_errors;
    uint64_t poll_timeout;

    latency_stats_t latency;

    double pps;
    double mbps;

    uint64_t start_time_ns;
    uint64_t last_update_ns;
} reflector_stats_t;

/* Statistics output format */
typedef enum {
    STATS_FORMAT_TEXT,
    STATS_FORMAT_JSON,
    STATS_FORMAT_CSV
} stats_format_t;

/* Configuration structure */
typedef struct {
    char           ifname[MAX_IFNAME_LEN];
    int            ifindex;
    uint8_t        mac[6];
    int            num_workers;
    bool           enable_stats;
    bool           promiscuous;
    bool           zero_copy;
    int            batch_size;
    int            frame_size;
    int            num_frames;
    int            queue_id;
    bool           busy_poll;
    int            poll_timeout_ms;
    bool           measure_latency;
    stats_format_t stats_format;
    int            stats_interval_sec;
    int            cpu_affinity;
    bool           use_huge_pages;
    bool           software_checksum;

    bool  use_dpdk;
    bool  use_af_xdp;
    char *dpdk_args;

    uint16_t ito_port;
    bool     filter_oui;
    uint8_t  oui[3];
    bool     filter_dst_mac;

    reflect_mode_t reflect_mode;
    sig_filter_t   sig_filter;

    bool enable_ipv6;
    bool enable_vlan;
} reflector_config_t;

/* Packet descriptor */
typedef struct {
    uint8_t *data;
    uint32_t len;
    uint64_t addr;
    uint64_t timestamp;
} packet_t;

/* Platform-specific context */
typedef struct platform_ctx platform_ctx_t;

/* Worker thread context */
typedef struct {
    int                 worker_id;
    int                 queue_id;
    int                 cpu_id;
    platform_ctx_t     *pctx;
    reflector_config_t *config;
    reflector_stats_t   stats;
    pthread_mutex_t     stats_mutex;
    bool                stats_mutex_initialized;
    bool                thread_started;
    atomic_bool         running;
} worker_ctx_t;

/* Reflector context */
typedef struct {
    reflector_config_t config;
    platform_ctx_t   **platform_contexts;
    worker_ctx_t      *workers;
#ifdef __APPLE__
    dispatch_group_t  worker_group;
    dispatch_queue_t *worker_queues;
#else
    pthread_t *worker_tids;
#endif
    reflector_stats_t global_stats;
    atomic_bool       running;
    int               num_workers;
} reflector_ctx_t;

/* Platform abstraction interface */
typedef struct {
    const char *name;
    int (*init)(reflector_ctx_t *rctx, worker_ctx_t *wctx);
    void (*cleanup)(worker_ctx_t *wctx);
    int (*recv_batch)(worker_ctx_t *wctx, packet_t *pkts, int max_pkts);
    int (*send_batch)(worker_ctx_t *wctx, packet_t *pkts, int num_pkts);
    void (*release_batch)(worker_ctx_t *wctx, packet_t *pkts, int num_pkts);
} platform_ops_t;

/* Function declarations */
int  reflector_init(reflector_ctx_t *rctx, const char *ifname);
void reflector_cleanup(reflector_ctx_t *rctx);
int  reflector_start(reflector_ctx_t *rctx);
void reflector_stop(reflector_ctx_t *rctx);
int  reflector_set_config(reflector_ctx_t *rctx, const reflector_config_t *config);
int  reflector_update_filter(reflector_ctx_t *rctx, uint16_t port, sig_filter_t signature_filter);
void reflector_get_config(const reflector_ctx_t *rctx, reflector_config_t *config);
void reflector_get_stats(const reflector_ctx_t *rctx, reflector_stats_t *stats);
void reflector_reset_stats(reflector_ctx_t *rctx);

int      get_interface_index(const char *ifname);
int      get_interface_mac(const char *ifname, uint8_t mac[6]);
int      get_num_rx_queues(const char *ifname);
void     print_nic_recommendations(const char *ifname);
bool     is_dpdk_available(void);
int      get_nic_vendor(const char *ifname, uint16_t *vendor_id, uint16_t *device_id);
int      get_nic_speed(const char *ifname);
void     print_af_packet_warning(const char *ifname);
void     print_recommended_nics(void);
int      get_queue_cpu_affinity(const char *ifname, int queue_id);
uint64_t get_timestamp_ns(void);
int      drop_privileges(void);

bool is_ito_packet(const uint8_t *data, uint32_t len, const reflector_config_t *config);
bool is_ito_packet_extended(const uint8_t *data, uint32_t len, const reflector_config_t *config,
                            bool *is_ipv6, bool *is_vlan);
ito_sig_type_t get_ito_signature_type(const uint8_t *data, uint32_t len);
void           reflect_packet_inplace(uint8_t *data, uint32_t len);
void           reflect_packet_with_checksum(uint8_t *data, uint32_t len, bool software_checksum);
void           reflect_packet_with_mode(uint8_t *data, uint32_t len, reflect_mode_t mode,
                                        bool software_checksum);
void           reflect_netally_packet(uint8_t *data, uint32_t len, reflect_mode_t mode);
void reflect_packet_ipv6(uint8_t *data, uint32_t len, reflect_mode_t mode, bool software_checksum);
bool is_vlan_tagged(const uint8_t *data, uint32_t len, uint16_t *inner_ethertype,
                    uint32_t *vlan_offset);

void update_signature_stats(reflector_stats_t *stats, ito_sig_type_t sig_type);
void update_latency_stats(latency_stats_t *latency, uint64_t latency_ns);
void update_error_stats(reflector_stats_t *stats, error_category_t err_cat);
int  packet_platform_set_guard_port(worker_ctx_t *wctx, uint16_t port);
void reflector_print_stats_formatted(const reflector_stats_t *stats, stats_format_t format);
void reflector_print_stats_json(const reflector_stats_t *stats);
void reflector_print_stats_csv(const reflector_stats_t *stats);

const platform_ops_t *get_platform_ops(void);

typedef enum {
    LOG_DEBUG,
    LOG_INFO,
    LOG_WARN,
    LOG_ERROR
} log_level_t;
void reflector_log(log_level_t level, const char *fmt, ...);
void reflector_set_log_level(log_level_t level);

#endif /* REFLECTOR_H */
