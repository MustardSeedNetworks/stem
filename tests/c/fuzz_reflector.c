/*
 * fuzz_reflector.c - libFuzzer harness for the reflector's frame path.
 *
 * ADR 0008's fuzzing was scoped to the dataplane's RFC 2544 / Y.1564 parsers.
 * The reflector reads a second set of attacker-controlled bytes off the wire
 * and, unlike those parsers, it writes back into the same buffer: a frame that
 * classifies as reflectable is reflected in place. Its 2026-09-14 amendment
 * (stem#1228, plan row STM-22) widened the gate to cover it.
 *
 *     make c-fuzz-reflector          # bounded run in CI / locally
 *
 * The harness is `worker_loop`'s per-frame sequence (src/reflector/core.c):
 * classify with `is_ito_packet`, take the signature type, then reflect through
 * the branch that type selects. The frame is copied into an exactly-sized heap
 * allocation, so a read or a write one byte outside the frame the classifier
 * accepted is an ASAN finding rather than a silent overwrite of whatever the
 * ring buffer held next.
 *
 * The first byte of each input picks the configuration rather than pinning one:
 * the reflect mode and the signature filter both change which bytes are read
 * and written, and a harness that fuzzed a single mode would leave the others
 * uncovered.
 *
 * Copyright (c) 2026 Mustard Seed Networks. All rights reserved.
 */

#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "../../include/reflector.h"

int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size)
{
    if (size < 2 || size > 0xFFFF) {
        return 0; /* one control byte plus a frame, kept frame-sized */
    }

    const uint8_t  control = data[0];
    const uint8_t *frame   = data + 1;
    const uint32_t len     = (uint32_t)(size - 1);

    reflector_config_t config = {0};
    config.ito_port           = 3842;
    config.reflect_mode       = (reflect_mode_t)(control % 3);
    config.sig_filter         = (sig_filter_t)((control >> 2) % 9);
    config.software_checksum  = (control & 0x40) != 0;
    config.filter_dst_mac     = (control & 0x80) != 0;
    config.enable_ipv6        = true;
    config.enable_vlan        = true;
    /* With filter_dst_mac set the frame must carry this MAC to be accepted;
     * the fuzzer finds it, and the unfiltered half of the corpus covers the
     * path meanwhile. */
    memcpy(config.mac, "\x02\x00\x00\x00\x00\x01", 6);

    /* Exactly sized: the reflect functions write in place, so an overrun in
     * either direction lands in ASAN's redzone rather than in a neighbour. */
    uint8_t *buf = malloc(len);
    if (!buf) {
        return 0;
    }
    memcpy(buf, frame, len);

    if (is_ito_packet(buf, len, &config)) {
        ito_sig_type_t sig_type = get_ito_signature_type(buf, len);

        if (sig_type == ITO_SIG_TYPE_PROBEOT || sig_type == ITO_SIG_TYPE_DATAOT ||
            sig_type == ITO_SIG_TYPE_LATENCY) {
            reflect_netally_packet(buf, len, config.reflect_mode);
        } else {
            reflect_packet_with_mode(buf, len, config.reflect_mode, config.software_checksum);
        }
    }

    /* The VLAN and IPv6 helpers are reached from the same frame on the
     * configurations that enable them; drive them directly too, since the
     * classifier rejects most inputs before they are consulted. */
    uint16_t inner_ethertype = 0;
    uint32_t vlan_offset     = 0;
    (void)is_vlan_tagged(buf, len, &inner_ethertype, &vlan_offset);

    bool is_ipv6 = false;
    bool is_vlan = false;
    (void)is_ito_packet_extended(buf, len, &config, &is_ipv6, &is_vlan);

    free(buf);
    return 0;
}
