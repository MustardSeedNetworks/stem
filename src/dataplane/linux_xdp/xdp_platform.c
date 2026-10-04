/*
 * xdp_platform.c - AF_XDP platform implementation for RFC2544 Test Master
 *
 * High-performance packet I/O using AF_XDP (eXpress Data Path).
 * Requires Linux kernel 5.4+ for best performance.
 *
 * Performance target: 10-40 Gbps
 */

#include "platform_config.h"
#include "rfc2544.h"
#include "rfc2544_internal.h"

#if HAVE_AF_XDP

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <arpa/inet.h>
#include <linux/if_ether.h>
#include <linux/if_link.h>
#include <linux/if_xdp.h>
#include <linux/sockios.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/socket.h>

#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <dirent.h>
#include <net/if.h>
#include <poll.h>
#include <pthread.h>
#include <unistd.h>
#include <xdp/libxdp.h>
#include <xdp/xsk.h>

#include "stem_errno.h"

/* worker_ctx_t and rfc2544_ctx_t are defined in rfc2544_internal.h */

typedef struct {
    uint8_t *data;
    uint32_t len;
    uint64_t timestamp;
    uint32_t seq_num;
    void    *platform_data;
} packet_t;

/* XDP configuration. NUM_FRAMES is the UMEM share of one receive queue. */
#define NUM_FRAMES 4096
#define FRAME_SIZE XSK_UMEM__DEFAULT_FRAME_SIZE
#define BATCH_SIZE 64

/* Free UMEM frames, as a stack: each frame is in it at most once, so it never
 * holds more than the UMEM has. */
typedef struct {
    uint64_t *frames;
    uint32_t  count;
} frame_allocator_t;

/* One socket per receive queue. All share the UMEM; each has its own fill and
 * completion rings, as the kernel requires for a queue of its own. */
typedef struct {
    struct xsk_socket   *xsk;
    struct xsk_ring_prod fill_ring;
    struct xsk_ring_cons comp_ring;
    struct xsk_ring_cons rx_ring;
    int                  fd;
} queue_socket_t;

/* Platform context */
typedef struct {
    struct xsk_umem  *umem;
    void             *umem_area;
    size_t            umem_size;
    frame_allocator_t frame_alloc;

    /* Queue 0's socket also transmits; the others only receive. */
    queue_socket_t      *queues;
    int                  num_queues;
    int                  next_rx_queue;
    struct xsk_ring_prod tx_ring;

    /* Interface info */
    int     if_index;
    uint8_t if_mac[6];

    /* Statistics */
    uint64_t tx_wakeups;
    uint64_t rx_wakeups;
} platform_ctx_t;

/* ============================================================================
 * Frame Allocator
 * ============================================================================ */

static int frame_alloc_init(frame_allocator_t *alloc, uint32_t num_frames)
{
    alloc->frames = calloc(num_frames, sizeof(uint64_t));
    if (!alloc->frames) {
        return -ENOMEM;
    }

    /* Hand out low addresses first */
    for (uint32_t i = 0; i < num_frames; i++) {
        alloc->frames[i] = (uint64_t)(num_frames - 1 - i) * FRAME_SIZE;
    }
    alloc->count = num_frames;

    return 0;
}

static void frame_alloc_cleanup(frame_allocator_t *alloc)
{
    free(alloc->frames);
    alloc->frames = NULL;
    alloc->count  = 0;
}

/* Callers size their requests to alloc->count. */
static uint64_t frame_alloc_get(frame_allocator_t *alloc)
{
    return alloc->frames[--alloc->count];
}

static void frame_alloc_put(frame_allocator_t *alloc, uint64_t frame)
{
    alloc->frames[alloc->count++] = frame;
}

/* ============================================================================
 * XDP Platform Operations
 * ============================================================================ */

static int is_rx_queue(const struct dirent *entry)
{
    return strncmp(entry->d_name, "rx-", 3) == 0;
}

