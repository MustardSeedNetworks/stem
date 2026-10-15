/*
 * xdp_platform.c - Linux AF_XDP platform implementation
 *
 * Copyright (c) 2025 Kris Armstrong
 *
 * This implements the platform abstraction layer using AF_XDP sockets
 * for zero-copy packet I/O on Linux. Achieves line-rate performance
 * through:
 * - Zero-copy UMEM shared with kernel
 * - Batched packet processing
 * - Busy-polling mode
 * - Per-queue AF_XDP sockets
 */

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <linux/if_link.h>
#include <linux/if_xdp.h>
#include <sys/mman.h>
#include <sys/socket.h>

#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <poll.h>
#include <unistd.h>
#include <xdp/xsk.h>

#include "reflector.h"
#include "stem_errno.h"
#include "stem_xdp_caps.h"
#include "stem_xsk_prog.h"

/* Platform-specific context for AF_XDP */
struct platform_ctx {
    struct xsk_socket_info {
        struct xsk_ring_cons rx;
        struct xsk_ring_prod tx;
        struct xsk_umem_info {
            struct xsk_ring_prod fq; /* Fill queue */
            struct xsk_ring_cons cq; /* Completion queue */
            struct xsk_umem     *umem;
            void                *buffer;
            uint64_t             buffer_size;
        } umem;
        struct xsk_socket *xsk;
        uint32_t           outstanding_tx;
    } xsk_info;

    /* Queue 0's worker attaches the program and owns it; the others add
     * their sockets to its map. */
    stem_xsk_prog_t xsk_prog;

    uint32_t frame_size;
    uint32_t num_frames;

    /* Kernel drop total already reported; XDP_STATISTICS is cumulative. */
    uint64_t drops_reported;
};

/*
 * Configure UMEM (User Memory) for zero-copy packet buffers
 */
static int configure_umem(struct platform_ctx *pctx, void *buffer, uint64_t size)
{
    struct xsk_umem_config cfg = {.fill_size      = NUM_FRAMES / 2,
                                  .comp_size      = NUM_FRAMES / 2,
                                  .frame_size     = pctx->frame_size,
                                  .frame_headroom = 0, /* XDP_PACKET_HEADROOM */
                                  .flags          = 0};

    int ret = xsk_umem__create(&pctx->xsk_info.umem.umem, buffer, size, &pctx->xsk_info.umem.fq,
                               &pctx->xsk_info.umem.cq, &cfg);

    if (ret) {
        reflector_log(LOG_ERROR, "Failed to create UMEM: %s", stem_strerror(-ret));
        return ret;
    }

    pctx->xsk_info.umem.buffer      = buffer;
    pctx->xsk_info.umem.buffer_size = size;

    return 0;
}

/*
 * Populate fill queue with buffers for kernel to use
 */
static void populate_fill_queue(struct platform_ctx *pctx, uint32_t num)
{
    uint32_t idx;

    if (xsk_ring_prod__reserve(&pctx->xsk_info.umem.fq, num, &idx) != num) {
        reflector_log(LOG_ERROR, "Failed to reserve fill queue entries");
        return;
    }

    for (uint32_t i = 0; i < num; i++) {
        uint64_t addr                                             = (uint64_t)i * pctx->frame_size;
        *xsk_ring_prod__fill_addr(&pctx->xsk_info.umem.fq, idx++) = addr;
    }

    xsk_ring_prod__submit(&pctx->xsk_info.umem.fq, num);
}

/*
 * Name the mode the kernel attached the interface's XDP program in.
 */
static const char *xdp_attach_mode(int ifindex)
{
    LIBBPF_OPTS(bpf_xdp_query_opts, opts);
    if (bpf_xdp_query(ifindex, 0, &opts) != 0) {
        return "unknown";
    }
    switch (opts.attach_mode) {
    case XDP_ATTACHED_DRV:
        return "driver";
    case XDP_ATTACHED_SKB:
        return "SKB";
    case XDP_ATTACHED_HW:
        return "offload";
    case XDP_ATTACHED_MULTI:
        return "multiple";
    default:
        return "unknown";
    }
}

