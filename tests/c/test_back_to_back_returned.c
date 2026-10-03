/*
 * test_back_to_back_returned.c - a back-to-back result says what came back (#1275)
 *
 * With no reflector every burst loses every frame, so the search stops at the
 * first burst and reports a maximum burst of zero frames: the same result as a
 * peer that answers but drops part of the first burst. The Go verdict tells
 * the two apart by frames_recv, so the count must be what the peer returned.
 */

#include "fake_platform.h"

static int fail(const char *what)
{
    fprintf(stderr, "FAIL: %s\n", what);
    return 1;
}

static int run(bool with_reflector, burst_result_t *result)
{
    rfc2544_ctx_t *ctx = fake_context(with_reflector);
    if (!ctx) {
        return fail("fake context");
    }
    /* One trial of a burst whose double passes the search's 1M-frame cap, so
     * a reflector that returns everything ends the search after one pass. */
    ctx->config.initial_burst = 600000;
    ctx->config.burst_trials  = 1;

    int ret = rfc2544_back_to_back_test(ctx, FRAME_SIZE_64, result);
    rfc2544_cleanup(ctx);
    if (ret < 0) {
        return fail("rfc2544_back_to_back_test");
    }
    fprintf(stderr, "reflector=%s: max_burst=%lu trials=%u frames_recv=%lu\n",
            with_reflector ? "yes" : "no", (unsigned long)result->max_burst, result->trials,
            (unsigned long)result->frames_recv);
    return 0;
}

int main(void)
{
    burst_result_t silent = {0};
    if (run(false, &silent) != 0) {
        return 1;
    }
    if (silent.frames_recv != 0 || silent.max_burst != 0) {
        return fail("a peer that returned nothing must report no frames and no burst");
    }

    burst_result_t answered = {0};
    if (run(true, &answered) != 0) {
        return 1;
    }
    if (answered.frames_recv == 0) {
        return fail("frames_recv is not what the reflector returned");
    }
    return 0;
}
