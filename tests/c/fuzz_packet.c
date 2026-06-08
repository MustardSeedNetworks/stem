/*
 * fuzz_packet.c - libFuzzer harness for the RFC2544 / Y.1564 frame parsers.
 *
 * The dataplane parsers are the one place attacker-controlled bytes meet C, so
 * they get continuous fuzzing under AddressSanitizer. Build + run with:
 *
 *     make c-fuzz            # bounded run in CI / locally
 *     clang -std=c23 -g -O1 -fsanitize=fuzzer,address -Iinclude \
 *         -o fuzz_packet tests/c/fuzz_packet.c src/dataplane/common/packet.c
 *     ./fuzz_packet -max_total_time=60
 *
 * The harness models a real consumer: when a validator accepts a frame, it
 * reads the full 24-byte payload the validator vouched for. If a parser ever
 * accepts a frame too short to hold that payload, ASAN trips on the read.
 *
 * Copyright (c) 2026 Mustard Seed Networks. All rights reserved.
 */

#include <stddef.h>
#include <stdint.h>

#include "../../include/rfc2544.h"

/* Implemented in src/dataplane/common/packet.c. */
bool     rfc2544_is_valid_response(const uint8_t *data, uint32_t len);
bool     custom_is_valid_response(const uint8_t *data, uint32_t len, const char *signature);
bool     y1564_is_valid_response(const uint8_t *data, uint32_t len);
uint32_t rfc2544_get_seq_num(const uint8_t *data, uint32_t len);
uint64_t rfc2544_get_tx_timestamp(const uint8_t *data, uint32_t len);
uint32_t y1564_get_seq_num(const uint8_t *data, uint32_t len);
uint32_t y1564_get_service_id(const uint8_t *data, uint32_t len);

#define HDR_LEN     42
#define PAYLOAD_LEN 24

static void read_full_payload(const uint8_t *data)
{
    volatile uint8_t sink = 0;
    for (int i = HDR_LEN; i < HDR_LEN + PAYLOAD_LEN; i++) {
        sink ^= data[i];
    }
    (void)sink;
}

int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size)
{
    if (size > 0xFFFF) {
        return 0; /* keep inputs frame-sized */
    }
    uint32_t len = (uint32_t)size;

    if (rfc2544_is_valid_response(data, len)) {
        read_full_payload(data);
    }
    if (custom_is_valid_response(data, len, RFC2544_SIGNATURE)) {
        read_full_payload(data);
    }
    if (y1564_is_valid_response(data, len)) {
        read_full_payload(data);
    }

    /* Extractors validate internally; fuzz them directly too. */
    (void)rfc2544_get_seq_num(data, len);
    (void)rfc2544_get_tx_timestamp(data, len);
    (void)y1564_get_seq_num(data, len);
    (void)y1564_get_service_id(data, len);
    return 0;
}
