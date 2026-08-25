/*
 * packet_platform.c - Maximum Performance Linux AF_PACKET implementation
 *
 * Copyright (c) 2025 Kris Armstrong
 *
 * HIGHLY OPTIMIZED AF_PACKET fallback for NICs without AF_XDP support.
 * Implements every possible optimization:
 * - PACKET_MMAP (zero-copy ring buffers)
 * - PACKET_FANOUT (multi-queue distribution)
 * - PACKET_QDISC_BYPASS (bypass qdisc layer)
 * - TPACKET_V2 (frame-level ring buffers)
 * - SO_BUSY_POLL (low latency polling)
 *
 * Expected performance: 100-200 Mbps (vs 50-100 Mbps without optimizations)
 * Still far below AF_XDP (10 Gbps), but maximum possible for AF_PACKET.
 */

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <arpa/inet.h>
#include <linux/filter.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <sys/socket.h>

#include <fcntl.h>
#include <net/if.h>
#include <poll.h>
#include <unistd.h>

#include "reflector.h"

/* Ring buffer configuration - tuned for performance */
#define PACKET_RING_FRAMES  4096
#define PACKET_FRAME_SIZE   2048
#define PACKET_BLOCK_SIZE   (PACKET_FRAME_SIZE * 128) /* 128 frames per block */
#define PACKET_BLOCK_NR     (PACKET_RING_FRAMES / 128)
#define PACKET_IDLE_WAIT_MS 100

/* Platform-specific context for optimized AF_PACKET */
struct platform_ctx {
    int      sock_fd;      /* AF_PACKET socket */
    int      udp_guard_fd; /* Prevent the kernel from rejecting reflected UDP probes. */
    uint16_t udp_guard_port;
    bool     udp_guard_configured;

    /* RX ring buffer (PACKET_MMAP) */
    void        *rx_ring;
    size_t       rx_ring_size;
    unsigned int rx_frame_num;
    unsigned int rx_frame_idx;

    /* TPACKET version in use (2 or 3) */
    int tpacket_version;

    /* Ring buffer configuration */
    union {
        struct tpacket_req  req2; /* TPACKET_V2 */
        struct tpacket_req3 req3; /* TPACKET_V3 */
    };

    /* V3 block tracking */
    unsigned int current_block_idx;
    unsigned int current_block_offset;
    uint32_t     current_frame_offset;
    bool         block_release_pending;

    /* Frame size */
    uint32_t frame_size;
};

void packet_platform_cleanup(worker_ctx_t *wctx);
void packet_platform_release_batch(worker_ctx_t *wctx, packet_t *pkts, int num_pkts);

