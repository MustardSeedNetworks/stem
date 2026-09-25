/*
 * test_generator_limited_loss.c - a trial that lost frames is not a throughput,
 * however far short of its offered load the generator fell (#1446)
 *
 * #1233 made the search stop at a trial the generator could not drive to its
 * offered load, and report what that trial carried. It did so before looking
 * at the trial's loss, so a generator-limited trial with nothing reflected was
 * reported as the throughput: a run against a host with no reflector said
 * `success` at the generator's own transmit rate. On any host that cannot
 * offer half a 10 G line rate, that is the first trial of every run.
 *
 * Every trial here is generator-limited and nothing comes back.
 */

#include "fake_platform.h"

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

static int throughput_without_reflector(void)
{
    rfc2544_ctx_t *ctx = fake_context(false);
    if (!ctx) {
        return fail("fake context");
    }
    throughput_result_t result;
    uint32_t            count = 1;
    int                 ret   = rfc2544_throughput_test(ctx, FRAME_SIZE_64, &result, &count);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        return fail("rfc2544_throughput_test");
    }

    fprintf(stderr, "throughput, no reflector: %.0f pps at %.2f%% offered, iterations=%u\n",
            result.max_rate_pps, result.offered_rate_pct, result.iterations);

    if (result.iterations == 0) {
        return fail("no trial ran, so the assertions below prove nothing");
    }
    if (result.max_rate_pps != 0.0 || result.max_rate_mbps != 0.0 ||
        result.offered_rate_pct != 0.0) {
        return fail("a run in which every frame was lost reported a throughput");
    }
    if (result.iterations < 2) {
        return fail("a lossy generator-limited trial must not end the search");
    }
    return 0;
}

static int forwarding_rate_without_reflector(void)
{
    rfc2544_ctx_t *ctx = fake_context(false);
    if (!ctx) {
        return fail("fake context");
    }
    rfc2889_config_t fwd_config;
    rfc2889_default_config(&fwd_config);
    fwd_config.frame_size         = FRAME_SIZE_64;
    fwd_config.trial_duration_sec = 1;
    fwd_config.warmup_sec         = 0;

    rfc2889_fwd_result_t result;
    int                  ret = rfc2889_forwarding_test(ctx, &fwd_config, &result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        return fail("rfc2889_forwarding_test");
    }

    fprintf(stderr, "forwarding rate, no reflector: %.0f fps at %.2f%% offered, frames_tx=%lu\n",
            result.max_rate_fps, result.offered_rate_pct, (unsigned long)result.frames_tx);

    if (result.frames_tx == 0) {
        return fail("no trial transmitted, so the assertions below prove nothing");
    }
    if (result.max_rate_fps != 0.0 || result.aggregate_rate_mbps != 0.0 ||
        result.offered_rate_pct != 0.0) {
        return fail("a search in which every frame was lost reported a forwarding rate");
    }
    return 0;
}

int main(void)
{
    int failures = throughput_without_reflector();
    failures += forwarding_rate_without_reflector();
    return failures == 0 ? 0 : 1;
}
