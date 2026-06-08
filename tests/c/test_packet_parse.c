/*
 * test_packet_parse.c - Unit tests for the RFC2544 / Y.1564 frame validators.
 *
 * Regression guard for the parse-length bug: the *_is_valid_response functions
 * used to accept a 64-/65-byte frame (the Ethernet minimum) even though the
 * 24-byte payload they vouch for runs to offset 66, leaving a 1-2 byte
 * out-of-bounds read for any consumer that reads the payload. The guard now
 * requires the full header+payload length (66). Run this binary under
 * AddressSanitizer (`make c-test-asan`) so any regression to a short bound
 * trips ASAN on the full-payload read below as well as the assertions.
 *
 * Copyright (c) 2025-2026 Mustard Seed Networks. All rights reserved.
 */

#include <stdint.h>
#include <string.h>

#include "../../include/rfc2544.h"
#include "test_framework.h"

/* Implemented in src/dataplane/common/packet.c (not exported via the header). */
bool     rfc2544_is_valid_response(const uint8_t *data, uint32_t len);
bool     custom_is_valid_response(const uint8_t *data, uint32_t len, const char *signature);
bool     y1564_is_valid_response(const uint8_t *data, uint32_t len);
uint32_t rfc2544_get_seq_num(const uint8_t *data, uint32_t len);
uint64_t rfc2544_get_tx_timestamp(const uint8_t *data, uint32_t len);

/* Frame geometry: Ethernet 14 + IPv4 20 + UDP 8 = 42, payload 24 → full 66. */
#define HDR_LEN     42
#define PAYLOAD_LEN 24
#define FULL_LEN    (HDR_LEN + PAYLOAD_LEN)

/* Build a frame of exactly `len` bytes into `buf` with the given signature at
 * the payload offset (so the signature memcmp would pass if the guard let it). */
static void build_frame(uint8_t *buf, uint32_t len, const char *sig)
{
    memset(buf, 0, len);
    if (len >= HDR_LEN + RFC2544_SIG_LEN) {
        memcpy(buf + HDR_LEN, sig, RFC2544_SIG_LEN);
    }
}

/* A consumer's full-payload read; under ASAN this faults if a too-short frame
 * was wrongly accepted. */
static uint8_t consume_full_payload(const uint8_t *buf)
{
    volatile uint8_t sink = 0;
    for (int i = HDR_LEN; i < FULL_LEN; i++) {
        sink ^= buf[i];
    }
    return sink;
}

TEST(rfc2544_rejects_64_byte_frame)
{
    uint8_t buf[64];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_FALSE(rfc2544_is_valid_response(buf, 64));
}

TEST(rfc2544_rejects_65_byte_frame)
{
    uint8_t buf[65];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_FALSE(rfc2544_is_valid_response(buf, 65));
}

TEST(rfc2544_accepts_66_byte_frame_and_reads_in_bounds)
{
    uint8_t buf[FULL_LEN];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_TRUE(rfc2544_is_valid_response(buf, FULL_LEN));
    /* Reading the whole payload of an accepted frame must stay in bounds. */
    (void)consume_full_payload(buf);
}

TEST(custom_rejects_short_frames)
{
    uint8_t buf64[64];
    uint8_t buf65[65];
    build_frame(buf64, sizeof(buf64), RFC2544_SIGNATURE);
    build_frame(buf65, sizeof(buf65), RFC2544_SIGNATURE);
    ASSERT_FALSE(custom_is_valid_response(buf64, 64, RFC2544_SIGNATURE));
    ASSERT_FALSE(custom_is_valid_response(buf65, 65, RFC2544_SIGNATURE));
}

TEST(custom_accepts_full_frame)
{
    uint8_t buf[FULL_LEN];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_TRUE(custom_is_valid_response(buf, FULL_LEN, RFC2544_SIGNATURE));
}

TEST(y1564_rejects_short_frames)
{
    uint8_t buf64[64];
    uint8_t buf65[65];
    build_frame(buf64, sizeof(buf64), Y1564_SIGNATURE);
    build_frame(buf65, sizeof(buf65), Y1564_SIGNATURE);
    ASSERT_FALSE(y1564_is_valid_response(buf64, 64));
    ASSERT_FALSE(y1564_is_valid_response(buf65, 65));
}

TEST(y1564_accepts_full_frame)
{
    uint8_t buf[FULL_LEN];
    build_frame(buf, sizeof(buf), Y1564_SIGNATURE);
    ASSERT_TRUE(y1564_is_valid_response(buf, FULL_LEN));
}

TEST(null_and_zero_length_are_rejected)
{
    ASSERT_FALSE(rfc2544_is_valid_response(NULL, FULL_LEN));
    uint8_t buf[FULL_LEN];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_FALSE(rfc2544_is_valid_response(buf, 0));
}

TEST(extractors_safe_on_short_frames)
{
    /* Extractors validate first; a short frame must yield 0, never an OOB read. */
    uint8_t buf[64];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE);
    ASSERT_EQ((uint32_t)0, rfc2544_get_seq_num(buf, 64));
    ASSERT_EQ((uint64_t)0, rfc2544_get_tx_timestamp(buf, 64));
}

int main(void)
{
    TEST_SUITE("RFC2544 / Y.1564 frame validators");
    RUN_TEST(rfc2544_rejects_64_byte_frame);
    RUN_TEST(rfc2544_rejects_65_byte_frame);
    RUN_TEST(rfc2544_accepts_66_byte_frame_and_reads_in_bounds);
    RUN_TEST(custom_rejects_short_frames);
    RUN_TEST(custom_accepts_full_frame);
    RUN_TEST(y1564_rejects_short_frames);
    RUN_TEST(y1564_accepts_full_frame);
    RUN_TEST(null_and_zero_length_are_rejected);
    RUN_TEST(extractors_safe_on_short_frames);
    TEST_SUMMARY();
}