static int rx_queue_count(const char *interface)
{
    char path[128];
    snprintf(path, sizeof(path), "/sys/class/net/%s/queues", interface);
    struct dirent **queues = NULL;
    int             count  = scandir(path, &queues, is_rx_queue, NULL);
    if (count < 0) {
        return -errno;
    }
    for (int i = 0; i < count; i++) {
        free(queues[i]);
    }
    free(queues);
    return count;
}

static void xdp_cleanup(worker_ctx_t *wctx)
{
    if (!wctx || !wctx->pctx) {
        return;
    }

    platform_ctx_t *pctx = wctx->pctx;

    /* libxdp detaches its program with the interface's last socket, and the
     * UMEM can go only once no socket uses it. */
    for (int q = pctx->num_queues - 1; q >= 0; q--) {
        if (pctx->queues[q].xsk) {
            xsk_socket__delete(pctx->queues[q].xsk);
        }
    }
    free(pctx->queues);
    if (pctx->umem) {
        xsk_umem__delete(pctx->umem);
    }
    if (pctx->umem_area) {
        munmap(pctx->umem_area, pctx->umem_size);
    }

    frame_alloc_cleanup(&pctx->frame_alloc);
    free(pctx);
    wctx->pctx = NULL;
}

static int fill_queue(platform_ctx_t *pctx, queue_socket_t *qs, uint32_t frames)
{
    uint32_t idx;
    if (xsk_ring_prod__reserve(&qs->fill_ring, frames, &idx) != frames) {
        return -ENOSPC;
    }
    for (uint32_t i = 0; i < frames; i++) {
        *xsk_ring_prod__fill_addr(&qs->fill_ring, idx++) = frame_alloc_get(&pctx->frame_alloc);
    }
    xsk_ring_prod__submit(&qs->fill_ring, frames);
    return 0;
}

