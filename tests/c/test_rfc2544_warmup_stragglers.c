/*
 * test_rfc2544_warmup_stragglers.c - only measured frames count as received (#1551)
 *
 * A trial resets its counters when warmup ends, but frames sent before that
 * are still in flight, and so are the last frames of the trial before it.
 * Counting them on receipt made a trial report more frames back than it sent,
 * and loss clamped to 0, so frame loss, throughput and back-to-back hid real
 * loss. This dataplane delivers every frame late and drops a known set of
 * measured frames: every trial must report exactly that loss.
 */

#include <time.h>

#include "fake_platform.h"

/* Defined in packet.c, which has no header of its own. */
uint32_t rfc2544_get_seq_num(const uint8_t *data, uint32_t len);
uint32_t custom_get_seq_num(const uint8_t *data, uint32_t len, const char *signature);

#define QUEUE_DEPTH 8192
#define DROP_EVERY  97
#define FRAME_SIZE  512
#define RATE_PCT    0.5

/* Under the 100 ms the trial waits for stragglers, so all of a trial's own
 * frames come back; over it, so the end of a trial lands in the next one. */
#define SHORT_DELAY_NS 50000000ULL
#define LONG_DELAY_NS  150000000ULL

static packet_t    delayed[QUEUE_DEPTH];
static uint64_t    due_ns[QUEUE_DEPTH];
static int         delayed_head;
static int         delayed_count;
static uint64_t    delay_ns;
static uint64_t    dropped;
static const char *signature;

static uint64_t now_ns(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

static uint32_t seq_of(const packet_t *pkt)
{
    if (signature) {
        return custom_get_seq_num(pkt->data, pkt->len, signature);
    }
    return rfc2544_get_seq_num(pkt->data, pkt->len);
}

/* Warmup frames all carry sequence number 0 and measured frames count up from
 * 0, so a nonzero sequence number is always a measured frame. */
static int delaying_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    usleep(20);
    for (int i = 0; i < count; i++) {
        uint32_t seq = seq_of(&pkts[i]);
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
        due_ns[slot]  = now_ns() + delay_ns;
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

static int trial(rfc2544_ctx_t *ctx, uint32_t warmup_sec, trial_result_t *result)
{
    if (signature) {
        return run_trial_custom(ctx, FRAME_SIZE, RATE_PCT, 1, warmup_sec, signature, 1, result);
    }
    return run_trial(ctx, FRAME_SIZE, RATE_PCT, 1, warmup_sec, result);
}

/* Runs one trial and checks that its loss is exactly the measured frames the
 * dataplane dropped. */
static int measured_trial(rfc2544_ctx_t *ctx, const char *name, uint32_t warmup_sec)
{
    trial_result_t result;
    delay_ns = SHORT_DELAY_NS;
    dropped  = 0;
    if (trial(ctx, warmup_sec, &result) < 0) {
        fprintf(stderr, "FAIL %s: the trial did not run\n", name);
        return 1;
    }
    fprintf(stderr, "%s: sent=%llu recv=%llu dropped=%llu loss=%.4f%%\n", name,
            (unsigned long long)result.packets_sent, (unsigned long long)result.packets_recv,
            (unsigned long long)dropped, result.loss_pct);
    if (result.packets_sent < DROP_EVERY) {
        fprintf(stderr, "FAIL %s: too few frames sent to lose one\n", name);
        return 1;
    }
    if (result.packets_recv != result.packets_sent - dropped) {
        fprintf(stderr, "FAIL %s: received is not sent minus the measured frames dropped\n", name);
        return 1;
    }
    if (result.loss_pct <= 0.0) {
        fprintf(stderr, "FAIL %s: a trial that lost measured frames reported no loss\n", name);
        return 1;
    }
    return 0;
}

static int run_cases(rfc2544_ctx_t *ctx)
{
    int failures = measured_trial(ctx, "warmup", 1);

    /* The earlier trial's tail is still in flight when the next one, which
     * has no warmup, starts measuring. */
    trial_result_t earlier;
    delay_ns = LONG_DELAY_NS;
    if (trial(ctx, 0, &earlier) < 0) {
        fprintf(stderr, "FAIL: the earlier trial did not run\n");
        return failures + 1;
    }
    if (delayed_count == 0) {
        fprintf(stderr, "FAIL: nothing from the earlier trial is still in flight\n");
        return failures + 1;
    }
    return failures + measured_trial(ctx, "after an earlier trial", 0);
}

int main(void)
{
    rfc2544_ctx_t *ctx = fake_context(true);
    if (!ctx) {
        fprintf(stderr, "FAIL: fake context\n");
        return 1;
    }
    ctx->platform = &delaying_ops;

    fprintf(stderr, "run_trial\n");
    signature    = NULL;
    int failures = run_cases(ctx);

    fprintf(stderr, "run_trial_custom\n");
    signature = "Y.1731 ";
    failures += run_cases(ctx);

    rfc2544_cleanup(ctx);
    return failures == 0 ? 0 : 1;
}