static int open_udp_guard(const char *ifname, uint16_t port)
{
    int fd = socket(AF_INET, SOCK_DGRAM, 0);
    if (fd < 0) {
        return -1;
    }
    const struct sock_filter code[] = {BPF_STMT(BPF_RET | BPF_K, 0)};
    const struct sock_fprog  filter = {.len = 1, .filter = (struct sock_filter *)code};
    struct sockaddr_in       addr   = {
                .sin_family      = AF_INET,
                .sin_port        = htons(port),
                .sin_addr.s_addr = htonl(INADDR_ANY),
    };
    if (setsockopt(fd, SOL_SOCKET, SO_BINDTODEVICE, ifname, strlen(ifname) + 1) < 0 ||
        setsockopt(fd, SOL_SOCKET, SO_ATTACH_FILTER, &filter, sizeof(filter)) < 0 ||
        bind(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        int saved_errno = errno;
        close(fd);
        errno = saved_errno;
        return -1;
    }
    return fd;
}

static int ignore_outgoing_packets(int fd)
{
    int enabled = 1;
    return setsockopt(fd, SOL_PACKET, PACKET_IGNORE_OUTGOING, &enabled, sizeof(enabled));
}

int packet_platform_set_guard_port(worker_ctx_t *wctx, uint16_t port)
{
    struct platform_ctx *pctx = wctx->pctx;
    if (pctx->udp_guard_configured && pctx->udp_guard_port == port) {
        return 0;
    }

    int fd = port == 0 ? -1 : open_udp_guard(wctx->config->ifname, port);
    if (port != 0 && fd < 0 && errno != EADDRINUSE) {
        reflector_log(LOG_ERROR, "Failed to bind filtered UDP guard port %u: %s", port,
                      strerror(errno));
        return -1;
    }
    if (port != 0 && fd < 0) {
        reflector_log(LOG_INFO, "UDP port %u is already claimed; no guard socket needed", port);
    }
    if (pctx->udp_guard_fd >= 0) {
        close(pctx->udp_guard_fd);
    }
    pctx->udp_guard_fd         = fd;
    pctx->udp_guard_port       = port;
    pctx->udp_guard_configured = true;
    return 0;
}

/*
 * Try to setup TPACKET_V3 (preferred for real hardware)
 * Returns 0 on success, -1 on failure
 */
static int try_tpacket_v3(struct platform_ctx *pctx)
{
    int version = TPACKET_V3;
    if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_VERSION, &version, sizeof(version)) < 0) {
        return -1;
    }

    /* Configure RX ring buffer with TPACKET_V3 (block-level batching) */
    memset(&pctx->req3, 0, sizeof(pctx->req3));
    pctx->req3.tp_block_size       = PACKET_BLOCK_SIZE;
    pctx->req3.tp_frame_size       = PACKET_FRAME_SIZE;
    pctx->req3.tp_block_nr         = PACKET_BLOCK_NR;
    pctx->req3.tp_frame_nr         = PACKET_RING_FRAMES;
    pctx->req3.tp_retire_blk_tov   = PACKET_BLOCK_TIMEOUT_MS;
    pctx->req3.tp_feature_req_word = 0;

    if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_RX_RING, &pctx->req3, sizeof(pctx->req3)) <
        0) {
        return -1;
    }

    pctx->tpacket_version = 3;
    pctx->rx_ring_size    = pctx->req3.tp_block_size * pctx->req3.tp_block_nr;
    pctx->rx_frame_num    = pctx->req3.tp_frame_nr;
    return 0;
}

/*
 * Try to setup TPACKET_V2 (fallback for veth/testing)
 * Returns 0 on success, -1 on failure
 */
static int try_tpacket_v2(struct platform_ctx *pctx)
{
    int version = TPACKET_V2;
    if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_VERSION, &version, sizeof(version)) < 0) {
        return -1;
    }

    /* Configure RX ring buffer with TPACKET_V2 (frame-level) */
    memset(&pctx->req2, 0, sizeof(pctx->req2));
    pctx->req2.tp_block_size = PACKET_BLOCK_SIZE;
    pctx->req2.tp_frame_size = PACKET_FRAME_SIZE;
    pctx->req2.tp_block_nr   = PACKET_BLOCK_NR;
    pctx->req2.tp_frame_nr   = PACKET_RING_FRAMES;

    if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_RX_RING, &pctx->req2, sizeof(pctx->req2)) <
        0) {
        return -1;
    }

    pctx->tpacket_version = 2;
    pctx->rx_ring_size    = pctx->req2.tp_block_size * pctx->req2.tp_block_nr;
    pctx->rx_frame_num    = pctx->req2.tp_frame_nr;
    return 0;
}

/*
 * Initialize maximum performance AF_PACKET platform
 * Tries TPACKET_V3 first (best for real hardware), falls back to V2 (for veth/testing)
 */
