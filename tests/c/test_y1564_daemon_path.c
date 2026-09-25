/*
 * test_y1564_daemon_path.c - Y.1564 must run on the context the daemon builds (#1411)
 *
 * The daemon's servicetest executor creates a fresh context, configures it from
 * rfc2544_default_config and calls y1564_config_test directly. Two things were
 * missing on that path: the Y.1564 step table (the default config left it all
 * zeros, so every step was "0% CIR") and a platform (only the RFC 2544 entry
 * points prepared one, so y1564_run_step returned -EINVAL before sending).
 *
 * The second half opens an AF_PACKET socket on lo, so it needs CAP_NET_RAW, as
 * test_platform_fallback does.
 */

#include <stdio.h>

#include "rfc2544.h"
#include "rfc2544_internal.h"

static const double want_steps[Y1564_CONFIG_STEPS] = {25.0, 50.0, 75.0, 100.0};

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

int main(void)
{
    rfc2544_config_t config;
    rfc2544_default_config(&config);
    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        if (config.y1564.config_steps[i] != want_steps[i]) {
            fprintf(stderr, "step %d: %.1f%%, want %.1f%%\n", i + 1, config.y1564.config_steps[i],
                    want_steps[i]);
            return fail("the default config carries no Y.1564 step table");
        }
    }

    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return fail("rfc2544_init on lo");
    }
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.y1564.step_duration_sec = 1;
    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_configure");
    }

    y1564_service_t service = {
        .service_id = 1,
        .frame_size = 512,
        .enabled    = true,
    };
    y1564_default_sla(&service.sla);
    service.sla.cir_mbps = 10.0;

    y1564_config_result_t result;
    int                   ret = y1564_config_test(ctx, &service, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        fprintf(stderr, "y1564_config_test: %d\n", ret);
        return fail("a Y.1564 config test on an unprepared context did not run");
    }

    for (int i = 0; i < Y1564_CONFIG_STEPS; i++) {
        const y1564_step_result_t *step = &result.steps[i];
        fprintf(stderr, "step %u: %.0f%% CIR, %llu frames sent\n", step->step,
                step->offered_rate_pct, (unsigned long long)step->frames_tx);
        if (step->offered_rate_pct != want_steps[i]) {
            return fail("a step ran at the wrong share of the CIR");
        }
        if (step->frames_tx == 0) {
            return fail("a step sent no frames");
        }
        if (i > 0 && step->frames_tx <= result.steps[i - 1].frames_tx) {
            return fail("a higher step did not send more frames than the one below it");
        }
    }
    return 0;
}
