/*
 * bench_reflect.c - packets-per-second microbenchmark for the reflect hot path
 *
 * stem exists to measure network performance, so a throughput regression in the
 * reflect path is a correctness regression. Nothing else in the suite measures
 * per-packet cost: tests/c/ covers behaviour, tests/load/ drives the HTTP API,
 * and the sanitizer targets cover memory safety.
 *
 * Output is one machine-readable record per case on stdout:
 *
 *     BENCH <case-name> <packets-per-second>
 *
 * scripts/bench-compare.sh runs this for two builds on the same machine and
 * compares the rates. Absolute numbers are not meaningful across runners --
 * they vary with CPU model, frequency scaling and noisy neighbours -- so
 * nothing here is compared against a recorded constant.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <time.h>

#include "reflector.h"

/* A full-size-ish frame keeps the per-packet cost dominated by header work
 * rather than by the memcpy of a jumbo payload. */
#define FRAME_LEN   128
#define IPV4_OFFSET 14
#define IPV6_OFFSET 14
#define UDP_OFFSET  34

/* Iterations per repeat, and repeats per case. The best (fastest) repeat is
 * reported: on a shared runner the slow repeats are noise from other tenants,
 * and the floor is the most stable estimator of the code's own cost. */
#define ITERATIONS 1000000
#define REPEATS    7