/*
 * Initialize AF_XDP socket
 */
static int init_xsk(const reflector_ctx_t *rctx, worker_ctx_t *wctx)
{
    struct platform_ctx *pctx = wctx->pctx;
    reflector_config_t  *cfg  = wctx->config;
    int                  ret;

    const stem_xsk_prog_t *prog = &pctx->xsk_prog;
    if (wctx->queue_id == 0) {
        ret = stem_xsk_prog_attach(&pctx->xsk_prog, cfg->ifindex, (uint32_t)rctx->num_workers,
                                   cfg->ito_port);
        if (ret) {
            reflector_log(LOG_ERROR, "Failed to attach the XDP program to %s: %s", cfg->ifname,
                          stem_strerror(-ret));
            return ret;
        }
    } else {
        prog = &rctx->platform_contexts[0]->xsk_prog;
    }

    /* Neither XDP_ZEROCOPY nor XDP_COPY: the kernel binds zero-copy where the
     * driver supports it and copy mode where it does not. Demanding zero-copy
     * failed the bind on every copy-mode driver, and retrying in copy mode
     * would rebind a queue the kernel has not released yet (stem#1532). */
    struct xsk_socket_config xsk_cfg = {.rx_size      = NUM_FRAMES / 2,
                                        .tx_size      = NUM_FRAMES / 2,
                                        .libxdp_flags = XSK_LIBXDP_FLAGS__INHIBIT_PROG_LOAD,
                                        .bind_flags   = XDP_USE_NEED_WAKEUP};

    /* Create AF_XDP socket */
    ret = xsk_socket__create(&pctx->xsk_info.xsk, cfg->ifname, wctx->queue_id,
                             pctx->xsk_info.umem.umem, &pctx->xsk_info.rx, &pctx->xsk_info.tx,
                             &xsk_cfg);

    if (ret) {
        reflector_log(LOG_ERROR, "Failed to create XSK socket: %s", stem_strerror(-ret));
        stem_xsk_prog_detach(&pctx->xsk_prog);
        return ret;
    }

    ret = stem_xsk_prog_add(prog, (uint32_t)wctx->queue_id, xsk_socket__fd(pctx->xsk_info.xsk));
    if (ret) {
        reflector_log(LOG_ERROR, "Failed to add queue %d's socket to the XDP map: %s",
                      wctx->queue_id, stem_strerror(-ret));
        xsk_socket__delete(pctx->xsk_info.xsk);
        pctx->xsk_info.xsk = NULL;
        stem_xsk_prog_detach(&pctx->xsk_prog);
        return ret;
    }

    struct xdp_options opts     = {0};
    socklen_t          opts_len = sizeof(opts);
    if (getsockopt(xsk_socket__fd(pctx->xsk_info.xsk), SOL_XDP, XDP_OPTIONS, &opts, &opts_len) ==
        0) {
        reflector_log(LOG_INFO, "AF_XDP queue %d bound in %s mode, XDP program attached in %s mode",
                      wctx->queue_id, (opts.flags & XDP_OPTIONS_ZEROCOPY) ? "zero-copy" : "copy",
                      xdp_attach_mode(cfg->ifindex));
    }

    return 0;
}

/*
 * Initialize platform (AF_XDP)
 */
