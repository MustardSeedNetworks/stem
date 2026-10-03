/*
 * test_y1564_warmup_stragglers.c - frames sent in warmup are not counted (#1457)
 *
 * A step resets its counters when warmup ends, but frames sent just before are
 * still in flight. Counting them on receipt made a step report more frames back
 * than it sent, and FLR clamped to 0, so real loss above the SLA threshold
 * passed. This dataplane delivers every frame 50 ms late, so each step's last
 * warmup frames arrive inside its measured window, and drops a known set of
 * measured frames: every step must report exactly that loss.
 */

#include <time.h>

#include <arpa/inet.h>

#include "fake_platform.h"

#define DELAY_NS    50000000ULL
#define QUEUE_DEPTH 8192
#define DROP_EVERY  97

/* Offset of the Y.1564 sequence number: Ethernet, IPv4 and UDP headers, then
 * the signature. */
#define SEQ_OFFSET (14 + 20 + 8 + Y1564_SIG_LEN)

static packet_t delayed[QUEUE_DEPTH];
static uint64_t due_ns[QUEUE_DEPTH];
static int      delayed_head;
static int      delayed_count;
static uint64_t dropped;

static uint64_t now_ns(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

/* Warmup frames all carry sequence number 0 and measured frames count up from
 * 0, so a nonzero sequence number is always a measured frame. */
static int delaying_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    usleep(20);
    for (int i = 0; i < count; i++) {
        uint32_t seq;
        memcpy(&seq, pkts[i].data + SEQ_OFFSET, sizeof(seq));
        seq = ntohl(seq);
        if (seq != 0 && seq % DROP_EVERY == 0) {
            dropped++;
            continue;
        }
        if (delayed_count == QUEUE_DEPTH) {
            continue;
        }
        uint8_t *copy = malloc(pkts[i].len);
        if (!copy) {
            continue;
        }
        memcpy(copy, pkts[i].data, pkts[i].len);
        int slot      = (delayed_head + delayed_count++) % QUEUE_DEPTH;
        delayed[slot] = (packet_t){.data = copy, .len = pkts[i].len};
        due_ns[slot]  = now_ns() + DELAY_NS;
    }
    return count;
}

static int delaying_recv(worker_ctx_t *wctx, packet_t *pkts, int max_count)
{
    (void)wctx;
    uint64_t now = now_ns();
    int      n   = 0;
    while (n < max_count && delayed_count > 0 && due_ns[delayed_head] <= now) {
        pkts[n]           = delayed[delayed_head];
        pkts[n].timestamp = now;
        n++;
        delayed_head = (delayed_head + 1) % QUEUE_DEPTH;
        delayed_count--;
    }
    return n;
}

static const struct platform_ops delaying_ops = {
    .name             = "delaying",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = delaying_send,
    .recv_batch       = delaying_recv,
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
    ctx->platform                       = &delaying_ops;

    y1564_service_t service = {.service_id = 1, .frame_size = 512, .enabled = true};
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps = 10.0;

    y1564_config_result_t result;
    int                   ret = y1564_config_test(ctx, &service, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_config_test: %d\n", ret);
        return fail("the configuration test did not run");
    }

    uint64_t lost = 0;
    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        const y1564_step_result_t *step = &result.steps[i];
        fprintf(stderr, "step %u: tx=%llu rx=%llu flr=%.4f%% pass=%d\n", step->step,
                (unsigned long long)step->frames_tx, (unsigned long long)step->frames_rx,
                step->flr_pct, step->step_pass);
        if (step->frames_tx < DROP_EVERY) {
            return fail("a step sent too few frames to lose one");
        }
        if (step->frames_rx >= step->frames_tx) {
            return fail("a step that lost measured frames reported none lost");
        }
        if (step->flr_pass) {
            return fail("a step that lost over 1 % of its frames passed FLR");
        }
        lost += step->frames_tx - step->frames_rx;
    }
    fprintf(stderr, "lost %llu, dropped %llu\n", (unsigned long long)lost,
            (unsigned long long)dropped);
    if (lost != dropped) {
        return fail("the steps' loss is not the measured frames the dataplane dropped");
    }
    return 0;
}
