/*
 * test_y1564_nothing_back.c - FD and FDV are not measured when no frame returns (#1732)
 *
 * A step that sent frames but got none back kept FD and FDV at zero for want of
 * a sample, and every SLA threshold accepted those, so both read as passed
 * beside a 100 % FLR. They must read as not measured and fail; a step whose
 * frames come back is still judged on its delay.
 */

#include "fake_platform.h"

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

static rfc2544_ctx_t *y1564_context(bool reflected)
{
    rfc2544_ctx_t *ctx = fake_context(reflected);
    if (ctx) {
        ctx->config.y1564.step_duration_sec = 1;
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

/* Delay criteria must be measured and passed with frames back, unmeasured and failed without. */
static bool delay_judged(bool reflected, uint64_t frames_tx, uint64_t frames_rx,
                         bool delay_measured, bool fd_pass, bool fdv_pass)
{
    if (frames_tx == 0) {
        return false;
    }
    if (reflected) {
        return frames_rx > 0 && delay_measured && fd_pass && fdv_pass;
    }
    return frames_rx == 0 && !delay_measured && !fd_pass && !fdv_pass;
}

static int config_test(bool reflected)
{
    rfc2544_ctx_t *ctx = y1564_context(reflected);
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
        fprintf(stderr, "config, %s: step %u tx=%llu rx=%llu measured=%d fd=%d fdv=%d\n",
                reflected ? "reflected" : "unanswered", step->step,
                (unsigned long long)step->frames_tx, (unsigned long long)step->frames_rx,
                step->delay_measured, step->fd_pass, step->fdv_pass);
        if (!delay_judged(reflected, step->frames_tx, step->frames_rx, step->delay_measured,
                          step->fd_pass, step->fdv_pass)) {
            return fail(reflected ? "a reflected step's delay was not measured and passed"
                                  : "an unanswered step's delay read as measured or passed");
        }
    }
    if (result.service_pass != reflected) {
        return fail(reflected ? "a reflected service failed" : "an unanswered service passed");
    }
    return 0;
}

static int perf_test(bool reflected)
{
    rfc2544_ctx_t *ctx = y1564_context(reflected);
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

    fprintf(stderr, "perf, %s: tx=%llu rx=%llu measured=%d fd=%d fdv=%d\n",
            reflected ? "reflected" : "unanswered", (unsigned long long)result.frames_tx,
            (unsigned long long)result.frames_rx, result.delay_measured, result.fd_pass,
            result.fdv_pass);
    if (!delay_judged(reflected, result.frames_tx, result.frames_rx, result.delay_measured,
                      result.fd_pass, result.fdv_pass)) {
        return fail(reflected
                        ? "a reflected performance test's delay was not measured and passed"
                        : "an unanswered performance test's delay read as measured or passed");
    }
    if (result.service_pass != reflected) {
        return fail(reflected ? "a reflected performance test failed"
                              : "an unanswered performance test passed");
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
