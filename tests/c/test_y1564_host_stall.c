/*
 * test_y1564_host_stall.c - a test-host stall is reported, not hidden (#1717)
 *
 * A reply is timestamped when it is polled, so a stall of the test host adds
 * delay the path never did, and nothing in the result said so. This dataplane
 * stalls one receive poll in the first step's measured window for longer than
 * either delay threshold: that step must report the stall and be host limited,
 * the other steps must not, and no verdict may change.
 */

#include <arpa/inet.h>

#include "fake_platform.h"

#define STALL_US        120000
#define STALL_SEQ       300
#define THRESHOLD_MS    50.0
#define MIN_REPORTED_MS 100.0

/* Offset of the Y.1564 sequence number: Ethernet, IPv4 and UDP headers, then
 * the signature. */
#define SEQ_OFFSET (14 + 20 + 8 + Y1564_SIG_LEN)

static bool stalled;

/* Warmup frames all carry sequence number 0 and measured frames count up from
 * 0, so STALL_SEQ is first seen inside step 1's measured window. */
static int stalling_recv(worker_ctx_t *wctx, packet_t *pkts, int max_count)
{
    int n = fake_recv(wctx, pkts, max_count);
    for (int i = 0; i < n && !stalled; i++) {
        uint32_t seq;
        memcpy(&seq, pkts[i].data + SEQ_OFFSET, sizeof(seq));
        if (ntohl(seq) == STALL_SEQ) {
            stalled = true;
            usleep(STALL_US);
        }
    }
    return n;
}

static const struct platform_ops stalling_ops = {
    .name             = "stalling",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = fake_send,
    .recv_batch       = stalling_recv,
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
    ctx->config.y1564.step_duration_sec = 1;
    ctx->platform                       = &stalling_ops;

    y1564_service_t service = {.service_id = 1, .frame_size = 512, .enabled = true};
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps         = 10.0;
    service.sla.fd_threshold_ms  = THRESHOLD_MS;
    service.sla.fdv_threshold_ms = THRESHOLD_MS;

    y1564_config_result_t result;
    int                   ret = y1564_config_test(ctx, &service, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_config_test: %d\n", ret);
        return fail("the configuration test did not run");
    }
    if (!stalled) {
        return fail("step 1 never sent the frame that stalls the host");
    }

    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        const y1564_step_result_t *step = &result.steps[i];
        fprintf(stderr, "step %u: tx=%llu stall=%.2f ms host_limited=%d pass=%d\n", step->step,
                (unsigned long long)step->frames_tx, step->host_stall_max_ms, step->host_limited,
                step->step_pass);
        /* The fake echoes each frame with its TX timestamp, so the stall adds
         * no delay here and every step passes on its own measurement. */
        if (!step->step_pass) {
            return fail("a host stall changed a step's verdict");
        }
        if (i == 0) {
            if (step->host_stall_max_ms < MIN_REPORTED_MS) {
                return fail("the stalled step did not report the stall");
            }
            if (!step->host_limited) {
                return fail("a stall longer than the delay thresholds is not host limited");
            }
        } else if (step->host_limited || step->host_stall_max_ms >= THRESHOLD_MS) {
            return fail("a step without a stall is host limited");
        }
    }
    return 0;
}
