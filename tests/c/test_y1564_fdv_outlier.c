/*
 * test_y1564_fdv_outlier.c - one late frame does not set FDV (#1681)
 *
 * FDV was the maximum delay minus the minimum, so a single frame 25 ms late on
 * a path that delays every other frame by exactly 100 us failed the default
 * 5 ms threshold. This dataplane does exactly that in every step: FDV must be
 * the 99.9th-percentile delay minus the minimum (zero here), and the late
 * frame must still show in the step's maximum delay.
 */

#include <arpa/inet.h>

#include "fake_platform.h"

#define BASE_DELAY_NS 100000ULL
#define LATE_DELAY_NS 25000000ULL
#define LATE_SEQ      500

/* Offset of the Y.1564 sequence number: Ethernet, IPv4 and UDP headers, then
 * the signature. */
#define SEQ_OFFSET (14 + 20 + 8 + Y1564_SIG_LEN)

static uint64_t late_frames;

static int skewed_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    usleep(20);
    for (int i = 0; i < count && fake_echoed_count < FAKE_ECHO_DEPTH; i++) {
        uint32_t seq;
        memcpy(&seq, pkts[i].data + SEQ_OFFSET, sizeof(seq));
        uint64_t delay = BASE_DELAY_NS;
        if (ntohl(seq) == LATE_SEQ) {
            delay = LATE_DELAY_NS;
            late_frames++;
        }
        uint8_t *copy = malloc(pkts[i].len);
        if (!copy) {
            continue;
        }
        memcpy(copy, pkts[i].data, pkts[i].len);
        fake_echoed[fake_echoed_count++] =
            (packet_t){.data = copy, .len = pkts[i].len, .timestamp = pkts[i].timestamp + delay};
    }
    return count;
}

static const struct platform_ops skewed_ops = {
    .name             = "skewed",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = skewed_send,
    .recv_batch       = fake_recv,
    .release_batch    = fake_release,
    .get_tx_timestamp = fake_timestamp,
    .get_rx_timestamp = fake_timestamp,
};

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

int main(void)
{
    rfc2544_ctx_t *ctx = fake_context(true);
    if (!ctx) {
        return fail("fake context");
    }
    ctx->config.y1564.step_duration_sec = 2;
    ctx->platform                       = &skewed_ops;

    /* At 100 Mbps the first step offers 25 Mbps, about 6,000 frames a second,
     * so even a loaded runner returns well over the 1,000 frames a 99.9th
     * percentile needs to leave one frame out. */
    y1564_service_t service = {.service_id = 1, .frame_size = 512, .enabled = true};
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps = 100.0;

    y1564_config_result_t result;
    int                   ret = y1564_config_test(ctx, &service, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_config_test: %d\n", ret);
        return fail("the configuration test did not run");
    }

    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        const y1564_step_result_t *step = &result.steps[i];
        fprintf(stderr, "step %u: rx=%llu fd min/max=%.3f/%.3f ms fdv=%.3f ms pass=%d\n",
                step->step, (unsigned long long)step->frames_rx, step->fd_min_ms, step->fd_max_ms,
                step->fdv_ms, step->fdv_pass);
        if (step->frames_rx < 1000) {
            return fail("a step returned too few frames to leave one out of the 99.9th percentile");
        }
        if (step->fd_max_ms < LATE_DELAY_NS / 1e6) {
            return fail("a step's maximum delay lost the late frame");
        }
        if (step->fdv_ms > 0.001) {
            return fail("one late frame set a step's FDV");
        }
        if (!step->fdv_pass) {
            return fail("one late frame failed a step's FDV");
        }
    }
    if (late_frames < Y1564_CONFIG_STEPS) {
        return fail("the dataplane delayed fewer frames than there are steps");
    }
    return 0;
}
