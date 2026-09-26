/*
 * test_throughput_measured.c - a throughput result must be a measurement (#1233)
 *
 * The binary search judged each trial on loss alone. A software generator that
 * cannot reach the offered load still gets every frame it sent returned, so
 * loss reads 0 % and the search walks the rate to the top of its range; the
 * result then reported the offered rate as achieved. That is how a run which
 * put ~110 pps on the wire reported 9990 Mbps at 99.90 % of line rate.
 *
 * The search is exercised on the fake platform with a reflector that returns
 * every frame. This used to run on `lo`, where nothing is reflected: every
 * trial lost every frame and the test passed only because a generator-limited
 * trial was reported without its loss being checked (#1446). `lo` is still the
 * right place to prove the real generator is not held to the ~100 pps ceiling
 * of #1239, so that is measured from a trial there directly.
 */

#include "fake_platform.h"

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

static int search_reports_what_the_wire_carried(void)
{
    rfc2544_ctx_t *ctx = fake_context(true);
    if (!ctx) {
        return fail("fake context");
    }

    throughput_result_t result;
    uint32_t            count        = 1;
    int                 ret          = rfc2544_throughput_test(ctx, FRAME_SIZE_64, &result, &count);
    const uint64_t      max_pps      = rfc2544_calc_pps(ctx->line_rate, FRAME_SIZE_64);
    const double        demanded_pps = (double)max_pps * result.offered_rate_pct / 100.0;
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        return fail("rfc2544_throughput_test");
    }

    fprintf(stderr,
            "offered %.2f%% (%.0f pps demanded), measured %.0f pps / %.2f Mbps / %.4f%%, "
            "generator_limited=%s, iterations=%u\n",
            result.offered_rate_pct, demanded_pps, result.max_rate_pps, result.max_rate_mbps,
            result.max_rate_pct, result.generator_limited ? "true" : "false", result.iterations);

    if (result.iterations == 0 || result.max_rate_pps <= 0.0) {
        return fail("no lossless trial was reported, so the assertions below prove nothing");
    }
    /* #1466: the Go verdict tells a silent peer from a lossy one by this count. */
    if (result.frames_received == 0 || result.frames_received > result.frames_tested) {
        return fail("frames_received is not what the reflector returned");
    }

    /* The defect: all three max_rate_* fields were derived from the offered
     * rate, so max_rate_pps came back equal to demanded_pps and max_rate_pct
     * equal to offered_rate_pct. */
    if (result.max_rate_pps >= demanded_pps * (1.0 - RFC2544_GENERATOR_TOLERANCE)) {
        return fail("reported pps is the offered load, not a measurement");
    }
    if (result.max_rate_pct >= result.offered_rate_pct) {
        return fail("reported % of line rate is the offered load, not a measurement");
    }
    if (!result.generator_limited) {
        return fail("a trial this far short of its offered load must be flagged");
    }

    /* Flagging the shortfall is only honest if the search stopped there:
     * every later iteration would measure the same ceiling. */
    if (result.iterations > 1) {
        return fail("the search must stop at the first generator-limited trial");
    }
    return 0;
}

/* #1239: a blocking, timed receive after every frame held the AF_PACKET
 * generator to ~100 pps on any medium. The floor is 100x that ceiling. */
static int real_generator_clears_the_blocking_receive_ceiling(void)
{
    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return fail("rfc2544_init on lo");
    }

    rfc2544_config_t config;
    rfc2544_default_config(&config);
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.frame_size = FRAME_SIZE_64;
    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_configure");
    }
    ctx->line_rate = 10000000000ULL;

    trial_result_t trial;
    int            ret = run_trial(ctx, FRAME_SIZE_64, 50.0, 1, 0, &trial);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        return fail("run_trial on lo");
    }

    fprintf(stderr, "lo trial: sent %lu frames, %.0f pps achieved\n",
            (unsigned long)trial.packets_sent, trial.achieved_pps);

    if (trial.achieved_pps < 10000.0) {
        return fail("the generator is still held to the blocking-receive ceiling (#1239)");
    }
    return 0;
}

int main(void)
{
    int failures = search_reports_what_the_wire_carried();
    failures += real_generator_clears_the_blocking_receive_ceiling();
    return failures == 0 ? 0 : 1;
}
