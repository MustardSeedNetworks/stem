/*
 * fake_platform.h - a dataplane that needs no socket, for the search tests
 *
 * Its send is slow enough that every trial is generator-limited, and its
 * receive either returns every frame sent (a working reflector) or nothing (no
 * reflector), so a search's outcome is deterministic and needs no privileges.
 * The layout below mirrors `struct platform_ops` and `packet_t` in core.c, as
 * each platform implementation does.
 */

#ifndef STEM_TESTS_FAKE_PLATFORM_H
#define STEM_TESTS_FAKE_PLATFORM_H

#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <unistd.h>

#include "rfc2544.h"
#include "rfc2544_internal.h"

typedef struct {
    uint8_t *data;
    uint32_t len;
    uint64_t timestamp;
    uint32_t seq_num;
    void    *platform_data;
} packet_t;

struct platform_ops {
    const char *name;
    int (*init)(rfc2544_ctx_t *ctx, worker_ctx_t *wctx);
    void (*cleanup)(worker_ctx_t *wctx);
    int (*send_batch)(worker_ctx_t *wctx, packet_t *pkts, int count);
    int (*recv_batch)(worker_ctx_t *wctx, packet_t *pkts, int max_count);
    void (*release_batch)(worker_ctx_t *wctx, packet_t *pkts, int count);
    uint64_t (*get_tx_timestamp)(worker_ctx_t *wctx, packet_t *pkt);
    uint64_t (*get_rx_timestamp)(worker_ctx_t *wctx, packet_t *pkt);
};

#define FAKE_ECHO_DEPTH 64

static bool     fake_reflector_present;
static packet_t fake_echoed[FAKE_ECHO_DEPTH];
static int      fake_echoed_count;

static int fake_init(rfc2544_ctx_t *ctx, worker_ctx_t *wctx)
{
    (void)ctx;
    (void)wctx;
    return 0;
}

static void fake_cleanup(worker_ctx_t *wctx)
{
    (void)wctx;
}

/* Tens of microseconds a frame hold the generator to a few thousand frames a
 * second, orders of magnitude short of the 7.4 Mpps a 64-byte trial at 50 % of
 * 10 Gbps demands. */
static int fake_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    usleep(20);
    for (int i = 0; i < count && fake_reflector_present && fake_echoed_count < FAKE_ECHO_DEPTH;
         i++) {
        uint8_t *copy = malloc(pkts[i].len);
        if (!copy) {
            continue;
        }
        memcpy(copy, pkts[i].data, pkts[i].len);
        fake_echoed[fake_echoed_count++] =
            (packet_t){.data = copy, .len = pkts[i].len, .timestamp = pkts[i].timestamp};
    }
    return count;
}

static int fake_recv(worker_ctx_t *wctx, packet_t *pkts, int max_count)
{
    (void)wctx;
    int n = fake_echoed_count < max_count ? fake_echoed_count : max_count;
    memcpy(pkts, fake_echoed, (size_t)n * sizeof(packet_t));
    memmove(fake_echoed, fake_echoed + n, (size_t)(fake_echoed_count - n) * sizeof(packet_t));
    fake_echoed_count -= n;
    return n;
}

static void fake_release(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    for (int i = 0; i < count; i++) {
        free(pkts[i].data);
        pkts[i].data = NULL;
    }
}

static uint64_t fake_timestamp(worker_ctx_t *wctx, packet_t *pkt)
{
    (void)wctx;
    return pkt->timestamp;
}

static const struct platform_ops fake_ops = {
    .name             = "fake",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = fake_send,
    .recv_batch       = fake_recv,
    .release_batch    = fake_release,
    .get_tx_timestamp = fake_timestamp,
    .get_rx_timestamp = fake_timestamp,
};

/* A 10 Gbps context whose dataplane is the fake: prepare_platform keeps a
 * platform that already has workers, so no socket is ever opened. */
static rfc2544_ctx_t *fake_context(bool with_reflector)
{
    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return NULL;
    }

    rfc2544_config_t config;
    rfc2544_default_config(&config);
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.frame_size         = FRAME_SIZE_64;
    config.trial_duration_sec = 1;
    config.warmup_sec         = 0;
    config.resolution_pct     = 1;
    config.max_iterations     = 4;
    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return NULL;
    }

    ctx->line_rate = 10000000000ULL;
    ctx->workers   = calloc(1, sizeof(worker_ctx_t));
    if (!ctx->workers) {
        rfc2544_cleanup(ctx);
        return NULL;
    }
    ctx->num_workers = 1;
    ctx->platform    = &fake_ops;

    fake_reflector_present = with_reflector;
    return ctx;
}

#endif /* STEM_TESTS_FAKE_PLATFORM_H */
