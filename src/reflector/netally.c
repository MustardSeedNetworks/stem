// SPDX-License-Identifier: BUSL-1.1

#include "reflector.h"

static bool has_complete_ipv4_header(const uint8_t *data, uint32_t len)
{
    if (len < ETH_HDR_LEN + IP_HDR_MIN_LEN || data[ETH_HDR_LEN] >> 4 != 4) {
        return false;
    }
    uint32_t ip_header_len = (data[ETH_HDR_LEN] & 0x0f) * 4;
    return ip_header_len >= IP_HDR_MIN_LEN && len >= ETH_HDR_LEN + ip_header_len;
}

void reflect_netally_packet(uint8_t *data, uint32_t len, reflect_mode_t mode)
{
    if (mode != REFLECT_MODE_MAC && !has_complete_ipv4_header(data, len)) {
        return;
    }

    if (mode != REFLECT_MODE_MAC) {
        data[ETH_HDR_LEN + 1] ^= ITO_TOS_WIGGLE;
    }
    reflect_packet_with_mode(data, len, mode, true);
}
