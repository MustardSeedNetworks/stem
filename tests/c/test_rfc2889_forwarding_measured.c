/*
 * test_rfc2889_forwarding_measured.c - a forwarding rate must be a
 * measurement (#1242)
 *
 * rfc2889_forwarding_test had #1233's defect in a different result struct: it
 * judged each trial on loss alone and derived all three rate fields from the
 * rate its binary search settled on. A generator that cannot reach the offered
 * load gets back every frame it sent, loss reads 0 %, and the search walks to
 * the top of its range, so the result printed ~99.9 % of line rate over a link
 * that carried a few thousand frames.
 *
 * This runs a real search on `lo` at 64-byte frames against a pinned 10 Gbps
 * line rate: the generator is nowhere near 7.4 Mpps, so the shortfall is
 * unmissable.
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
    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_configure");
    }

    /* The line rate `lo` reports is not 10 Gbps everywhere, so pin it: the
     * assertions below are about the gap between offered and achieved. */
    ctx->line_rate = 10000000000ULL;

    rfc2889_config_t fwd_config;
    rfc2889_default_config(&fwd_config);
    fwd_config.frame_size         = FRAME_SIZE_64;
    fwd_config.trial_duration_sec = 1;
    fwd_config.warmup_sec         = 0;

    rfc2889_fwd_result_t result;
    if (rfc2889_forwarding_test(ctx, &fwd_config, &result) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2889_forwarding_test");
    }

    const uint64_t max_pps      = rfc2544_calc_pps(ctx->line_rate, FRAME_SIZE_64);
    const double   demanded_fps = (double)max_pps * result.offered_rate_pct / 100.0;
    rfc2544_cleanup(ctx);

    fprintf(stderr,
            "offered %.2f%% (%.0f fps demanded), measured %.0f fps / %.2f Mbps / %.4f%%, "
            "generator_limited=%s, frames_tx=%lu\n",
            result.offered_rate_pct, demanded_fps, result.max_rate_fps, result.aggregate_rate_mbps,
            result.max_rate_pct, result.generator_limited ? "true" : "false",
            (unsigned long)result.frames_tx);

    if (result.frames_tx == 0) {
        return fail("no trial transmitted, so the assertions below prove nothing");
    }

    /* The defect: max_rate_fps was max_pps * best_rate / 100, the offered
     * load, and max_rate_pct was best_rate itself. No trial can have carried
     * more frames per second than every trial together transmitted. */
    if (result.max_rate_fps * fwd_config.trial_duration_sec > (double)result.frames_tx) {
        return fail("reported fps exceeds the frames the search transmitted");
    }
    if (result.max_rate_fps >= demanded_fps * (1.0 - RFC2544_GENERATOR_TOLERANCE)) {
        return fail("reported fps is the offered load, not a measurement");
    }
    if (result.max_rate_pct >= result.offered_rate_pct) {
        return fail("reported % of line rate is the offered load, not a measurement");
    }
    const double wire_mbps = result.max_rate_fps * (FRAME_SIZE_64 + 20) * 8 / 1e6;
    if (result.aggregate_rate_mbps > wire_mbps * 1.001) {
        return fail("reported Mbps is not derived from the measured fps");
    }
    if (!result.generator_limited) {
        return fail("a trial this far short of its offered load must be flagged");
    }

    /* Flagging the shortfall is only honest if the search stopped there:
     * every later trial would measure the same ceiling. One 1 s trial at the
     * measured rate transmits about max_rate_fps frames; the old search ran
     * several. */
    if ((double)result.frames_tx > 2.0 * result.max_rate_fps * fwd_config.trial_duration_sec) {
        return fail("the search must stop at the first generator-limited trial");
    }

    return 0;
}
