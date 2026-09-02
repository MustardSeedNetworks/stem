/**
 * @file platform_stub.h
 * @brief Stub definitions for non-Linux platforms
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 *
 * Provides type stubs so common dataplane code (pacing, protocol logic,
 * IMIX patterns, etc.) can compile on macOS for development and testing.
 * Network I/O backends (AF_PACKET, AF_XDP) remain Linux-only.
 *
 * Usage: compile with -DSTUB_PLATFORM to activate stubs.
 */

#ifndef PLATFORM_STUB_H
#define PLATFORM_STUB_H

#ifdef STUB_PLATFORM

#include <stdint.h>

/* Stub platform operations - no-op implementations for non-Linux */
typedef struct platform_ops {
    int (*init)(void *ctx, const char *interface);
    int (*send)(void *ctx, const uint8_t *pkt, uint32_t len);
    int (*recv)(void *ctx, uint8_t *buf, uint32_t buf_len);
    void (*cleanup)(void *ctx);
} platform_ops_t;

/* Ethernet protocol constants (normally from linux/if_ether.h) */
#ifndef ETH_P_IP
#define ETH_P_IP 0x0800
#endif

#ifndef ETH_P_IPV6
#define ETH_P_IPV6 0x86DD
#endif

#ifndef ETH_ALEN
#define ETH_ALEN 6
#endif

/* IOCTL constants (normally from linux/sockios.h) */
#ifndef SIOCETHTOOL
#define SIOCETHTOOL 0x8946
#endif

#endif /* STUB_PLATFORM */

#endif /* PLATFORM_STUB_H */
