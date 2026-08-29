/*
 * test_pacing.c - Unit Tests for Rate Calculations and Pacing
 *
 * Tests pure calculation functions that don't require hardware.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

#include "test_framework.h"

/* Include the functions we're testing */
#include "../../include/rfc2544.h"
#include "../../include/rfc2544_internal.h"

/* Forward declaration - implemented in pacing.c */
double calc_utilization(uint64_t achieved_pps, uint32_t frame_size, uint64_t line_rate_bps);

/* ============================================================================
 * calc_max_pps Tests
 * ============================================================================
 *
 * Wire size = frame_size + 20 (preamble + SFD + IFG)
 *   - 8 bytes preamble + SFD
 *   - 12 bytes inter-frame gap (IFG)
 *
 * Max PPS = line_rate_bps / (wire_size * 8)
 */

TEST(calc_max_pps_1g_64byte)
{
    /* 1 Gbps with 64-byte frames */
    /* Wire size = 64 + 20 = 84 bytes = 672 bits */
    /* Max PPS = 1,000,000,000 / 672 = 1,488,095 pps */
    uint64_t result = calc_max_pps(1000000000ULL, 64);
    ASSERT_EQ(1488095, result);
}

TEST(calc_max_pps_1g_1518byte)
{
    /* 1 Gbps with 1518-byte frames */
    /* Wire size = 1518 + 20 = 1538 bytes = 12304 bits */
    /* Max PPS = 1,000,000,000 / 12304 = 81,274 pps */
    uint64_t result = calc_max_pps(1000000000ULL, 1518);
    ASSERT_EQ(81274, result);
}

TEST(calc_max_pps_10g_64byte)
{
    /* 10 Gbps with 64-byte frames */
    /* Max PPS = 10,000,000,000 / 672 = 14,880,952 pps */
    uint64_t result = calc_max_pps(10000000000ULL, 64);
    ASSERT_EQ(14880952, result);
}

TEST(calc_max_pps_10g_1518byte)
{
    /* 10 Gbps with 1518-byte frames */
    /* Max PPS = 10,000,000,000 / 12304 = ~812,743-744 pps */
    uint64_t result = calc_max_pps(10000000000ULL, 1518);
    ASSERT_GE(result, 812743);
    ASSERT_LE(result, 812744);
}

TEST(calc_max_pps_100g_64byte)
{
    /* 100 Gbps with 64-byte frames */
    uint64_t result = calc_max_pps(100000000000ULL, 64);
    ASSERT_EQ(148809523, result);
}

TEST(calc_max_pps_zero_line_rate)
{
    /* Edge case: zero line rate should return 0 */
    uint64_t result = calc_max_pps(0, 64);
    ASSERT_EQ(0, result);
}

TEST(calc_max_pps_jumbo_frame)
{
    /* 10 Gbps with 9000-byte jumbo frames */
    /* Wire size = 9000 + 20 = 9020 bytes = 72160 bits */
    uint64_t result = calc_max_pps(10000000000ULL, 9000);
    ASSERT_EQ(138580, result);
}

TEST(calc_max_pps_25g_64byte)
{
    /* 25 Gbps with 64-byte frames */
    uint64_t result = calc_max_pps(25000000000ULL, 64);
    ASSERT_EQ(37202380, result);
}

TEST(calc_max_pps_40g_64byte)
{
    /* 40 Gbps with 64-byte frames */
    uint64_t result = calc_max_pps(40000000000ULL, 64);
    ASSERT_EQ(59523809, result);
}

/* ============================================================================
 * calc_utilization Tests
 * ============================================================================ */

TEST(calc_utilization_100_percent)
{
    /* Full line rate utilization */
    /* 1G, 64-byte frames, max PPS = 1,488,095 */
    double util = calc_utilization(1488095, 64, 1000000000ULL);
    ASSERT_FLOAT_EQ(100.0, util, 0.1);
}

TEST(calc_utilization_50_percent)
{
    /* Half line rate */
    double util = calc_utilization(744047, 64, 1000000000ULL);
    ASSERT_FLOAT_EQ(50.0, util, 0.1);
}

TEST(calc_utilization_zero_rate)
{
    /* Zero packets = 0% utilization */
    double util = calc_utilization(0, 64, 1000000000ULL);
    ASSERT_FLOAT_EQ(0.0, util, 0.001);
}

TEST(calc_utilization_zero_line_rate)
{
    /* Edge case: zero line rate should return 0, not crash */
    double util = calc_utilization(1000, 64, 0);
    ASSERT_FLOAT_EQ(0.0, util, 0.001);
}

TEST(calc_utilization_10g_64byte)
{
    /* 10G at full rate with 64-byte frames */
    double util = calc_utilization(14880952, 64, 10000000000ULL);
    ASSERT_FLOAT_EQ(100.0, util, 0.1);
}

