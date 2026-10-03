/*
 * test_wire_frame_size.c - every trial path puts the requested frame size on
 * the wire (#1265)
 *
 * A frame size is the Ethernet frame including its FCS, which the NIC appends.
 * The RFC 2544 path already handed the platform a buffer four bytes shorter;
 * the custom-signature path (TSN, TrafficGen, Y.1731, MEF) and Y.1564 handed it
 * the whole frame size, so each frame went out four bytes longer than asked.
 */

#include "fake_platform.h"

static uint32_t sent_len_min;
static uint32_t sent_len_max;

static int recording_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    for (int i = 0; i < count; i++) {
        if (pkts[i].len < sent_len_min) {
            sent_len_min = pkts[i].len;
        }
        if (pkts[i].len > sent_len_max) {
            sent_len_max = pkts[i].len;
        }
    }
    return fake_send(wctx, pkts, count);
}

static const struct platform_ops recording_ops = {
    .name             = "recording",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = recording_send,
    .recv_batch       = fake_recv,
    .release_batch    = fake_release,
    .get_tx_timestamp = fake_timestamp,
    .get_rx_timestamp = fake_timestamp,
};

static rfc2544_ctx_t *recording_context(void)
{
    rfc2544_ctx_t *ctx = fake_context(true);
    if (ctx) {
        ctx->platform = &recording_ops;
    }
    sent_len_min = UINT32_MAX;
    sent_len_max = 0;
    return ctx;
}

static int check_sent(const char *path, uint32_t frame_size)
{
    uint32_t want = frame_size - RFC2544_FCS_SIZE;
    fprintf(stderr, "%s at %u: sent %u..%u bytes, want %u\n", path, frame_size, sent_len_min,
            sent_len_max, want);
    if (sent_len_max == 0) {
        fprintf(stderr, "FAIL: %s sent nothing\n", path);
        return 1;
    }
    if (sent_len_min != want || sent_len_max != want) {
        fprintf(stderr, "FAIL: %s did not leave the FCS to the NIC\n", path);
        return 1;
    }
    return 0;
}

static int rfc2544_path(uint32_t frame_size)
{
    rfc2544_ctx_t *ctx = recording_context();
    if (!ctx) {
        return 1;
    }
    trial_result_t result;
    int            ret = run_trial(ctx, frame_size, 1.0, 1, 0, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "FAIL: run_trial: %d\n", ret);
        return 1;
    }
    return check_sent("run_trial", frame_size);
}

static int custom_path(uint32_t frame_size)
{
    rfc2544_ctx_t *ctx = recording_context();
    if (!ctx) {
        return 1;
    }
    trial_result_t result;
    int            ret = run_trial_custom(ctx, frame_size, 1.0, 1, 0, "TRAFGEN", 1, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "FAIL: run_trial_custom: %d\n", ret);
        return 1;
    }
    return check_sent("run_trial_custom", frame_size);
}

static int y1564_path(uint32_t frame_size)
{
    rfc2544_ctx_t *ctx = recording_context();
    if (!ctx) {
        return 1;
    }
    ctx->config.y1564.step_duration_sec = 1;
    y1564_service_t service = {.service_id = 1, .frame_size = frame_size, .enabled = true};
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps = 10.0;
    y1564_perf_result_t result;
    int                 ret = y1564_perf_test(ctx, &service, 1, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "FAIL: y1564_perf_test: %d\n", ret);
        return 1;
    }
    return check_sent("y1564_perf_test", frame_size);
}

int main(void)
{
    static const uint32_t sizes[] = {128, 1518};
    for (size_t i = 0; i < sizeof(sizes) / sizeof(sizes[0]); i++) {
        if (rfc2544_path(sizes[i]) || custom_path(sizes[i]) || y1564_path(sizes[i])) {
            return 1;
        }
    }
    return 0;
}
