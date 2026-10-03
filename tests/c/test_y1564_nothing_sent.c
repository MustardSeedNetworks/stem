/*
 * test_y1564_nothing_sent.c - a Y.1564 step that transmits nothing fails (#1482)
 *
 * A step whose frames never left the host (a frame larger than the interface
 * MTU is the usual cause) kept FLR, FD and FDV at zero, every SLA threshold
 * accepted those, and the service passed without a frame on the wire. A step
 * that did send is still judged on its measurements.
 */

#include "fake_platform.h"

/* The kernel refusing every frame, as it does one above the MTU. */
static int refusing_send(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    (void)wctx;
    (void)pkts;
    (void)count;
    usleep(20);
    return 0;
}

static const struct platform_ops refusing_ops = {
    .name             = "refusing",
    .init             = fake_init,
    .cleanup          = fake_cleanup,
    .send_batch       = refusing_send,
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

static rfc2544_ctx_t *y1564_context(bool sends)
{
    rfc2544_ctx_t *ctx = fake_context(true);
    if (!ctx) {
        return NULL;
    }
    ctx->config.y1564.step_duration_sec = 1;
    if (!sends) {
        ctx->platform = &refusing_ops;
    }
    return ctx;
}

static y1564_service_t test_service(void)
{
    y1564_service_t service = {.service_id = 1, .frame_size = 512, .enabled = true};
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps = 10.0;
    return service;
}

static int config_test(bool sends)
{
    rfc2544_ctx_t *ctx = y1564_context(sends);
    if (!ctx) {
        return fail("fake context");
    }
    y1564_service_t       service = test_service();
    y1564_config_result_t result;
    int                   ret = y1564_config_test(ctx, &service, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_config_test: %d\n", ret);
        return fail("the configuration test did not run");
    }

    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        const y1564_step_result_t *step = &result.steps[i];
        fprintf(stderr, "config, %s: step %u tx=%llu rx=%llu pass=%d\n",
                sends ? "sending" : "refusing", step->step, (unsigned long long)step->frames_tx,
                (unsigned long long)step->frames_rx, step->step_pass);
        if (sends && (step->frames_tx == 0 || !step->step_pass)) {
            return fail("a step that sent and got every frame back did not pass");
        }
        if (!sends && (step->frames_tx != 0 || step->step_pass || step->flr_pass || step->fd_pass ||
                       step->fdv_pass)) {
            return fail("a step that transmitted no frames passed a criterion");
        }
    }
    if (result.service_pass != sends) {
        return fail(sends ? "a service whose steps all passed failed"
                          : "a service that transmitted no frames passed");
    }
    return 0;
}

static int perf_test(bool sends)
{
    rfc2544_ctx_t *ctx = y1564_context(sends);
    if (!ctx) {
        return fail("fake context");
    }
    y1564_service_t     service = test_service();
    y1564_perf_result_t result;
    int                 ret = y1564_perf_test(ctx, &service, 1, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_perf_test: %d\n", ret);
        return fail("the performance test did not run");
    }

    fprintf(stderr, "perf, %s: tx=%llu rx=%llu pass=%d\n", sends ? "sending" : "refusing",
            (unsigned long long)result.frames_tx, (unsigned long long)result.frames_rx,
            result.service_pass);
    if (sends && (result.frames_tx == 0 || !result.service_pass)) {
        return fail("a performance test that sent and got every frame back did not pass");
    }
    if (!sends && (result.frames_tx != 0 || result.service_pass || result.flr_pass ||
                   result.fd_pass || result.fdv_pass)) {
        return fail("a performance test that transmitted no frames passed a criterion");
    }
    return 0;
}

int main(void)
{
    if (config_test(false) || config_test(true) || perf_test(false) || perf_test(true)) {
        return 1;
    }
    return 0;
}