static int xdp_init(rfc2544_ctx_t *ctx, worker_ctx_t *wctx)
{
    /* RSS spreads the reflected stream over every receive queue, so a socket
     * on queue 0 alone would count the rest as loss (stem#1533). */
    int num_queues = rx_queue_count(ctx->config.interface);
    if (num_queues < 0) {
        fprintf(stderr, "[xdp] Failed to count receive queues on %s: %s\n", ctx->config.interface,
                stem_strerror(-num_queues));
        return num_queues;
    }
    if (num_queues == 0) {
        fprintf(stderr, "[xdp] %s reports no receive queues\n", ctx->config.interface);
        return -ENODEV;
    }

    platform_ctx_t *pctx = calloc(1, sizeof(platform_ctx_t));
    if (!pctx) {
        return -ENOMEM;
    }
    wctx->pctx = pctx;

    int ret;

    pctx->if_index = if_nametoindex(ctx->config.interface);
    if (pctx->if_index == 0) {
        fprintf(stderr, "[xdp] Failed to get interface index for %s\n", ctx->config.interface);
        xdp_cleanup(wctx);
        return -ENODEV;
    }

    pctx->queues = calloc((size_t)num_queues, sizeof(queue_socket_t));
    if (!pctx->queues) {
        xdp_cleanup(wctx);
        return -ENOMEM;
    }
    pctx->num_queues = num_queues;

    /* Increase RLIMIT_MEMLOCK for UMEM */
    struct rlimit rlim = {RLIM_INFINITY, RLIM_INFINITY};
    if (setrlimit(RLIMIT_MEMLOCK, &rlim)) {
        fprintf(stderr, "[xdp] Warning: Failed to increase RLIMIT_MEMLOCK\n");
    }

    uint32_t total_frames = NUM_FRAMES * (uint32_t)num_queues;
    pctx->umem_size       = (size_t)total_frames * FRAME_SIZE;
    void *area            = mmap(NULL, pctx->umem_size, PROT_READ | PROT_WRITE,
                                 MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB, -1, 0);
    if (area == MAP_FAILED) {
        /* Fall back to regular pages */
        area =
            mmap(NULL, pctx->umem_size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
        if (area == MAP_FAILED) {
            fprintf(stderr, "[xdp] Failed to allocate UMEM\n");
            xdp_cleanup(wctx);
            return -ENOMEM;
        }
    }
    pctx->umem_area = area;

    ret = frame_alloc_init(&pctx->frame_alloc, total_frames);
    if (ret < 0) {
        xdp_cleanup(wctx);
        return ret;
    }

    struct xsk_umem_config umem_cfg = {
        .fill_size      = NUM_FRAMES / 2,
        .comp_size      = NUM_FRAMES / 2,
        .frame_size     = FRAME_SIZE,
        .frame_headroom = XSK_UMEM__DEFAULT_FRAME_HEADROOM,
        .flags          = 0,
    };

    ret = xsk_umem__create(&pctx->umem, pctx->umem_area, pctx->umem_size,
                           &pctx->queues[0].fill_ring, &pctx->queues[0].comp_ring, &umem_cfg);
    if (ret) {
        fprintf(stderr, "[xdp] Failed to create UMEM: %s\n", stem_strerror(-ret));
        pctx->umem = NULL;
        xdp_cleanup(wctx);
        return ret;
    }

    /* libxdp attaches its default program with the first socket and adds each
     * socket to the program's map under its queue, so every queue's frames
     * reach the socket bound to it. Without a program nothing reaches an RX
     * ring (stem#1328). No mode flag: libxdp binds the socket before it
     * attaches the program, so retrying in another mode after a failed attach
     * rebinds a queue the kernel has not released yet and fails EBUSY. Unset,
     * the attach itself picks native or generic. */
    struct xsk_socket_config xsk_cfg = {
        .rx_size    = NUM_FRAMES / 2,
        .tx_size    = NUM_FRAMES / 2,
        .bind_flags = XDP_USE_NEED_WAKEUP,
    };

    for (int q = 0; q < num_queues; q++) {
        queue_socket_t *qs = &pctx->queues[q];
        if (q == 0) {
            ret = xsk_socket__create(&qs->xsk, ctx->config.interface, 0, pctx->umem, &qs->rx_ring,
                                     &pctx->tx_ring, &xsk_cfg);
        } else {
            ret = xsk_socket__create_shared(&qs->xsk, ctx->config.interface, (uint32_t)q,
                                            pctx->umem, &qs->rx_ring, NULL, &qs->fill_ring,
                                            &qs->comp_ring, &xsk_cfg);
        }
        if (ret) {
            fprintf(stderr, "[xdp] Failed to create XDP socket on %s queue %d: %s\n",
                    ctx->config.interface, q, stem_strerror(-ret));
            qs->xsk = NULL;
            xdp_cleanup(wctx);
            return ret;
        }
        qs->fd = xsk_socket__fd(qs->xsk);

        ret = fill_queue(pctx, qs, NUM_FRAMES / 2);
        if (ret < 0) {
            fprintf(stderr, "[xdp] Failed to populate the fill ring of queue %d\n", q);
            xdp_cleanup(wctx);
            return ret;
        }
    }

    /* Get interface MAC */
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock >= 0) {
        struct ifreq ifr;
        memset(&ifr, 0, sizeof(ifr));
        memcpy(ifr.ifr_name, ctx->config.interface,
               strnlen(ctx->config.interface, sizeof(ifr.ifr_name) - 1));
        if (ioctl(sock, SIOCGIFHWADDR, &ifr) == 0) {
            memcpy(pctx->if_mac, ifr.ifr_hwaddr.sa_data, 6);
            memcpy(ctx->local_mac, pctx->if_mac, 6);
        }
        close(sock);
    }

    fprintf(stderr, "[xdp] Initialized on %s, %d receive queue(s)\n", ctx->config.interface,
            num_queues);

    return 0;
}

