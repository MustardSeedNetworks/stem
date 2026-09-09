#include <errno.h>
#include <stdio.h>

#include "rfc2544.h"
#include "rfc2544_internal.h"

int main(void)
{
    rfc2544_ctx_t *ctx = NULL;
    if (rfc2544_init(&ctx, "lo") < 0) {
        return 1;
    }

    rfc2544_config_t config;
    rfc2544_default_config(&config);
    snprintf(config.interface, sizeof(config.interface), "lo");
    config.force_packet       = true;
    config.frame_size         = 128;
    config.trial_duration_sec = 1;
    config.warmup_sec         = 0;

    if (rfc2544_configure(ctx, &config) < 0) {
        rfc2544_cleanup(ctx);
        return 1;
    }

    latency_result_t result;
    int              ret = rfc2544_latency_test(ctx, config.frame_size, 10.0, &result);

    if (ret == 0) {
        config.force_packet = false;
        if (rfc2544_configure(ctx, &config) < 0 || ctx->platform || ctx->workers) {
            rfc2544_cleanup(ctx);
            return 1;
        }
    }
    rfc2544_cleanup(ctx);

    if (ret != 0 && ret != -EIO) {
        fprintf(stderr, "unexpected latency result: %d\n", ret);
        return 1;
    }
    return 0;
}