static uint64_t now_ns(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

static void build_ipv4_udp(uint8_t frame[FRAME_LEN])
{
    const uint8_t dst_mac[6] = {0x02, 0, 0, 0, 0, 1};
    const uint8_t src_mac[6] = {0x00, 0xc0, 0x17, 0x57, 0x01, 0x7c};
    const uint8_t src_ip[4]  = {10, 44, 10, 183};
    const uint8_t dst_ip[4]  = {10, 44, 40, 23};

    memset(frame, 0, FRAME_LEN);
    memcpy(frame, dst_mac, sizeof(dst_mac));
    memcpy(frame + 6, src_mac, sizeof(src_mac));
    frame[12]              = 0x08; /* IPv4 */
    frame[13]              = 0x00;
    frame[IPV4_OFFSET]     = 0x45; /* v4, IHL 5 */
    frame[IPV4_OFFSET + 2] = 0x00;
    frame[IPV4_OFFSET + 3] = FRAME_LEN - IPV4_OFFSET;
    frame[IPV4_OFFSET + 8] = 64; /* TTL */
    frame[IPV4_OFFSET + 9] = 17; /* UDP */
    memcpy(frame + IPV4_OFFSET + 12, src_ip, sizeof(src_ip));
    memcpy(frame + IPV4_OFFSET + 16, dst_ip, sizeof(dst_ip));
    frame[UDP_OFFSET]     = 0xc3;
    frame[UDP_OFFSET + 1] = 0x50;
    frame[UDP_OFFSET + 2] = 0x0f;
    frame[UDP_OFFSET + 3] = 0x02;
    frame[UDP_OFFSET + 4] = 0x00;
    frame[UDP_OFFSET + 5] = FRAME_LEN - UDP_OFFSET;
}

static void build_ipv6_udp(uint8_t frame[FRAME_LEN])
{
    const uint8_t dst_mac[6] = {0x02, 0, 0, 0, 0, 1};
    const uint8_t src_mac[6] = {0x00, 0xc0, 0x17, 0x57, 0x01, 0x7c};

    memset(frame, 0, FRAME_LEN);
    memcpy(frame, dst_mac, sizeof(dst_mac));
    memcpy(frame + 6, src_mac, sizeof(src_mac));
    frame[12]              = 0x86; /* IPv6 */
    frame[13]              = 0xdd;
    frame[IPV6_OFFSET]     = 0x60; /* version 6 */
    frame[IPV6_OFFSET + 4] = 0x00;
    frame[IPV6_OFFSET + 5] = FRAME_LEN - IPV6_OFFSET - 40;
    frame[IPV6_OFFSET + 6] = 17; /* next header: UDP */
    frame[IPV6_OFFSET + 7] = 64; /* hop limit */
    /* src fe80::1, dst fe80::2 */
    frame[IPV6_OFFSET + 8]  = 0xfe;
    frame[IPV6_OFFSET + 9]  = 0x80;
    frame[IPV6_OFFSET + 23] = 0x01;
    frame[IPV6_OFFSET + 24] = 0xfe;
    frame[IPV6_OFFSET + 25] = 0x80;
    frame[IPV6_OFFSET + 39] = 0x02;
}

static void build_netally_probe(uint8_t frame[FRAME_LEN])
{
    /* A fixed-length wire signature, not a C string: the field carries exactly
     * ITO_SIG_LEN bytes with no terminator. Spelled as a byte array rather than
     * a string literal so that is explicit in the type. */
    static const uint8_t signature[ITO_SIG_LEN] = {'P', 'R', 'O', 'B', 'E', 'O', 'T'};

    build_ipv4_udp(frame);
    memcpy(frame + UDP_OFFSET + UDP_HDR_LEN + ITO_SIG_OFFSET, signature, sizeof(signature));
}

/*
 * Runs one case REPEATS times and reports the best packets-per-second.
 *
 * `fn` reflects a single packet; the frame is rebuilt from `seed` before each
 * repeat so every repeat does identical work regardless of what the previous
 * one wrote into the buffer.
 */
static void run_case(const char *name, const uint8_t *seed, void (*fn)(uint8_t *, uint32_t))
{
    double   best_pps = 0.0;
    uint64_t sink     = 0;

    for (int r = 0; r < REPEATS; r++) {
        uint8_t frame[FRAME_LEN];
        memcpy(frame, seed, FRAME_LEN);

        const uint64_t start = now_ns();
        for (int i = 0; i < ITERATIONS; i++) {
            fn(frame, FRAME_LEN);
        }
        const uint64_t elapsed = now_ns() - start;

        /* Consume the reflected frame so the optimiser cannot treat the loop
         * as dead. Done after the timer stops, so it is not measured. */
        for (int b = 0; b < FRAME_LEN; b++) {
            sink += frame[b];
        }

        /* A zero-duration repeat would mean the loop was optimised away; treat
         * it as a failed measurement rather than dividing by zero. */
        if (elapsed == 0) {
            continue;
        }
        const double pps = (double)ITERATIONS * 1e9 / (double)elapsed;
        if (pps > best_pps) {
            best_pps = pps;
        }
    }

    printf("BENCH %s %.0f\n", name, best_pps);
    /* sink is folded into stderr rather than stdout so it cannot disturb the
     * machine-readable records, while still being observably used. */
    if (sink == 0) {
        fprintf(stderr, "%s: reflected frame summed to zero\n", name);
    }
}

static void reflect_mac(uint8_t *data, uint32_t len)
{
    reflect_packet_with_mode(data, len, REFLECT_MODE_MAC, false);
}

static void reflect_mac_ip(uint8_t *data, uint32_t len)
{
    reflect_packet_with_mode(data, len, REFLECT_MODE_MAC_IP, false);
}

static void reflect_mac_ip_csum(uint8_t *data, uint32_t len)
{
    reflect_packet_with_mode(data, len, REFLECT_MODE_ALL, true);
}

static void reflect_v6(uint8_t *data, uint32_t len)
{
    reflect_packet_ipv6(data, len, REFLECT_MODE_MAC_IP, false);
}

static void reflect_v6_csum(uint8_t *data, uint32_t len)
{
    reflect_packet_ipv6(data, len, REFLECT_MODE_ALL, true);
}

static void reflect_netally(uint8_t *data, uint32_t len)
{
    reflect_netally_packet(data, len, REFLECT_MODE_MAC_IP);
}

int main(void)
{
    uint8_t v4[FRAME_LEN];
    uint8_t v6[FRAME_LEN];
    uint8_t probe[FRAME_LEN];

    build_ipv4_udp(v4);
    build_ipv6_udp(v6);
    build_netally_probe(probe);

    run_case("reflect_inplace_v4", v4, reflect_packet_inplace);
    run_case("reflect_mode_mac_v4", v4, reflect_mac);
    run_case("reflect_mode_mac_ip_v4", v4, reflect_mac_ip);
    run_case("reflect_mode_all_csum_v4", v4, reflect_mac_ip_csum);
    run_case("reflect_ipv6", v6, reflect_v6);
    run_case("reflect_ipv6_csum", v6, reflect_v6_csum);
    run_case("reflect_netally_probe", probe, reflect_netally);

    return 0;
}
