/*
 * test_platform_fallback.c - the test master must not be AF_XDP-or-nothing (#1232)
 *
 * AF_XDP is selected on every Linux build because HAVE_AF_XDP only asks whether
 * <linux/if_xdp.h> was present at compile time, which says nothing about the
 * running host. Where the UMEM cannot be created the reflector degrades to
 * AF_PACKET and keeps working; the test master used to fail the run.
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
    const platform_ops_t *preferred = rfc2544_preferred_platform();
    if (!preferred) {
        return fail("no dataplane platform is available on Linux");
    }

    const platform_ops_t *second = rfc2544_fallback_platform(preferred);
#if HAVE_AF_XDP
    /* AF_XDP is preferred and must degrade to something else. */
    if (!second || second == preferred) {
        return fail("AF_XDP must fall back to AF_PACKET");
    }
    if (rfc2544_fallback_platform(second) != NULL) {
        return fail("AF_PACKET is the last resort and must not fall back");
    }
#else
    if (second != NULL) {
        return fail("AF_PACKET is the last resort and must not fall back");
    }
#endif

    /* End to end: a run that does not force AF_PACKET must still reach a
     * platform. Before the fix this leaves ctx->platform NULL wherever AF_XDP
     * cannot initialize, which is every container and VM without XDP support. */
    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return fail("rfc2544_init on lo");
    }

    rfc2544_config_t config;
    rfc2544_default_config(&config);
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.force_packet       = false;
    config.frame_size         = 128;
    config.trial_duration_sec = 1;
    config.warmup_sec         = 0;
    config.resolution_pct     = 99;
    config.max_iterations     = 1;

    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return fail("rfc2544_configure");
    }

    throughput_result_t throughput[1];
    uint32_t            count = 1;
    (void)rfc2544_throughput_test(ctx, FRAME_SIZE_64, throughput, &count);

    const int reached_a_platform = ctx->platform != NULL;
    fprintf(stderr, "platform after an unforced run: %s\n",
            !reached_a_platform          ? "none"
            : ctx->platform == preferred ? "preferred"
                                         : "fallback");
    rfc2544_cleanup(ctx);

    if (!reached_a_platform) {
        return fail("an unforced run reached no platform at all");
    }
    return 0;
}
