/*
 * test_throughput_measured.c - a throughput result must be a measurement (#1233)
 *
 * The binary search judged each trial on loss alone. A software generator that
 * cannot reach the offered load still gets every frame it sent returned, so
 * loss reads 0 % and the search walks the rate to the top of its range; the
 * result then reported the offered rate as achieved. That is how a run which
 * put ~110 pps on the wire reported 9990 Mbps at 99.90 % of line rate.
 *
 * This runs a real trial on `lo`, where the generator is nowhere near 10 Gbps
 * and nothing reflects the frames back, so the shortfall is unmissable.
 */

#include <stdio.h>

#include "rfc2544.h"
#include "rfc2544_internal.h"

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

int main(void)
{
    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return fail("rfc2544_init on lo");
    }

    rfc2544_config_t config;
    rfc2544_default_config(&config);
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.frame_size         = FRAME_SIZE_64;
    config.trial_duration_sec = 1;
    config.warmup_sec         = 0;
    config.resolution_pct     = 1;
    config.max_iterations     = 4;

    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_configure");
    }

    /* The line rate `lo` reports is not 10 Gbps everywhere, so pin it: the
     * assertions below are about the gap between offered and achieved, and
     * that gap has to be a known quantity. */
    ctx->line_rate = 10000000000ULL;

    throughput_result_t result;
    uint32_t            count = 1;
    if (rfc2544_throughput_test(ctx, FRAME_SIZE_64, &result, &count) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_throughput_test");
    }

    const uint64_t max_pps      = rfc2544_calc_pps(ctx->line_rate, FRAME_SIZE_64);
    const double   demanded_pps = (double)max_pps * result.offered_rate_pct / 100.0;
    rfc2544_cleanup(ctx);

    fprintf(stderr,
            "offered %.2f%% (%.0f pps demanded), measured %.0f pps / %.2f Mbps / %.4f%%, "
            "generator_limited=%s, iterations=%u\n",
            result.offered_rate_pct, demanded_pps, result.max_rate_pps, result.max_rate_mbps,
            result.max_rate_pct, result.generator_limited ? "true" : "false", result.iterations);

    if (result.iterations == 0) {
        return fail("no iterations ran, so the assertions below prove nothing");
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