static int xdp_send_batch(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    if (!wctx || !wctx->pctx || !pkts || count <= 0) {
        return -EINVAL;
    }

    platform_ctx_t *pctx = wctx->pctx;
    queue_socket_t *txq  = &pctx->queues[0];
    int             sent = 0;

    /* Complete any pending TX */
    uint32_t     idx_comp;
    unsigned int completed = xsk_ring_cons__peek(&txq->comp_ring, BATCH_SIZE, &idx_comp);
    if (completed > 0) {
        for (unsigned int i = 0; i < completed; i++) {
            uint64_t addr = *xsk_ring_cons__comp_addr(&txq->comp_ring, idx_comp++);
            frame_alloc_put(&pctx->frame_alloc, addr);
        }
        xsk_ring_cons__release(&txq->comp_ring, completed);
    }

    /* Reserve no more TX slots than there are free frames: a reserved slot
     * cannot be handed back, and the next submit would publish it unwritten. */
    uint32_t want =
        (uint32_t)count < pctx->frame_alloc.count ? (uint32_t)count : pctx->frame_alloc.count;
    uint32_t     idx_tx;
    unsigned int reserved = xsk_ring_prod__reserve(&pctx->tx_ring, want, &idx_tx);
    if (reserved == 0) {
        /* TX ring full, need wakeup */
        if (xsk_ring_prod__needs_wakeup(&pctx->tx_ring)) {
            sendto(txq->fd, NULL, 0, MSG_DONTWAIT, NULL, 0);
            pctx->tx_wakeups++;
        }
        return 0;
    }

    /* Copy packets to UMEM and fill TX descriptors */
    for (unsigned int i = 0; i < reserved; i++) {
        uint64_t addr = frame_alloc_get(&pctx->frame_alloc);

        /* Copy packet data to UMEM frame */
        uint8_t *frame = (uint8_t *)pctx->umem_area + addr;
        memcpy(frame, pkts[sent].data, pkts[sent].len);

        /* Fill TX descriptor */
        struct xdp_desc *desc = xsk_ring_prod__tx_desc(&pctx->tx_ring, idx_tx++);
        desc->addr            = addr;
        desc->len             = pkts[sent].len;

        sent++;
        wctx->tx_packets++;
        wctx->tx_bytes += pkts[sent - 1].len;
    }

    xsk_ring_prod__submit(&pctx->tx_ring, sent);

    /* Kick TX if needed */
    if (xsk_ring_prod__needs_wakeup(&pctx->tx_ring)) {
        sendto(txq->fd, NULL, 0, MSG_DONTWAIT, NULL, 0);
        pctx->tx_wakeups++;
    }

    return sent;
}

/* Moves up to max_count frames from one queue's RX ring into pkts and refills
 * that queue's fill ring. */
