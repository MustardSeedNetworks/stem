/*
 * test_stubs.c - Stub Functions for Unit Tests
 *
 * Provides minimal implementations of functions that are referenced
 * by protocol modules but not needed for config/default tests.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

#include <stdarg.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

/* Context types (minimal definitions) */
typedef void rfc2544_ctx_t;
typedef void platform_ops_t;
typedef void worker_ctx_t;

/* Stub implementations - these are not called during config tests */
platform_ops_t *rfc2544_get_platform(rfc2544_ctx_t *ctx)
{
    (void)ctx;
    return NULL;
}

worker_ctx_t *rfc2544_get_worker(rfc2544_ctx_t *ctx)
{
    (void)ctx;
    return NULL;
}

uint64_t rfc2544_get_line_rate_ctx(rfc2544_ctx_t *ctx)
{
    (void)ctx;
    return 0;
}

void rfc2544_get_macs(rfc2544_ctx_t *ctx, uint8_t *src, uint8_t *dst)
{
    (void)ctx;
    (void)src;
    (void)dst;
}

void rfc2544_get_ips(rfc2544_ctx_t *ctx, uint32_t *src, uint32_t *dst)
{
    (void)ctx;
    (void)src;
    (void)dst;
}

void rfc2544_get_ports(const rfc2544_ctx_t *ctx, uint16_t *source_port, uint16_t *remote_port)
{
    (void)ctx;
    *source_port = 12345;
    *remote_port = 3842;
}

void rfc2544_log(rfc2544_ctx_t *ctx, const char *fmt, ...)
{
    (void)ctx;
    (void)fmt;
}

void rfc2544_log_internal(const char *fmt, ...)
{
    (void)fmt;
}

bool rfc2544_is_cancelled(rfc2544_ctx_t *ctx)
{
    (void)ctx;
    return false;
}

int run_trial(rfc2544_ctx_t *ctx, uint32_t frame_size, uint64_t target_pps, uint32_t duration_ms,
              uint64_t *rx_count, uint64_t *tx_count, uint64_t *latency_ns, uint64_t *jitter_ns)
{
    (void)ctx;
    (void)frame_size;
    (void)target_pps;
    (void)duration_ms;
    (void)rx_count;
    (void)tx_count;
    (void)latency_ns;
    (void)jitter_ns;
    return 0;
}

int run_trial_custom(rfc2544_ctx_t *ctx, void *custom_cfg, uint64_t *rx_count, uint64_t *tx_count)
{
    (void)ctx;
    (void)custom_cfg;
    (void)rx_count;
    (void)tx_count;
    return 0;
}

/* Y.1564 packet stubs */
void *y1564_create_packet_template(void *cfg, uint32_t size)
{
    (void)cfg;
    (void)size;
    return NULL;
}

void y1564_stamp_packet(void *pkt, uint64_t seq, uint64_t ts)
{
    (void)pkt;
    (void)seq;
    (void)ts;
}

bool y1564_is_valid_response(void *pkt, uint32_t len, void *cfg)
{
    (void)pkt;
    (void)len;
    (void)cfg;
    return false;
}

int y1564_get_service_id(void *pkt, uint32_t len)
{
    (void)pkt;
    (void)len;
    return 0;
}

uint64_t y1564_get_tx_timestamp(void *pkt, uint32_t len)
{
    (void)pkt;
    (void)len;
    return 0;
}

/* Note: pacing and timer functions are provided by pacing.c */
