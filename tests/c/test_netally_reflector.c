// SPDX-License-Identifier: BUSL-1.1

#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "reflector.h"

#if !defined(PACKET_BLOCK_TIMEOUT_MS)
#error "PACKET_BLOCK_TIMEOUT_MS must be defined"
#endif

static_assert(PACKET_BLOCK_TIMEOUT_MS == 1, "reflector RX blocks must retire promptly");

#define FRAME_LEN   66
#define IPV4_OFFSET 14
#define UDP_OFFSET  34

static int failures = 0;

static void check_bytes(const char *name, const uint8_t *got, const uint8_t *want, size_t len)
{
    if (memcmp(got, want, len) != 0) {
        fprintf(stderr, "FAIL: %s\n", name);
        failures++;
    }
}

static void build_probe(uint8_t frame[FRAME_LEN])
{
    const uint8_t reflector_mac[6] = {0x02, 0, 0, 0, 0, 1};
    const uint8_t tester_mac[6]    = {0x00, 0xc0, 0x17, 0x57, 0x01, 0x7c};
    const uint8_t tester_ip[4]     = {10, 44, 10, 183};
    const uint8_t reflector_ip[4]  = {10, 44, 40, 23};

    memset(frame, 0, FRAME_LEN);
    memcpy(frame, reflector_mac, sizeof(reflector_mac));
    memcpy(frame + 6, tester_mac, sizeof(tester_mac));
    frame[12]              = 0x08;
    frame[13]              = 0x00;
    frame[IPV4_OFFSET]     = 0x45;
    frame[IPV4_OFFSET + 1] = 0x20;
    frame[IPV4_OFFSET + 9] = 17;
    memcpy(frame + IPV4_OFFSET + 12, tester_ip, sizeof(tester_ip));
    memcpy(frame + IPV4_OFFSET + 16, reflector_ip, sizeof(reflector_ip));
    frame[UDP_OFFSET]     = 0xc3;
    frame[UDP_OFFSET + 1] = 0x50;
    frame[UDP_OFFSET + 2] = 0x0f;
    frame[UDP_OFFSET + 3] = 0x02;
    memcpy(frame + UDP_OFFSET + UDP_HDR_LEN + ITO_SIG_OFFSET, "PROBEOT", ITO_SIG_LEN);
}

static void test_netally_handshake_reflection(void)
{
    uint8_t frame[FRAME_LEN];
    uint8_t original[FRAME_LEN];

    build_probe(frame);
    memcpy(original, frame, sizeof(original));
    reflect_netally_packet(frame, sizeof(frame), REFLECT_MODE_MAC_IP);

    check_bytes("destination MAC", frame, original + 6, 6);
    check_bytes("source MAC", frame + 6, original, 6);
    check_bytes("source IP", frame + IPV4_OFFSET + 12, original + IPV4_OFFSET + 16, 4);
    check_bytes("destination IP", frame + IPV4_OFFSET + 16, original + IPV4_OFFSET + 12, 4);
    check_bytes("UDP ports remain unchanged", frame + UDP_OFFSET, original + UDP_OFFSET, 4);
    if (frame[IPV4_OFFSET + 1] != (uint8_t)(original[IPV4_OFFSET + 1] ^ 0x01)) {
        fprintf(stderr, "FAIL: IPv4 ToS handshake bit was not toggled\n");
        failures++;
    }
}

static void test_netally_honors_all_mode(void)
{
    uint8_t frame[FRAME_LEN];
    uint8_t original[FRAME_LEN];

    build_probe(frame);
    memcpy(original, frame, sizeof(original));
    reflect_netally_packet(frame, sizeof(frame), REFLECT_MODE_ALL);

    check_bytes("source UDP port", frame + UDP_OFFSET, original + UDP_OFFSET + 2, 2);
    check_bytes("destination UDP port", frame + UDP_OFFSET + 2, original + UDP_OFFSET, 2);
}

static void test_netally_mac_mode_does_not_invalidate_ip_checksum(void)
{
    uint8_t frame[FRAME_LEN];
    uint8_t original[FRAME_LEN];

    build_probe(frame);
    memcpy(original, frame, sizeof(original));
    reflect_netally_packet(frame, sizeof(frame), REFLECT_MODE_MAC);

    if (frame[IPV4_OFFSET + 1] != original[IPV4_OFFSET + 1]) {
        fprintf(stderr, "FAIL: MAC-only reflection changed IPv4 ToS without updating checksum\n");
        failures++;
    }
}

static void test_netally_rejects_truncated_ipv4_options(void)
{
    uint8_t frame[FRAME_LEN];

    build_probe(frame);
    frame[IPV4_OFFSET] = 0x4f;

    reflector_config_t config = {0};
    config.ito_port           = ITO_UDP_PORT;
    config.sig_filter         = SIG_FILTER_ITO;

    if (is_ito_packet(frame, sizeof(frame), &config)) {
        fprintf(stderr, "FAIL: truncated IPv4 options packet was accepted\n");
        failures++;
    }

    reflect_netally_packet(frame, sizeof(frame), REFLECT_MODE_MAC_IP);
}

static void test_individual_ito_signature_filters(void)
{
    uint8_t frame[FRAME_LEN];
    build_probe(frame);

    reflector_config_t config = {0};
    config.ito_port           = ITO_UDP_PORT;
    config.sig_filter         = SIG_FILTER_PROBEOT;
    if (!is_ito_packet(frame, sizeof(frame), &config)) {
        fprintf(stderr, "FAIL: PROBEOT filter rejected PROBEOT packet\n");
        failures++;
    }

    config.sig_filter = SIG_FILTER_DATAOT;
    if (is_ito_packet(frame, sizeof(frame), &config)) {
        fprintf(stderr, "FAIL: DATA:OT filter accepted PROBEOT packet\n");
        failures++;
    }
}

int main(void)
{
    test_netally_handshake_reflection();
    test_netally_honors_all_mode();
    test_netally_mac_mode_does_not_invalidate_ip_checksum();
    test_netally_rejects_truncated_ipv4_options();
    test_individual_ito_signature_filters();
    if (failures != 0) {
        return 1;
    }
    puts("NetAlly reflector compatibility tests passed");
    return 0;
}
