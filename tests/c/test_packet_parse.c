/*
 * test_packet_parse.c - Unit tests for the RFC2544 / Y.1564 frame validators.
 *
 * Regression guard for the parse-length bug: the *_is_valid_response functions
 * must never accept a frame shorter than the payload they vouch for. The
 * compact RFC 2544 measurement payload fits safely in the standard 64-byte
 * frame while Y.1564 retains its larger protocol-specific payload. Run this binary under
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
void *rfc2544_create_packet_template(uint8_t *buffer, uint32_t frame_size, const uint8_t *src_mac,
                                     const uint8_t *dst_mac, uint32_t src_ip, uint32_t dst_ip,
                                     uint16_t src_port, uint16_t dst_port, uint32_t stream_id);
void  rfc2544_stamp_packet(void *payload, uint32_t seq_num, uint64_t timestamp_ns);

/* Frame geometry: Ethernet 14 + IPv4 20 + UDP 8 = 42. */
#define HDR_LEN             42
#define RFC2544_PAYLOAD_LEN 18
#define RFC2544_PACKET_LEN  (HDR_LEN + RFC2544_PAYLOAD_LEN)
#define CUSTOM_FULL_LEN     66
#define Y1564_FULL_LEN      66

/* Build a frame of exactly `len` bytes into `buf` with the given signature at
 * the payload offset (so the signature memcmp would pass if the guard let it). */
static void build_frame(uint8_t *buf, uint32_t len, const char *sig, size_t sig_len)
{
    memset(buf, 0, len);
    if (len >= HDR_LEN + sig_len) {
        memcpy(buf + HDR_LEN, sig, sig_len);
    }
}

/* A consumer's full-payload read; under ASAN this faults if a too-short frame
 * was wrongly accepted. */
static uint8_t consume_full_payload(const uint8_t *buf)
{
    volatile uint8_t sink = 0;
    for (int i = HDR_LEN; i < RFC2544_PACKET_LEN; i++) {
        sink ^= buf[i];
    }
    return sink;
}

TEST(rfc2544_rejects_frame_shorter_than_standard_minimum)
{
    uint8_t buf[RFC2544_PACKET_LEN - 1];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE, RFC2544_SIG_LEN);
    ASSERT_FALSE(rfc2544_is_valid_response(buf, sizeof(buf)));
}

TEST(rfc2544_accepts_64_byte_frame_and_reads_in_bounds)
{
    uint8_t buf[RFC2544_PACKET_LEN];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE, RFC2544_SIG_LEN);
    ASSERT_TRUE(rfc2544_is_valid_response(buf, sizeof(buf)));
    /* Reading the whole payload of an accepted frame must stay in bounds. */
    (void)consume_full_payload(buf);
}

TEST(rfc2544_64_byte_template_preserves_measurements)
{
    uint8_t        buf[RFC2544_PACKET_LEN];
    const uint8_t  mac[6]    = {0, 1, 2, 3, 4, 5};
    const uint32_t sequence  = 0x10203040;
    const uint64_t timestamp = 0x0102030405060708ULL;

    void *payload =
        rfc2544_create_packet_template(buf, sizeof(buf), mac, mac, 0, 0, 12345, 3842, 7);
    ASSERT_TRUE(payload != NULL);
    rfc2544_stamp_packet(payload, sequence, timestamp);

    ASSERT_EQ(sequence, rfc2544_get_seq_num(buf, sizeof(buf)));
    ASSERT_EQ(timestamp, rfc2544_get_tx_timestamp(buf, sizeof(buf)));
}

TEST(custom_rejects_short_frames)
{
    uint8_t buf[CUSTOM_FULL_LEN - 1];
    build_frame(buf, sizeof(buf), "RFC254 ", 7);
    ASSERT_FALSE(custom_is_valid_response(buf, sizeof(buf), RFC2544_SIGNATURE));
}

TEST(custom_accepts_full_frame)
{
    uint8_t buf[CUSTOM_FULL_LEN];
    build_frame(buf, sizeof(buf), "RFC254 ", 7);
    ASSERT_TRUE(custom_is_valid_response(buf, sizeof(buf), RFC2544_SIGNATURE));
}

TEST(y1564_rejects_short_frames)
{
    uint8_t buf[Y1564_FULL_LEN - 1];
    build_frame(buf, sizeof(buf), Y1564_SIGNATURE, Y1564_SIG_LEN);
    ASSERT_FALSE(y1564_is_valid_response(buf, sizeof(buf)));
}

TEST(y1564_accepts_full_frame)
{
    uint8_t buf[Y1564_FULL_LEN];
    build_frame(buf, sizeof(buf), Y1564_SIGNATURE, Y1564_SIG_LEN);
    ASSERT_TRUE(y1564_is_valid_response(buf, sizeof(buf)));
}

TEST(null_and_zero_length_are_rejected)
{
    ASSERT_FALSE(rfc2544_is_valid_response(NULL, RFC2544_PACKET_LEN));
    uint8_t buf[RFC2544_PACKET_LEN];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE, RFC2544_SIG_LEN);
    ASSERT_FALSE(rfc2544_is_valid_response(buf, 0));
}

TEST(extractors_safe_on_short_frames)
{
    /* Extractors validate first; a short frame must yield 0, never an OOB read. */
    uint8_t buf[RFC2544_PACKET_LEN - 1];
    build_frame(buf, sizeof(buf), RFC2544_SIGNATURE, RFC2544_SIG_LEN);
    ASSERT_EQ((uint32_t)0, rfc2544_get_seq_num(buf, sizeof(buf)));
    ASSERT_EQ((uint64_t)0, rfc2544_get_tx_timestamp(buf, sizeof(buf)));
}

int main(void)
{
    TEST_SUITE("RFC2544 / Y.1564 frame validators");
    RUN_TEST(rfc2544_rejects_frame_shorter_than_standard_minimum);
    RUN_TEST(rfc2544_accepts_64_byte_frame_and_reads_in_bounds);
    RUN_TEST(rfc2544_64_byte_template_preserves_measurements);
    RUN_TEST(custom_rejects_short_frames);
    RUN_TEST(custom_accepts_full_frame);
    RUN_TEST(y1564_rejects_short_frames);
    RUN_TEST(y1564_accepts_full_frame);
    RUN_TEST(null_and_zero_length_are_rejected);
    RUN_TEST(extractors_safe_on_short_frames);
    TEST_SUMMARY();

    return TEST_EXIT_STATUS();
}