static int recv_queue(worker_ctx_t *wctx, queue_socket_t *qs, packet_t *pkts, int max_count)
{
    platform_ctx_t *pctx     = wctx->pctx;
    int             received = 0;

    if (xsk_ring_prod__needs_wakeup(&qs->fill_ring)) {
        struct pollfd fds = {.fd = qs->fd, .events = POLLIN};
        poll(&fds, 1, 0); /* Non-blocking poll */
        pctx->rx_wakeups++;
    }

    uint32_t     idx_rx;
    unsigned int rcvd = xsk_ring_cons__peek(&qs->rx_ring, (uint32_t)max_count, &idx_rx);
    if (rcvd == 0) {
        return 0;
    }

    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    uint64_t now_ns = ts.tv_sec * 1000000000ULL + ts.tv_nsec;

    for (unsigned int i = 0; i < rcvd; i++) {
        const struct xdp_desc *desc  = xsk_ring_cons__rx_desc(&qs->rx_ring, idx_rx++);
        uint8_t               *frame = (uint8_t *)pctx->umem_area + desc->addr;

        pkts[received].data = malloc(desc->len);
        if (!pkts[received].data) {
            frame_alloc_put(&pctx->frame_alloc, desc->addr);
            continue;
        }

        memcpy(pkts[received].data, frame, desc->len);
        pkts[received].len           = desc->len;
        pkts[received].timestamp     = now_ns;
        pkts[received].platform_data = (void *)(uintptr_t)desc->addr;

        received++;
        wctx->rx_packets++;
        wctx->rx_bytes += desc->len;
    }

    xsk_ring_cons__release(&qs->rx_ring, rcvd);

    /* As for TX: reserve only slots a free frame will go into */
    uint32_t     want = rcvd < pctx->frame_alloc.count ? rcvd : pctx->frame_alloc.count;
    uint32_t     idx_fill;
    unsigned int slots = xsk_ring_prod__reserve(&qs->fill_ring, want, &idx_fill);
    for (unsigned int i = 0; i < slots; i++) {
        *xsk_ring_prod__fill_addr(&qs->fill_ring, idx_fill++) = frame_alloc_get(&pctx->frame_alloc);
    }
    xsk_ring_prod__submit(&qs->fill_ring, slots);

    return received;
}

static int xdp_recv_batch(worker_ctx_t *wctx, packet_t *pkts, int max_count)
{
    if (!wctx || !wctx->pctx || !pkts || max_count <= 0) {
        return -EINVAL;
    }

    platform_ctx_t *pctx     = wctx->pctx;
    int             received = 0;

    /* Start each batch at the next queue so a busy one cannot starve the rest */
    int first           = pctx->next_rx_queue;
    pctx->next_rx_queue = (first + 1) % pctx->num_queues;
    for (int i = 0; i < pctx->num_queues && received < max_count; i++) {
        queue_socket_t *qs = &pctx->queues[(first + i) % pctx->num_queues];
        received += recv_queue(wctx, qs, pkts + received, max_count - received);
    }

    return received;
}

static void xdp_release_batch(worker_ctx_t *wctx, packet_t *pkts, int count)
{
    if (!wctx || !wctx->pctx) {
        return;
    }

    platform_ctx_t *pctx = wctx->pctx;

    for (int i = 0; i < count; i++) {
        if (pkts[i].platform_data) {
            frame_alloc_put(&pctx->frame_alloc, (uint64_t)(uintptr_t)pkts[i].platform_data);
        }
        free(pkts[i].data);
        pkts[i].data          = NULL;
        pkts[i].platform_data = NULL;
    }
}

static uint64_t xdp_get_timestamp(worker_ctx_t *wctx, packet_t *pkt)
{
    (void)wctx;
    return pkt->timestamp;
}

/* Platform ops structure */
static const struct {
    const char *name;
    int (*init)(rfc2544_ctx_t *ctx, worker_ctx_t *wctx);
    void (*cleanup)(worker_ctx_t *wctx);
    int (*send_batch)(worker_ctx_t *wctx, packet_t *pkts, int count);
    int (*recv_batch)(worker_ctx_t *wctx, packet_t *pkts, int max_count);
    void (*release_batch)(worker_ctx_t *wctx, packet_t *pkts, int count);
    uint64_t (*get_tx_timestamp)(worker_ctx_t *wctx, packet_t *pkt);
    uint64_t (*get_rx_timestamp)(worker_ctx_t *wctx, packet_t *pkt);
} xdp_ops = {
    .name             = "AF_XDP",
    .init             = xdp_init,
    .cleanup          = xdp_cleanup,
    .send_batch       = xdp_send_batch,
    .recv_batch       = xdp_recv_batch,
    .release_batch    = xdp_release_batch,
    .get_tx_timestamp = xdp_get_timestamp,
    .get_rx_timestamp = xdp_get_timestamp,
};

const void *get_dataplane_xdp_platform_ops(void)
{
    return &xdp_ops;
}

#endif /* HAVE_AF_XDP */