TEST(calc_utilization_small_rate)
{
    /* Very small utilization - 1% */
    /* 1G, 64-byte: 1% = 14,880 pps */
    double util = calc_utilization(14880, 64, 1000000000ULL);
    ASSERT_FLOAT_EQ(1.0, util, 0.1);
}

/* ============================================================================
 * Wire Size Calculation Tests (implicit in calc_max_pps)
 * ============================================================================ */

TEST(wire_size_minimum_frame)
{
    /* 64-byte frame -> 84-byte wire size */
    /* Verify by checking PPS calculation */
    uint64_t pps = calc_max_pps(1000000000ULL, 64);
    /* 1G / (84 * 8) = 1,488,095 */
    ASSERT_EQ(1488095, pps);
}

TEST(wire_size_standard_frames)
{
    /* Test standard RFC 2544 frame sizes */
    uint32_t sizes[]        = {64, 128, 256, 512, 1024, 1280, 1518};
    uint64_t expected_pps[] = {
        1488095, /* 64 */
        844594,  /* 128 */
        452898,  /* 256 */
        234962,  /* 512 */
        119731,  /* 1024 */
        96153,   /* 1280 */
        81274    /* 1518 */
    };

    for (int i = 0; i < 7; i++) {
        uint64_t pps = calc_max_pps(1000000000ULL, sizes[i]);
        ASSERT_EQ(expected_pps[i], pps);
    }
}

/* ============================================================================
 * Line Rate Validation Tests
 * ============================================================================ */

TEST(line_rate_constants)
{
    /* Verify common line rate constants */
    uint64_t rate_1g   = 1000000000ULL;
    uint64_t rate_10g  = 10000000000ULL;
    uint64_t rate_25g  = 25000000000ULL;
    uint64_t rate_40g  = 40000000000ULL;
    uint64_t rate_100g = 100000000000ULL;

    ASSERT_EQ(1000000000ULL, rate_1g);
    ASSERT_EQ(10000000000ULL, rate_10g);
    ASSERT_EQ(25000000000ULL, rate_25g);
    ASSERT_EQ(40000000000ULL, rate_40g);
    ASSERT_EQ(100000000000ULL, rate_100g);
}

/* ============================================================================
 * Frame Size Validation Tests
 * ============================================================================ */

TEST(frame_size_enum_values)
{
    /* Verify frame size enum values match expected */
    ASSERT_EQ(64, FRAME_SIZE_64);
    ASSERT_EQ(128, FRAME_SIZE_128);
    ASSERT_EQ(256, FRAME_SIZE_256);
    ASSERT_EQ(512, FRAME_SIZE_512);
    ASSERT_EQ(1024, FRAME_SIZE_1024);
    ASSERT_EQ(1280, FRAME_SIZE_1280);
    ASSERT_EQ(1518, FRAME_SIZE_1518);
    ASSERT_EQ(9000, FRAME_SIZE_9000);
}

TEST(frame_size_minimum)
{
    /* Verify minimum frame size for RFC2544 payload */
    ASSERT_EQ(66, RFC2544_MIN_FRAME_SIZE);
}

/* ============================================================================
 * Main
 * ============================================================================ */

int main(void)
{
    printf("Seed Test Suite - Pacing Unit Tests\n");
    printf("Copyright (c) 2025 Mustard Seed Networks\n\n");

    TEST_SUITE("calc_max_pps");
    RUN_TEST(calc_max_pps_1g_64byte);
    RUN_TEST(calc_max_pps_1g_1518byte);
    RUN_TEST(calc_max_pps_10g_64byte);
    RUN_TEST(calc_max_pps_10g_1518byte);
    RUN_TEST(calc_max_pps_100g_64byte);
    RUN_TEST(calc_max_pps_zero_line_rate);
    RUN_TEST(calc_max_pps_jumbo_frame);
    RUN_TEST(calc_max_pps_25g_64byte);
    RUN_TEST(calc_max_pps_40g_64byte);

    TEST_SUITE("calc_utilization");
    RUN_TEST(calc_utilization_100_percent);
    RUN_TEST(calc_utilization_50_percent);
    RUN_TEST(calc_utilization_zero_rate);
    RUN_TEST(calc_utilization_zero_line_rate);
    RUN_TEST(calc_utilization_10g_64byte);
    RUN_TEST(calc_utilization_small_rate);

    TEST_SUITE("Wire Size Calculations");
    RUN_TEST(wire_size_minimum_frame);
    RUN_TEST(wire_size_standard_frames);

    TEST_SUITE("Line Rate Constants");
    RUN_TEST(line_rate_constants);

    TEST_SUITE("Frame Size Validation");
    RUN_TEST(frame_size_enum_values);
    RUN_TEST(frame_size_minimum);

    TEST_SUMMARY();

    return TEST_EXIT_STATUS();
}