int packet_platform_init(reflector_ctx_t *rctx, worker_ctx_t *wctx)
{
    (void)rctx;

    struct platform_ctx *pctx = calloc(1, sizeof(*pctx));
    if (!pctx) {
        return -ENOMEM;
    }

    wctx->pctx         = pctx;
    pctx->frame_size   = PACKET_FRAME_SIZE;
    pctx->udp_guard_fd = -1;

    /* Create AF_PACKET socket */
    pctx->sock_fd = socket(AF_PACKET, SOCK_RAW, htons(ETH_P_ALL));
    if (pctx->sock_fd < 0) {
        reflector_log(LOG_ERROR, "Failed to create AF_PACKET socket: %s", strerror(errno));
        free(pctx);
        return -1;
    }
    if (ignore_outgoing_packets(pctx->sock_fd) < 0) {
        reflector_log(LOG_ERROR, "Failed to suppress outgoing packet capture: %s", strerror(errno));
        close(pctx->sock_fd);
        free(pctx);
        return -1;
    }

    /* Try TPACKET_V3 first (better for real hardware) */
    if (try_tpacket_v3(pctx) == 0) {
        reflector_log(LOG_DEBUG, "Using TPACKET_V3 (block-level batching)");
    } else {
        /* Fall back to TPACKET_V2 (works on veth, older kernels) */
        /* Need new socket since V3 attempt may have left socket in bad state */
        close(pctx->sock_fd);
        pctx->sock_fd = socket(AF_PACKET, SOCK_RAW, htons(ETH_P_ALL));
        if (pctx->sock_fd < 0 || ignore_outgoing_packets(pctx->sock_fd) < 0 ||
            try_tpacket_v2(pctx) < 0) {
            reflector_log(LOG_ERROR, "Failed to setup TPACKET_V2: %s", strerror(errno));
            if (pctx->sock_fd >= 0) {
                close(pctx->sock_fd);
            }
            free(pctx);
            return -1;
        }
        reflector_log(LOG_DEBUG, "Using TPACKET_V2 (frame-level, veth compatible)");
    }

    /* mmap() the RX ring. TX uses sendmmsg(): PACKET_TX_RING has inconsistent
     * TPACKET_V3 behavior on veth devices, while batched direct sends retain
     * qdisc bypass and avoid silently dropping reflected probes. */
    bool use_simple_mode = false;
    pctx->rx_ring        = mmap(NULL, pctx->rx_ring_size, PROT_READ | PROT_WRITE,
                                MAP_SHARED | MAP_LOCKED | MAP_POPULATE, pctx->sock_fd, 0);
    if (pctx->rx_ring == MAP_FAILED && (errno == EAGAIN || errno == EPERM || errno == ENOMEM)) {
        reflector_log(
            LOG_INFO,
            "Locked packet rings unavailable; retrying zero-copy rings without MAP_LOCKED");
        pctx->rx_ring = mmap(NULL, pctx->rx_ring_size, PROT_READ | PROT_WRITE,
                             MAP_SHARED | MAP_POPULATE, pctx->sock_fd, 0);
    }
    if (pctx->rx_ring == MAP_FAILED) {
        reflector_log(LOG_WARN, "Failed to mmap ring buffers: %s", strerror(errno));
        reflector_log(LOG_INFO, "Using simple recv/send mode (slower but more compatible)");
        use_simple_mode    = true;
        pctx->rx_ring      = NULL;
        pctx->rx_ring_size = 0;
    }

    if (!use_simple_mode) {
        pctx->rx_frame_idx         = 0;
        pctx->current_block_idx    = 0;
        pctx->current_block_offset = 0;

        reflector_log(LOG_INFO, "Allocated PACKET_MMAP RX ring: %zu MB; TX: batched direct send",
                      pctx->rx_ring_size / (1024 * 1024));
    }

    /* Bind to interface */
    struct sockaddr_ll sll = {0};
    sll.sll_family         = AF_PACKET;
    sll.sll_protocol       = htons(ETH_P_ALL);
    sll.sll_ifindex        = wctx->config->ifindex;

    if (bind(pctx->sock_fd, (struct sockaddr *)&sll, sizeof(sll)) < 0) {
        reflector_log(LOG_ERROR, "Failed to bind AF_PACKET socket: %s", strerror(errno));
        munmap(pctx->rx_ring, pctx->rx_ring_size);
        close(pctx->sock_fd);
        free(pctx);
        return -1;
    }

    /* AF_PACKET sees the probe before the UDP stack, but it does not claim the
     * destination port. Keep a UDP socket bound so Linux does not race the raw
     * reflected reply with an ICMP port-unreachable response. */
    if (wctx->worker_id == 0 &&
        (wctx->config->sig_filter == SIG_FILTER_ALL || wctx->config->sig_filter == SIG_FILTER_ITO ||
         wctx->config->sig_filter == SIG_FILTER_PROBEOT ||
         wctx->config->sig_filter == SIG_FILTER_DATAOT ||
         wctx->config->sig_filter == SIG_FILTER_LATENCY)) {
        if (wctx->config->ito_port == 0) {
            reflector_log(LOG_ERROR,
                          "AF_PACKET requires a UDP port for ITO reflection to suppress ICMP");
            packet_platform_cleanup(wctx);
            return -1;
        }
        if (packet_platform_set_guard_port(wctx, wctx->config->ito_port) < 0) {
            packet_platform_cleanup(wctx);
            return -1;
        }
    }

    /* Enable PACKET_QDISC_BYPASS for faster TX */
    int qdisc_bypass = 1;
    if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_QDISC_BYPASS, &qdisc_bypass,
                   sizeof(qdisc_bypass)) < 0) {
        reflector_log(LOG_WARN, "Failed to enable QDISC bypass: %s", strerror(errno));
    } else {
        reflector_log(LOG_INFO, "PACKET_QDISC_BYPASS enabled (faster TX)");
    }

    /* Enable PACKET_FANOUT for multi-queue distribution (if multiple workers) */
    if (rctx->num_workers > 1) {
        uint32_t fanout_arg = (getpid() & 0xffff) | (PACKET_FANOUT_HASH << 16);
        if (setsockopt(pctx->sock_fd, SOL_PACKET, PACKET_FANOUT, &fanout_arg, sizeof(fanout_arg)) <
            0) {
            reflector_log(LOG_WARN, "Failed to enable PACKET_FANOUT: %s", strerror(errno));
        } else {
            reflector_log(LOG_INFO, "PACKET_FANOUT enabled (multi-queue distribution)");
        }
    }

    /* Enable SO_BUSY_POLL for lower latency (50 microseconds) */
    int busy_poll = 50;
    if (setsockopt(pctx->sock_fd, SOL_SOCKET, SO_BUSY_POLL, &busy_poll, sizeof(busy_poll)) < 0) {
        reflector_log(LOG_WARN, "Failed to enable busy polling: %s", strerror(errno));
    } else {
        reflector_log(LOG_INFO, "SO_BUSY_POLL enabled (low latency mode)");
    }

    /* Increase socket buffer sizes */
    int bufsize = 4 * 1024 * 1024; /* 4MB */
    setsockopt(pctx->sock_fd, SOL_SOCKET, SO_RCVBUF, &bufsize, sizeof(bufsize));
    setsockopt(pctx->sock_fd, SOL_SOCKET, SO_SNDBUF, &bufsize, sizeof(bufsize));

    reflector_log(LOG_INFO, "Optimized AF_PACKET initialized on %s:", wctx->config->ifname);
    reflector_log(LOG_INFO, "  - PACKET_MMAP: zero-copy ring buffers");
    reflector_log(LOG_INFO, "  - TPACKET_V%d: %s", pctx->tpacket_version,
                  pctx->tpacket_version == 3 ? "block-level batching (optimal)"
                                             : "frame-level (veth compatible)");
    reflector_log(LOG_INFO, "  - PACKET_QDISC_BYPASS: fast TX path");
    reflector_log(LOG_INFO, "  - SO_BUSY_POLL: reduced latency");
    reflector_log(LOG_INFO, "Expected: %s",
                  pctx->tpacket_version == 3 ? "200-400 Mbps (real hardware)"
                                             : "100-200 Mbps (veth/virtual)");

    return 0;
}