int xdp_platform_init(const reflector_ctx_t *rctx, worker_ctx_t *wctx)
{
    int caps = stem_xdp_caps_check();
    if (caps < 0) {
        reflector_log(LOG_ERROR, "%s: %s", STEM_XDP_CAPS_NEEDED, stem_strerror(-caps));
        return caps;
    }

    const reflector_config_t *cfg = wctx->config;
    /* The XDP program redirects only frames to the reflector's port; with any
     * port it would take the host's DNS and DHCP too (stem#1767). */
    if (cfg->ito_port == 0) {
        reflector_log(LOG_ERROR,
                      "AF_XDP requires a UDP port to leave the host's own traffic alone");
        return -EINVAL;
    }

    struct platform_ctx *pctx = calloc(1, sizeof(*pctx));
    if (!pctx) {
        reflector_log(LOG_ERROR, "Failed to allocate platform context");
        return -ENOMEM;
    }

    wctx->pctx       = pctx;
    pctx->xsk_prog   = STEM_XSK_PROG_NONE;
    pctx->frame_size = wctx->config->frame_size;
    pctx->num_frames = wctx->config->num_frames;

    /* Allocate UMEM buffer */
    uint64_t umem_size = (uint64_t)pctx->num_frames * pctx->frame_size;
    void    *umem_buffer;

    /* Try huge pages if enabled in config (better TLB utilization) */
    if (cfg->use_huge_pages) {
#ifndef MAP_HUGETLB
#define MAP_HUGETLB 0x40000 /* Linux-specific flag for huge pages */
#endif
        umem_buffer = mmap(NULL, umem_size, PROT_READ | PROT_WRITE,
                           MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB, -1, 0);

        if (umem_buffer == MAP_FAILED) {
            reflector_log(LOG_WARN,
                          "Huge pages requested but not available, falling back to normal pages");
            umem_buffer =
                mmap(NULL, umem_size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
        } else {
            reflector_log(LOG_INFO, "Using huge pages for UMEM (reduces TLB misses)");
        }
    } else {
        /* Use normal pages */
        umem_buffer =
            mmap(NULL, umem_size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    }

    if (umem_buffer == MAP_FAILED) {
        int saved_errno = errno;
        reflector_log(LOG_ERROR, "Failed to allocate UMEM: %s", stem_strerror(saved_errno));
        free(pctx);
        wctx->pctx = NULL; /* Prevent use-after-free */
        return saved_errno ? -saved_errno : -ENOMEM;
    }

    reflector_log(LOG_INFO, "Allocated UMEM: %lu MB (%u frames of %u bytes)",
                  umem_size / (1024 * 1024), pctx->num_frames, pctx->frame_size);

    /* Configure UMEM */
    int ret = configure_umem(pctx, umem_buffer, umem_size);
    if (ret) {
        munmap(umem_buffer, umem_size);
        free(pctx);
        wctx->pctx = NULL; /* Prevent use-after-free */
        return ret;
    }

    /* Initialize AF_XDP socket */
    ret = init_xsk(rctx, wctx);
    if (ret) {
        xsk_umem__delete(pctx->xsk_info.umem.umem);
        munmap(umem_buffer, umem_size);
        free(pctx);
        wctx->pctx = NULL; /* Prevent use-after-free */
        return ret;
    }

    /* Populate fill queue with initial buffers */
    populate_fill_queue(pctx, pctx->num_frames / 2);

    pctx->xsk_info.outstanding_tx = 0;

    reflector_log(LOG_INFO, "AF_XDP platform initialized for worker %d", wctx->worker_id);
    return 0;
}

/*
 * Cleanup platform
 */
void xdp_platform_cleanup(worker_ctx_t *wctx)
{
    struct platform_ctx *pctx = wctx->pctx;
    if (!pctx) {
        return;
    }

    stem_xsk_prog_detach(&pctx->xsk_prog);

    /* Delete AF_XDP socket */
    if (pctx->xsk_info.xsk) {
        xsk_socket__delete(pctx->xsk_info.xsk);
    }

    /* Delete UMEM */
    if (pctx->xsk_info.umem.umem) {
        munmap(pctx->xsk_info.umem.buffer, pctx->xsk_info.umem.buffer_size);
        xsk_umem__delete(pctx->xsk_info.umem.umem);
    }

    free(pctx);
    wctx->pctx = NULL;
}

/*
 * Return transmitted buffers from the completion queue to the fill queue.
 * Every UMEM frame in circulation fits in the fill queue, so the reservation
 * fails only if ownership was lost elsewhere; the completions then stay queued
 * rather than leaking their frames.
 */
static void xdp_recycle_completed_tx(struct platform_ctx *pctx)
{
    uint32_t idx_cq;
    uint32_t idx_fq;

    uint32_t completed = xsk_ring_cons__peek(&pctx->xsk_info.umem.cq, BATCH_SIZE, &idx_cq);
    if (completed == 0) {
        return;
    }
    if (xsk_ring_prod__reserve(&pctx->xsk_info.umem.fq, completed, &idx_fq) != completed) {
        xsk_ring_cons__cancel(&pctx->xsk_info.umem.cq, completed);
        return;
    }
    for (uint32_t i = 0; i < completed; i++) {
        *xsk_ring_prod__fill_addr(&pctx->xsk_info.umem.fq, idx_fq++) =
            *xsk_ring_cons__comp_addr(&pctx->xsk_info.umem.cq, idx_cq++);
    }
    xsk_ring_prod__submit(&pctx->xsk_info.umem.fq, completed);
    xsk_ring_cons__release(&pctx->xsk_info.umem.cq, completed);
    pctx->xsk_info.outstanding_tx -= completed;
}

/* Return received frames that will not be transmitted to the fill queue. */
static void xdp_return_to_fill(struct platform_ctx *pctx, const packet_t *pkts, int num_pkts)
{
    uint32_t idx_fq;

    if (xsk_ring_prod__reserve(&pctx->xsk_info.umem.fq, (uint32_t)num_pkts, &idx_fq) !=
        (uint32_t)num_pkts) {
        reflector_log(LOG_ERROR, "Fill queue full: %d UMEM frames lost", num_pkts);
        return;
    }
    for (int i = 0; i < num_pkts; i++) {
        *xsk_ring_prod__fill_addr(&pctx->xsk_info.umem.fq, idx_fq++) = pkts[i].addr;
    }
    xsk_ring_prod__submit(&pctx->xsk_info.umem.fq, (uint32_t)num_pkts);
}

/* Ask the kernel to transmit what the TX ring holds. */
static void xdp_kick_tx(struct platform_ctx *pctx)
{
    if (xsk_ring_prod__needs_wakeup(&pctx->xsk_info.tx)) {
        sendto(xsk_socket__fd(pctx->xsk_info.xsk), NULL, 0, MSG_DONTWAIT, NULL, 0);
    }
}

/*
 * Copy mode transmits a bounded number of descriptors per kick, so frames
 * queued behind them wait for the next one. Without this, the last frames of a
 * burst stay in the TX ring until another frame arrives, which at the end of a
 * trial is never.
 */
static void xdp_flush_tx(struct platform_ctx *pctx)
{
    xdp_recycle_completed_tx(pctx);
    if (pctx->xsk_info.outstanding_tx > 0) {
        xdp_kick_tx(pctx);
    }
}

/*
 * Receive batch of packets (zero-copy)
 */
int xdp_platform_recv_batch(worker_ctx_t *wctx, packet_t *pkts, int max_pkts)
{
    struct platform_ctx *pctx = wctx->pctx;
    uint32_t             idx_rx;
    int                  rcvd;

    /* Check if we need to wake up kernel (NEED_WAKEUP flag) */
    if (xsk_ring_prod__needs_wakeup(&pctx->xsk_info.umem.fq)) {
        recvfrom(xsk_socket__fd(pctx->xsk_info.xsk), NULL, 0, MSG_DONTWAIT, NULL, NULL);
    }

    /* Receive packets from RX ring */
    rcvd = xsk_ring_cons__peek(&pctx->xsk_info.rx, max_pkts, &idx_rx);
    if (rcvd == 0) {
        xdp_flush_tx(pctx);
        return 0;
    }

    /* Process received packets */
    for (int i = 0; i < rcvd; i++) {
        uint64_t addr = xsk_ring_cons__rx_desc(&pctx->xsk_info.rx, idx_rx)->addr;
        uint32_t len  = xsk_ring_cons__rx_desc(&pctx->xsk_info.rx, idx_rx++)->len;

        pkts[i].addr = addr;
        pkts[i].len  = len;
        pkts[i].data = xsk_umem__get_data(pctx->xsk_info.umem.buffer, addr);

        /* Only timestamp if latency measurement is enabled (avoid hot-path syscall overhead) */
        pkts[i].timestamp = wctx->config->measure_latency ? get_timestamp_ns() : 0;
    }

    /* Release RX descriptors */
    xsk_ring_cons__release(&pctx->xsk_info.rx, rcvd);

    return rcvd;
}

/*
 * Send batch of packets (zero-copy). Takes ownership of every buffer: a sent
 * frame belongs to the kernel until its completion is recycled, and an unsent
 * one goes straight back to the fill queue.
 */
int xdp_platform_send_batch(worker_ctx_t *wctx, packet_t *pkts, int num_pkts)
{
    struct platform_ctx *pctx = wctx->pctx;
    uint32_t             idx_tx;

    /* Validate num_pkts to prevent out-of-bounds access */
    if (unlikely(num_pkts < 0 || num_pkts > BATCH_SIZE)) {
        reflector_log(LOG_ERROR, "Invalid num_pkts: %d (must be 0-%d)", num_pkts, BATCH_SIZE);
        return 0;
    }

    /* Eagerly recycle completed TX buffers to prevent UMEM exhaustion */
    xdp_recycle_completed_tx(pctx);

    /* The reservation is all or nothing: a full TX ring drops the batch. */
    int reserved = xsk_ring_prod__reserve(&pctx->xsk_info.tx, num_pkts, &idx_tx);
    if (reserved == 0) {
        xdp_return_to_fill(pctx, pkts, num_pkts);
        return 0;
    }

    /* Submit packets to TX ring */
    /* Note: Stats are counted in core.c, not here (to avoid double-counting) */
    for (int i = 0; i < reserved; i++) {
        struct xdp_desc *tx_desc = xsk_ring_prod__tx_desc(&pctx->xsk_info.tx, idx_tx++);
        tx_desc->addr            = pkts[i].addr;
        tx_desc->len             = pkts[i].len;
    }

    xsk_ring_prod__submit(&pctx->xsk_info.tx, reserved);
    pctx->xsk_info.outstanding_tx += reserved;

    xdp_kick_tx(pctx);
    return reserved;
}

/*
 * Return received frames that were not transmitted to the fill queue.
 */
void xdp_platform_release_batch(worker_ctx_t *wctx, const packet_t *pkts, int num_pkts)
{
    struct platform_ctx *pctx = wctx->pctx;

    /* Validate num_pkts to prevent out-of-bounds access */
    if (unlikely(num_pkts < 0 || num_pkts > BATCH_SIZE)) {
        reflector_log(LOG_ERROR, "Invalid num_pkts: %d (must be 0-%d)", num_pkts, BATCH_SIZE);
        return;
    }

    xdp_return_to_fill(pctx, pkts, num_pkts);
}

/*
 * rx_dropped counts frames that found the fill queue empty, rx_ring_full ones
 * that found the RX ring full. A kernel older than 5.9 returns only the first
 * fields, so the rest stay zero.
 */
static uint64_t xdp_platform_take_kernel_drops(const worker_ctx_t *wctx)
{
    struct platform_ctx  *pctx = wctx->pctx;
    struct xdp_statistics st   = {0};
    socklen_t             len  = sizeof(st);
    if (getsockopt(xsk_socket__fd(pctx->xsk_info.xsk), SOL_XDP, XDP_STATISTICS, &st, &len) < 0) {
        return 0;
    }
    uint64_t total       = st.rx_dropped + st.rx_ring_full;
    uint64_t delta       = total - pctx->drops_reported;
    pctx->drops_reported = total;
    return delta;
}

/* Platform operations structure */
static const platform_ops_t xdp_platform_ops = {
    .name              = "Linux AF_XDP",
    .init              = xdp_platform_init,
    .cleanup           = xdp_platform_cleanup,
    .recv_batch        = xdp_platform_recv_batch,
    .send_batch        = xdp_platform_send_batch,
    .release_batch     = xdp_platform_release_batch,
    .take_kernel_drops = xdp_platform_take_kernel_drops,
};

const platform_ops_t *get_xdp_platform_ops(void)
{
    return &xdp_platform_ops;
}