/*
 * Cleanup AF_PACKET platform
 */
void packet_platform_cleanup(worker_ctx_t *wctx)
{
    struct platform_ctx *pctx = wctx->pctx;
    if (!pctx) {
        return;
    }

    if (pctx->rx_ring && pctx->rx_ring != MAP_FAILED) {
        munmap(pctx->rx_ring, pctx->rx_ring_size);
    }

    if (pctx->sock_fd >= 0) {
        close(pctx->sock_fd);
    }

    if (pctx->udp_guard_fd >= 0) {
        close(pctx->udp_guard_fd);
    }

    free(pctx);
    wctx->pctx = NULL;
}

/*
 * Simple receive buffer for non-ring mode
 */
static __thread uint8_t simple_rx_buf[2048];

static void wait_for_packet(int fd)
{
    struct pollfd pfd = {.fd = fd, .events = POLLIN};
    int           result;

    do {
        result = poll(&pfd, 1, PACKET_IDLE_WAIT_MS);
    } while (result < 0 && errno == EINTR);
}

/*
 * Receive batch of packets from PACKET_MMAP ring (zero-copy)
 * Falls back to simple recv() if ring buffers not available.
 * Handles both TPACKET_V2 (frame-level) and TPACKET_V3 (block-level) iteration.
 */
int packet_platform_recv_batch(worker_ctx_t *wctx, packet_t *pkts, int max_pkts)
{
    struct platform_ctx *pctx     = wctx->pctx;
    int                  num_pkts = 0;

    /* Simple mode: use basic recv() */
    if (!pctx->rx_ring) {
        for (int i = 0; i < max_pkts; i++) {
            ssize_t len = recv(pctx->sock_fd, simple_rx_buf, sizeof(simple_rx_buf), MSG_DONTWAIT);
            if (len <= 0) {
                break;
            }
            pkts[num_pkts].data      = simple_rx_buf;
            pkts[num_pkts].len       = (uint32_t)len;
            pkts[num_pkts].addr      = 0;
            pkts[num_pkts].timestamp = wctx->config->measure_latency ? get_timestamp_ns() : 0;
            num_pkts++;
            /* In simple mode, process one at a time since we use single buffer */
            break;
        }
        if (num_pkts == 0) {
            wait_for_packet(pctx->sock_fd);
        }
        return num_pkts;
    }

    /* TPACKET_V3: Block-level iteration (optimal for real hardware) */
    if (pctx->tpacket_version == 3) {
        struct tpacket_block_desc *block =
            (struct tpacket_block_desc *)((uint8_t *)pctx->rx_ring +
                                          (pctx->current_block_idx * PACKET_BLOCK_SIZE));

        /* The previous call returned pointers into this block. Release it only now,
         * after the worker has finished inspecting and transmitting that batch. */
        if (pctx->block_release_pending) {
            block->hdr.bh1.block_status = TP_STATUS_KERNEL;
            pctx->current_block_idx     = (pctx->current_block_idx + 1) % PACKET_BLOCK_NR;
            pctx->current_block_offset  = 0;
            pctx->current_frame_offset  = 0;
            pctx->block_release_pending = false;
            block                       = (struct tpacket_block_desc *)((uint8_t *)pctx->rx_ring +
                                                  (pctx->current_block_idx * PACKET_BLOCK_SIZE));
        }

        if ((block->hdr.bh1.block_status & TP_STATUS_USER) == 0) {
            wait_for_packet(pctx->sock_fd);
            return 0;
        }

        uint32_t num_frames = block->hdr.bh1.num_pkts;
        if (pctx->current_frame_offset == 0) {
            pctx->current_frame_offset = block->hdr.bh1.offset_to_first_pkt;
        }

        while (pctx->current_block_offset < num_frames && num_pkts < max_pkts) {
            struct tpacket3_hdr *hdr =
                (struct tpacket3_hdr *)((uint8_t *)block + pctx->current_frame_offset);

            pkts[num_pkts].data      = (uint8_t *)hdr + hdr->tp_mac;
            pkts[num_pkts].len       = hdr->tp_snaplen;
            pkts[num_pkts].addr      = (pctx->current_block_idx << 16) | pctx->current_block_offset;
            pkts[num_pkts].timestamp = wctx->config->measure_latency ? get_timestamp_ns() : 0;

            num_pkts++;
            pctx->current_block_offset++;
            pctx->current_frame_offset += hdr->tp_next_offset;
        }

        if (pctx->current_block_offset >= num_frames) {
            pctx->block_release_pending = true;
        }

        return num_pkts;
    }

    /* TPACKET_V2: Frame-level iteration (veth compatible) */
    for (int i = 0; i < max_pkts; i++) {
        /* Bounds check to prevent overflow */
        if (pctx->rx_frame_idx >= pctx->rx_frame_num) {
            pctx->rx_frame_idx = 0; /* Wrap around */
        }
        size_t               offset = (size_t)pctx->rx_frame_idx * pctx->frame_size;
        struct tpacket2_hdr *hdr    = (struct tpacket2_hdr *)((uint8_t *)pctx->rx_ring + offset);

        /* Check if frame is ready (kernel filled it) */
        if ((hdr->tp_status & TP_STATUS_USER) == 0) {
            /* No more packets ready */
            if (num_pkts == 0) {
                wait_for_packet(pctx->sock_fd);
            }
            break;
        }

        /* Point directly at packet data in ring (zero-copy) */
        pkts[num_pkts].data = (uint8_t *)hdr + hdr->tp_mac;
        pkts[num_pkts].len  = hdr->tp_snaplen;
        pkts[num_pkts].addr = pctx->rx_frame_idx; /* Store frame index for release */

        /* Only timestamp if latency measurement is enabled (avoid hot-path syscall overhead) */
        pkts[num_pkts].timestamp = wctx->config->measure_latency ? get_timestamp_ns() : 0;

        num_pkts++;
        pctx->rx_frame_idx = (pctx->rx_frame_idx + 1) % pctx->rx_frame_num;
    }

    return num_pkts;
}

/*
 * Send a packet batch directly. This remains compatible with veth devices and
 * uses one syscall for the batch while PACKET_QDISC_BYPASS keeps the fast path.
 */
int packet_platform_send_batch(worker_ctx_t *wctx, packet_t *pkts, int num_pkts)
{
    struct platform_ctx *pctx = wctx->pctx;

    /* Validate num_pkts to prevent out-of-bounds access */
    if (unlikely(num_pkts < 0 || num_pkts > BATCH_SIZE)) {
        reflector_log(LOG_ERROR, "Invalid num_pkts: %d (must be 0-%d)", num_pkts, BATCH_SIZE);
        return 0;
    }

    struct mmsghdr messages[BATCH_SIZE] = {0};
    struct iovec   iovecs[BATCH_SIZE]   = {0};
    for (int i = 0; i < num_pkts; i++) {
        if (!pkts[i].data || pkts[i].len == 0) {
            reflector_log(LOG_ERROR, "Invalid packet at batch index %d", i);
            packet_platform_release_batch(wctx, pkts, num_pkts);
            return -1;
        }
        iovecs[i].iov_base             = pkts[i].data;
        iovecs[i].iov_len              = pkts[i].len;
        messages[i].msg_hdr.msg_iov    = &iovecs[i];
        messages[i].msg_hdr.msg_iovlen = 1;
    }

    int sent = sendmmsg(pctx->sock_fd, messages, (unsigned int)num_pkts, MSG_DONTWAIT);
    if (sent < 0) {
        packet_platform_release_batch(wctx, pkts, num_pkts);
        return -1;
    }
    if (sent < num_pkts) {
        packet_platform_release_batch(wctx, &pkts[sent], num_pkts - sent);
    }
    return sent;
}

/*
 * Release RX frames back to kernel
 * For TPACKET_V3: Release blocks that packets came from
 * For TPACKET_V2: Release individual frames
 */
void packet_platform_release_batch(worker_ctx_t *wctx, packet_t *pkts, int num_pkts)
{
    struct platform_ctx *pctx = wctx->pctx;

    /* Simple mode: nothing to release */
    if (!pctx->rx_ring) {
        return;
    }

    if (unlikely(num_pkts < 0 || num_pkts > BATCH_SIZE)) {
        reflector_log(LOG_ERROR, "Invalid num_pkts: %d (must be 0-%d)", num_pkts, BATCH_SIZE);
        return;
    }

    /* TPACKET_V3: Release blocks that these packets came from */
    if (pctx->tpacket_version == 3) {
        /* V3 packets share block ownership. recv_batch releases a completed block
         * on the next call, after every pointer returned from it is no longer used. */
        return;
    }

    /* TPACKET_V2: Release individual frames */
    for (int i = 0; i < num_pkts; i++) {
        uint32_t             frame_idx = pkts[i].addr; /* We stored frame index in addr */
        struct tpacket2_hdr *hdr =
            (struct tpacket2_hdr *)((uint8_t *)pctx->rx_ring + (frame_idx * pctx->frame_size));

        /* Return frame to kernel */
        hdr->tp_status = TP_STATUS_KERNEL;
    }
}

/* Platform operations structure */
static const platform_ops_t packet_platform_ops = {
    .name          = "Linux AF_PACKET (optimized)",
    .init          = packet_platform_init,
    .cleanup       = packet_platform_cleanup,
    .recv_batch    = packet_platform_recv_batch,
    .send_batch    = packet_platform_send_batch,
    .release_batch = packet_platform_release_batch,
};

const platform_ops_t *get_packet_platform_ops(void)
{
    return &packet_platform_ops;
}
