/*
 * core.c - Core reflector engine and worker thread management
 */

#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <unistd.h>
#ifndef __APPLE__
#include <pthread.h>
#include <sched.h>
#else
#include <dispatch/dispatch.h>
#endif
#include "platform_config.h"
#include "reflector.h"

/* Forward declarations */
#if HAVE_AF_XDP
extern const platform_ops_t *get_xdp_platform_ops(void);
#endif
#ifdef __linux__
extern const platform_ops_t *get_packet_platform_ops(void);
#endif
#ifdef __APPLE__
extern const platform_ops_t *get_bpf_platform_ops(void);
#endif

/* Global platform ops (set at runtime) */
static const platform_ops_t *platform_ops = NULL;

/* Batched statistics update structure (reduces cache line bouncing) */
typedef struct {
    uint64_t        packets_received;
    uint64_t        packets_reflected;
    uint64_t        bytes_received;
    uint64_t        bytes_reflected;
    uint64_t        sig_probeot_count;
    uint64_t        sig_dataot_count;
    uint64_t        sig_latency_count;
    uint64_t        sig_rfc2544_count;
    uint64_t        sig_y1564_count;
    uint64_t        sig_msn_count;
    uint64_t        sig_unknown_count;
    uint64_t        err_tx_failed;
    latency_stats_t latency_batch;
    int             batch_count;
} stats_batch_t;

/* Flush batched statistics to worker stats */
static inline void flush_stats_batch(reflector_stats_t *stats, stats_batch_t *batch)
{
    if (unlikely(batch->batch_count == 0)) {
        return;
    }

    /* Flush accumulated counters */
    stats->packets_received += batch->packets_received;
    stats->packets_reflected += batch->packets_reflected;
    stats->bytes_received += batch->bytes_received;
    stats->bytes_reflected += batch->bytes_reflected;

    /* Signature counters */
    stats->sig_probeot_count += batch->sig_probeot_count;
    stats->sig_dataot_count += batch->sig_dataot_count;
    stats->sig_latency_count += batch->sig_latency_count;
    stats->sig_rfc2544_count += batch->sig_rfc2544_count;
    stats->sig_y1564_count += batch->sig_y1564_count;
    stats->sig_msn_count += batch->sig_msn_count;
    stats->sig_unknown_count += batch->sig_unknown_count;

    /* Error counters */
    stats->err_tx_failed += batch->err_tx_failed;
    stats->tx_errors += batch->err_tx_failed;

    /* Merge latency statistics if any were collected */
    if (batch->latency_batch.count > 0) {
        stats->latency.count += batch->latency_batch.count;
        stats->latency.total_ns += batch->latency_batch.total_ns;

        /* Update min/max */
        if (stats->latency.count == batch->latency_batch.count) {
            /* First batch */
            stats->latency.min_ns = batch->latency_batch.min_ns;
            stats->latency.max_ns = batch->latency_batch.max_ns;
        } else {
            if (batch->latency_batch.min_ns < stats->latency.min_ns) {
                stats->latency.min_ns = batch->latency_batch.min_ns;
            }
            if (batch->latency_batch.max_ns > stats->latency.max_ns) {
                stats->latency.max_ns = batch->latency_batch.max_ns;
            }
        }

        /* Recalculate average */
        stats->latency.avg_ns = (double)stats->latency.total_ns / (double)stats->latency.count;
    }

    /* Reset batch */
    memset(batch, 0, sizeof(*batch));
}

static inline void flush_worker_stats(worker_ctx_t *wctx, stats_batch_t *batch)
{
    pthread_mutex_lock(&wctx->stats_mutex);
    flush_stats_batch(&wctx->stats, batch);
    pthread_mutex_unlock(&wctx->stats_mutex);
}

/* Worker main loop with batched statistics */
#ifdef __APPLE__
static void worker_loop(worker_ctx_t *wctx)
#else
static void *worker_thread(void *arg)
#endif
{
#ifndef __APPLE__
    worker_ctx_t *wctx = (worker_ctx_t *)arg;
#endif
    packet_t      pkts_rx[BATCH_SIZE];
    packet_t      pkts_tx[BATCH_SIZE];
    int           num_tx;
    stats_batch_t stats_batch = {0};

    /* Set CPU affinity if specified */
    if (wctx->cpu_id >= 0) {
#ifdef __linux__
        cpu_set_t cpuset;
        CPU_ZERO(&cpuset);
        CPU_SET(wctx->cpu_id, &cpuset);
        pthread_setaffinity_np(pthread_self(), sizeof(cpuset), &cpuset);
#endif
        reflector_log(LOG_DEBUG, "Worker %d pinned to CPU %d", wctx->worker_id, wctx->cpu_id);
    }

    reflector_log(LOG_INFO, "Worker %d started (queue %d)", wctx->worker_id, wctx->queue_id);

    while (wctx->running) {
        /* Receive batch */
        int rcvd = platform_ops->recv_batch(wctx, pkts_rx, BATCH_SIZE);
        if (rcvd <= 0) {
            continue;
        }

        /* Accumulate RX stats in local batch */
        stats_batch.packets_received += (uint64_t)rcvd;
        for (int i = 0; i < rcvd; i++) {
            stats_batch.bytes_received += pkts_rx[i].len;
        }

        /* Process and reflect ITO packets */
        num_tx = 0;
        for (int i = 0; i < rcvd; i++) {
            /* Prefetch next packet to hide memory latency */
            if (i + 1 < rcvd) {
                PREFETCH_READ(pkts_rx[i + 1].data);
            }

            if (is_ito_packet(pkts_rx[i].data, pkts_rx[i].len, wctx->config)) {
                /* Accumulate signature stats in local batch */
                ito_sig_type_t sig_type = get_ito_signature_type(pkts_rx[i].data, pkts_rx[i].len);
                switch (sig_type) {
                case ITO_SIG_TYPE_PROBEOT:
                    stats_batch.sig_probeot_count++;
                    break;
                case ITO_SIG_TYPE_DATAOT:
                    stats_batch.sig_dataot_count++;
                    break;
                case ITO_SIG_TYPE_LATENCY:
                    stats_batch.sig_latency_count++;
                    break;
                case SIG_TYPE_RFC2544:
                    stats_batch.sig_rfc2544_count++;
                    break;
                case SIG_TYPE_Y1564:
                    stats_batch.sig_y1564_count++;
                    break;
                case SIG_TYPE_MSN:
                    stats_batch.sig_msn_count++;
                    break;
                default:
                    stats_batch.sig_unknown_count++;
                    break;
                }

                if (sig_type == ITO_SIG_TYPE_PROBEOT || sig_type == ITO_SIG_TYPE_DATAOT ||
                    sig_type == ITO_SIG_TYPE_LATENCY) {
                    reflect_netally_packet(pkts_rx[i].data, pkts_rx[i].len,
                                           wctx->config->reflect_mode);
                } else {
                    reflect_packet_with_mode(pkts_rx[i].data, pkts_rx[i].len,
                                             wctx->config->reflect_mode,
                                             wctx->config->software_checksum);
                }

                /* Accumulate latency stats in local batch if enabled */
                if (wctx->config->measure_latency) {
                    uint64_t tx_time    = get_timestamp_ns();
                    uint64_t latency_ns = tx_time - pkts_rx[i].timestamp;

                    /* Update batch latency stats */
                    stats_batch.latency_batch.count++;
                    stats_batch.latency_batch.total_ns += latency_ns;

                    if (stats_batch.latency_batch.count == 1) {
                        stats_batch.latency_batch.min_ns = latency_ns;
                        stats_batch.latency_batch.max_ns = latency_ns;
                    } else {
                        if (latency_ns < stats_batch.latency_batch.min_ns) {
                            stats_batch.latency_batch.min_ns = latency_ns;
                        }
                        if (latency_ns > stats_batch.latency_batch.max_ns) {
                            stats_batch.latency_batch.max_ns = latency_ns;
                        }
                    }
                }

                /* Add to TX batch (stats counted after successful send) */
                pkts_tx[num_tx++] = pkts_rx[i];
            } else {
                /* Not ITO packet, release buffer */
                if (platform_ops->release_batch) {
                    platform_ops->release_batch(wctx, &pkts_rx[i], 1);
                }
            }
        }

        /* Send reflected packets */
        if (num_tx > 0) {
            int sent = platform_ops->send_batch(wctx, pkts_tx, num_tx);
            if (sent < 0) {
                /* Track TX failures in batch */
                stats_batch.err_tx_failed += (uint64_t)num_tx;
            } else if (sent > 0) {
                /* Count ONLY successfully sent packets */
                for (int i = 0; i < sent; i++) {
                    stats_batch.packets_reflected++;
                    stats_batch.bytes_reflected += pkts_tx[i].len;
                }
                /*
                 * Release transmitted buffers back to platform
                 * - AF_PACKET: Returns RX frames to kernel (packets were copied)
                 * - AF_XDP: Triggers CQ polling to recycle UMEM buffers (zero-copy)
                 * - macOS BPF: No-op (packets are copied, no buffer management)
                 */
                if (platform_ops->release_batch) {
                    platform_ops->release_batch(wctx, pkts_tx, sent);
                }
            }
        }

        /* Flush batch to worker stats every BATCH_SIZE packets or periodically */
        stats_batch.batch_count++;
        if (unlikely(stats_batch.batch_count >= STATS_FLUSH_BATCHES)) {
            flush_worker_stats(wctx, &stats_batch);
        }
    }

    /* Final flush before exiting */
    flush_worker_stats(wctx, &stats_batch);

    reflector_log(LOG_INFO, "Worker %d stopped", wctx->worker_id);
#ifndef __APPLE__
    return NULL;
#endif
}

/* Initialize reflector */
int reflector_init(reflector_ctx_t *rctx, const char *ifname)
{
    /* Validate input parameters */
    if (!rctx || !ifname) {
        return -EINVAL;
    }

    memset(rctx, 0, sizeof(*rctx));
    atomic_init(&rctx->running, false);

    /* Set defaults */
#ifdef __APPLE__
    strlcpy(rctx->config.ifname, ifname, MAX_IFNAME_LEN);
#else
    strncpy(rctx->config.ifname, ifname, MAX_IFNAME_LEN - 1);
    rctx->config.ifname[MAX_IFNAME_LEN - 1] = '\0';
#endif
    rctx->config.frame_size        = FRAME_SIZE;
    rctx->config.num_frames        = NUM_FRAMES;
    rctx->config.batch_size        = BATCH_SIZE;
    rctx->config.poll_timeout_ms   = 100;
    rctx->config.cpu_affinity      = -1;    /* Auto: use IRQ affinity */
    rctx->config.use_huge_pages    = false; /* Disabled by default */
    rctx->config.software_checksum = false; /* Use NIC offload by default */
    rctx->config.use_af_xdp        = true;

    /* ITO packet filtering defaults */
    rctx->config.ito_port     = ITO_UDP_PORT; /* Default: port 3842 */
    rctx->config.filter_oui   = true;         /* Filter by OUI by default */
    rctx->config.oui[0]       = NETALLY_OUI_BYTE0;
    rctx->config.oui[1]       = NETALLY_OUI_BYTE1;
    rctx->config.oui[2]       = NETALLY_OUI_BYTE2;
    rctx->config.reflect_mode = REFLECT_MODE_ALL; /* Full reflection by default */

    /* Get interface info */
    rctx->config.ifindex = get_interface_index(ifname);
    if (rctx->config.ifindex < 0) {
        return -1;
    }

    if (get_interface_mac(ifname, rctx->config.mac) < 0) {
        return -1;
    }

    /* Platform selection occurs in reflector_start after callers apply config. */
#ifdef __linux__
    platform_ops = get_packet_platform_ops();
#elif defined(__APPLE__)
    /* macOS BPF has architectural limitations */
    reflector_log(LOG_WARN,
                  "╔════════════════════════════════════════════════════════════════════╗");
    reflector_log(LOG_WARN, "║                   ⚠️  PLATFORM LIMITATION  ⚠️                      ║");
    reflector_log(LOG_WARN,
                  "╠════════════════════════════════════════════════════════════════════╣");
    reflector_log(LOG_WARN,
                  "║ Platform: macOS BPF (Berkeley Packet Filter)                      ║");
    reflector_log(LOG_WARN,
                  "║                                                                    ║");
    reflector_log(LOG_WARN,
                  "║ ARCHITECTURAL LIMIT: 10-50 Mbps maximum throughput                ║");
    reflector_log(LOG_WARN,
                  "║ Linux AF_XDP:        ~10 Gbps (200x faster)                       ║");
    reflector_log(LOG_WARN,
                  "║                                                                    ║");
    reflector_log(LOG_WARN,
                  "║ This is a macOS kernel limitation, not a bug in this software.    ║");
    reflector_log(LOG_WARN,
                  "║ BPF packet processing in userspace is inherently slow.            ║");
    reflector_log(LOG_WARN,
                  "║                                                                    ║");
    reflector_log(LOG_WARN,
                  "║ For high-performance testing (>50 Mbps):                          ║");
    reflector_log(LOG_WARN,
                  "║   • Use Linux with AF_XDP support                                  ║");
    reflector_log(LOG_WARN,
                  "║   • Install libxdp-dev on Ubuntu/Debian                            ║");
    reflector_log(LOG_WARN,
                  "║   • Use physical hardware (not VM)                                 ║");
    reflector_log(LOG_WARN,
                  "║                                                                    ║");
    reflector_log(LOG_WARN,
                  "║ Current macOS suitability:                                         ║");
    reflector_log(LOG_WARN,
                  "║   ✓ Development and debugging                                      ║");
    reflector_log(LOG_WARN,
                  "║   ✓ Low-rate testing (<10 Mbps)                                    ║");
    reflector_log(LOG_WARN,
                  "║   ✗ Production use                                                 ║");
    reflector_log(LOG_WARN,
                  "║   ✗ Performance testing (>50 Mbps)                                 ║");
    reflector_log(LOG_WARN,
                  "╚════════════════════════════════════════════════════════════════════╝");
    platform_ops = get_bpf_platform_ops();
#else
    reflector_log(LOG_ERROR, "Unsupported platform");
    return -1;
#endif

    /* Get number of queues (Linux only) */
#ifdef __linux__
    int num_queues           = get_num_rx_queues(ifname);
    rctx->config.num_workers = num_queues;
#else
    rctx->config.num_workers = 1;
#endif

    reflector_log(LOG_INFO, "Reflector initialized on %s (%d workers)", ifname,
                  rctx->config.num_workers);

    return 0;
}

/* Start reflector workers */
int reflector_start(reflector_ctx_t *rctx)
{
#ifdef __linux__
    if (rctx->config.use_af_xdp) {
#if HAVE_AF_XDP
        platform_ops = get_xdp_platform_ops();
        reflector_log(LOG_INFO, "Platform: AF_XDP (high-performance zero-copy mode)");
#else
        platform_ops = get_packet_platform_ops();
        reflector_log(LOG_WARN, "AF_XDP was requested but is unavailable; using AF_PACKET");
#endif
    } else {
        platform_ops = get_packet_platform_ops();
        reflector_log(LOG_INFO, "Platform: AF_PACKET (explicit selection)");
    }
#endif

    rctx->num_workers       = rctx->config.num_workers;
    rctx->workers           = calloc((size_t)rctx->num_workers, sizeof(worker_ctx_t));
    rctx->platform_contexts = calloc((size_t)rctx->num_workers, sizeof(platform_ctx_t *));
#ifdef __APPLE__
    rctx->worker_queues = calloc((size_t)rctx->num_workers, sizeof(dispatch_queue_t));
    rctx->worker_group  = dispatch_group_create();

    if (!rctx->workers || !rctx->platform_contexts || !rctx->worker_queues || !rctx->worker_group) {
        free(rctx->workers);
        free(rctx->platform_contexts);
        free(rctx->worker_queues);
        if (rctx->worker_group)
            dispatch_release(rctx->worker_group);
        return -ENOMEM;
    }
#else
    rctx->worker_tids = calloc(rctx->num_workers, sizeof(pthread_t));

    if (!rctx->workers || !rctx->platform_contexts || !rctx->worker_tids) {
        free(rctx->workers);
        free(rctx->platform_contexts);
        free(rctx->worker_tids);
        return -ENOMEM;
    }
#endif

    rctx->running = true;

    /* Initialize and start workers */
    for (int i = 0; i < rctx->num_workers; i++) {
        worker_ctx_t *wctx = &rctx->workers[i];
        atomic_init(&wctx->running, true);
        wctx->worker_id = i;
        wctx->queue_id  = i;
        /* Use explicit CPU affinity if configured, otherwise auto-detect from IRQ */
        wctx->cpu_id  = (rctx->config.cpu_affinity >= 0)
                            ? rctx->config.cpu_affinity
                            : get_queue_cpu_affinity(rctx->config.ifname, i);
        wctx->config  = &rctx->config;
        wctx->running = true;
        if (pthread_mutex_init(&wctx->stats_mutex, NULL) != 0) {
            reflector_log(LOG_ERROR, "Failed to initialize stats mutex for worker %d", i);
            reflector_stop(rctx);
            return -1;
        }
        wctx->stats_mutex_initialized = true;

        /* Initialize platform */
        if (platform_ops->init(rctx, wctx) < 0) {
#if defined(__linux__) && HAVE_AF_XDP
            /* Try AF_PACKET fallback on Linux if AF_XDP fails */
            if (platform_ops == get_xdp_platform_ops()) {
                reflector_log(
                    LOG_ERROR,
                    "═══════════════════════════════════════════════════════════════════════");
                reflector_log(
                    LOG_ERROR,
                    "║  🚨 AF_XDP INITIALIZATION FAILED - FALLING BACK TO AF_PACKET 🚨     ║");
                reflector_log(
                    LOG_ERROR,
                    "═══════════════════════════════════════════════════════════════════════");
                reflector_log(LOG_WARN, "");
                reflector_log(
                    LOG_WARN,
                    "╔════════════════════════════════════════════════════════════════════╗");
                reflector_log(
                    LOG_WARN,
                    "║              ⚠️  CRITICAL PERFORMANCE DEGRADATION  ⚠️               ║");
                reflector_log(
                    LOG_WARN,
                    "╠════════════════════════════════════════════════════════════════════╣");
                reflector_log(
                    LOG_WARN,
                    "║ AF_XDP initialization failed - falling back to AF_PACKET           ║");
                reflector_log(
                    LOG_WARN,
                    "║                                                                    ║");
                reflector_log(
                    LOG_WARN,
                    "║ PERFORMANCE IMPACT: 10-100x SLOWER than AF_XDP                     ║");
                reflector_log(
                    LOG_WARN,
                    "║                                                                    ║");
                reflector_log(
                    LOG_WARN,
                    "║ AF_PACKET Performance: ~50-100 Mbps max                            ║");
                reflector_log(
                    LOG_WARN,
                    "║ AF_XDP Performance:    ~10 Gbps (100x faster)                      ║");
                reflector_log(
                    LOG_WARN,
                    "║                                                                    ║");
                reflector_log(
                    LOG_WARN,
                    "║ Common causes:                                                     ║");
                reflector_log(
                    LOG_WARN,
                    "║   • NIC driver doesn't support XDP (check ROADMAP.md)              ║");
                reflector_log(
                    LOG_WARN,
                    "║   • Kernel too old (<5.4 required)                                 ║");
                reflector_log(
                    LOG_WARN,
                    "║   • Insufficient permissions (need CAP_NET_RAW + CAP_BPF)          ║");
                reflector_log(
                    LOG_WARN,
                    "║   • Network interface in use by other process                      ║");
                reflector_log(
                    LOG_WARN,
                    "║                                                                    ║");
                reflector_log(
                    LOG_WARN,
                    "║ Recommended actions:                                               ║");
                reflector_log(LOG_WARN,
                              "║   1. Check NIC compatibility: ethtool -i %s                   ║",
                              rctx->config.ifname);
                reflector_log(
                    LOG_WARN,
                    "║   2. Check kernel: uname -r (need ≥5.4)                            ║");
                reflector_log(
                    LOG_WARN,
                    "║   3. Use Intel/Mellanox NIC for best AF_XDP support               ║");
                reflector_log(
                    LOG_WARN,
                    "║   4. See docs/PERFORMANCE.md for details                           ║");
                reflector_log(
                    LOG_WARN,
                    "║                                                                    ║");
                reflector_log(
                    LOG_WARN,
                    "║ Continuing with AF_PACKET (reduced performance)...                 ║");
                reflector_log(
                    LOG_WARN,
                    "╚════════════════════════════════════════════════════════════════════╝");
                reflector_log(LOG_WARN, "");

                platform_ops = get_packet_platform_ops();
                if (platform_ops->init(rctx, wctx) < 0) {
                    reflector_log(LOG_ERROR, "Failed to initialize AF_PACKET for worker %d", i);
                    reflector_stop(rctx);
                    return -1;
                }
            } else {
                reflector_log(LOG_ERROR, "Failed to initialize platform for worker %d", i);
                reflector_stop(rctx);
                return -1;
            }
#else
            reflector_log(LOG_ERROR, "Failed to initialize platform for worker %d", i);
            reflector_stop(rctx);
            return -1;
#endif
        }

        rctx->platform_contexts[i] = wctx->pctx;
    }

    /* Every queue needs privileged socket setup. Drop privileges only after all
     * platform contexts are ready, then launch the packet workers. */
    if (drop_privileges() < 0) {
        reflector_log(LOG_WARN, "Failed to drop privileges (continuing anyway)");
    }

    for (int i = 0; i < rctx->num_workers; i++) {
        worker_ctx_t *wctx = &rctx->workers[i];

#ifdef __APPLE__
        /* Create GCD queue with QoS for low-latency packet processing */
        char queue_name[64];
        snprintf(queue_name, sizeof(queue_name), "com.reflector.worker%d", i);

        dispatch_queue_attr_t attr = dispatch_queue_attr_make_with_qos_class(
            DISPATCH_QUEUE_SERIAL,
            QOS_CLASS_USER_INTERACTIVE, /* Highest priority for packet processing */
            0                           /* Relative priority within QoS class */
        );

        rctx->worker_queues[i] = dispatch_queue_create(queue_name, attr);
        if (!rctx->worker_queues[i]) {
            reflector_log(LOG_ERROR, "Failed to create GCD queue for worker %d", i);
            reflector_stop(rctx);
            return -1;
        }

        /* Launch worker on GCD queue */
        dispatch_group_enter(rctx->worker_group);
        wctx->thread_started = true;
        dispatch_async(rctx->worker_queues[i], ^{
          worker_loop(wctx);
          dispatch_group_leave(rctx->worker_group);
        });
#else
        /* Create pthread (store TID for joining later) */
        if (pthread_create(&rctx->worker_tids[i], NULL, worker_thread, wctx) != 0) {
            reflector_log(LOG_ERROR, "Failed to create worker thread %d", i);
            reflector_stop(rctx);
            return -1;
        }
        wctx->thread_started = true;
#endif
    }

    reflector_log(LOG_INFO, "Reflector started with %d workers", rctx->num_workers);
    return 0;
}

/* Stop reflector */
void reflector_stop(reflector_ctx_t *rctx)
{
    rctx->running = false;

    if (rctx->workers) {
        for (int i = 0; i < rctx->num_workers; i++) {
            rctx->workers[i].running = false;
        }

#ifdef __APPLE__
        /* Wait for all GCD workers to finish */
        if (rctx->worker_group) {
            dispatch_group_wait(rctx->worker_group, DISPATCH_TIME_FOREVER);
        }
#else
        /* Wait for all pthread workers to exit */
        if (rctx->worker_tids) {
            for (int i = 0; i < rctx->num_workers; i++) {
                if (rctx->workers[i].thread_started) {
                    pthread_join(rctx->worker_tids[i], NULL);
                }
            }
        }
#endif

        reflector_get_stats(rctx, &rctx->global_stats);

        /* Cleanup platform contexts */
        for (int i = 0; i < rctx->num_workers; i++) {
            if (rctx->workers[i].pctx && platform_ops && platform_ops->cleanup) {
                platform_ops->cleanup(&rctx->workers[i]);
            }
            if (rctx->workers[i].stats_mutex_initialized) {
                pthread_mutex_destroy(&rctx->workers[i].stats_mutex);
            }
        }

#ifdef __APPLE__
        /* Release GCD resources */
        if (rctx->worker_queues) {
            for (int i = 0; i < rctx->num_workers; i++) {
                if (rctx->worker_queues[i]) {
                    dispatch_release(rctx->worker_queues[i]);
                }
            }
            free(rctx->worker_queues);
            rctx->worker_queues = NULL;
        }
        if (rctx->worker_group) {
            dispatch_release(rctx->worker_group);
            rctx->worker_group = NULL;
        }
#else
        free(rctx->worker_tids);
        rctx->worker_tids = NULL;
#endif

        free(rctx->workers);
        free(rctx->platform_contexts);
        rctx->workers           = NULL;
        rctx->platform_contexts = NULL;
    }

    reflector_log(LOG_INFO, "Reflector stopped");
}

/* Cleanup reflector */
void reflector_cleanup(reflector_ctx_t *rctx)
{
    if (rctx->running) {
        reflector_stop(rctx);
    }
}

/* Atomic load helper for 64-bit values (thread-safe stats reading) */
#define ATOMIC_LOAD64(x) __atomic_load_n(&(x), __ATOMIC_RELAXED)

/* Get aggregated statistics (thread-safe) */
void reflector_get_stats(const reflector_ctx_t *rctx, reflector_stats_t *stats)
{
    if (!rctx->workers) {
        *stats = rctx->global_stats;
        return;
    }
    memset(stats, 0, sizeof(*stats));

    for (int i = 0; i < rctx->num_workers; i++) {
        worker_ctx_t *worker = &rctx->workers[i];
        if (!worker->stats_mutex_initialized) {
            continue;
        }
        const reflector_stats_t *ws = &worker->stats;
        pthread_mutex_lock(&worker->stats_mutex);

        /* Basic packet counters - use atomic loads for thread safety */
        stats->packets_received += ATOMIC_LOAD64(ws->packets_received);
        stats->packets_reflected += ATOMIC_LOAD64(ws->packets_reflected);
        stats->packets_dropped += ATOMIC_LOAD64(ws->packets_dropped);
        stats->bytes_received += ATOMIC_LOAD64(ws->bytes_received);
        stats->bytes_reflected += ATOMIC_LOAD64(ws->bytes_reflected);

        /* Per-signature counters */
        stats->sig_probeot_count += ATOMIC_LOAD64(ws->sig_probeot_count);
        stats->sig_dataot_count += ATOMIC_LOAD64(ws->sig_dataot_count);
        stats->sig_latency_count += ATOMIC_LOAD64(ws->sig_latency_count);
        stats->sig_rfc2544_count += ATOMIC_LOAD64(ws->sig_rfc2544_count);
        stats->sig_y1564_count += ATOMIC_LOAD64(ws->sig_y1564_count);
        stats->sig_msn_count += ATOMIC_LOAD64(ws->sig_msn_count);
        stats->sig_unknown_count += ATOMIC_LOAD64(ws->sig_unknown_count);

        /* Error counters */
        stats->err_invalid_mac += ATOMIC_LOAD64(ws->err_invalid_mac);
        stats->err_invalid_ethertype += ATOMIC_LOAD64(ws->err_invalid_ethertype);
        stats->err_invalid_protocol += ATOMIC_LOAD64(ws->err_invalid_protocol);
        stats->err_invalid_signature += ATOMIC_LOAD64(ws->err_invalid_signature);
        stats->err_too_short += ATOMIC_LOAD64(ws->err_too_short);
        stats->err_tx_failed += ATOMIC_LOAD64(ws->err_tx_failed);
        stats->err_nomem += ATOMIC_LOAD64(ws->err_nomem);

        /* Legacy error counters */
        stats->rx_invalid += ATOMIC_LOAD64(ws->rx_invalid);
        stats->rx_nomem += ATOMIC_LOAD64(ws->rx_nomem);
        stats->tx_errors += ATOMIC_LOAD64(ws->tx_errors);

        /* Aggregate latency statistics */
        uint64_t lat_count = ATOMIC_LOAD64(ws->latency.count);
        if (lat_count > 0) {
            stats->latency.count += lat_count;
            stats->latency.total_ns += ATOMIC_LOAD64(ws->latency.total_ns);

            uint64_t lat_min = ATOMIC_LOAD64(ws->latency.min_ns);
            uint64_t lat_max = ATOMIC_LOAD64(ws->latency.max_ns);

            /* Update min/max across all workers */
            if (stats->latency.count == lat_count) {
                /* First worker with latency data */
                stats->latency.min_ns = lat_min;
                stats->latency.max_ns = lat_max;
            } else {
                if (lat_min < stats->latency.min_ns) {
                    stats->latency.min_ns = lat_min;
                }
                if (lat_max > stats->latency.max_ns) {
                    stats->latency.max_ns = lat_max;
                }
            }
        }
        pthread_mutex_unlock(&worker->stats_mutex);
    }

    /* Calculate average latency */
    if (stats->latency.count > 0) {
        stats->latency.avg_ns = (double)stats->latency.total_ns / (double)stats->latency.count;
    }
}

/* Reset statistics */
void reflector_reset_stats(reflector_ctx_t *rctx)
{
    if (!rctx) {
        return;
    }
    if (rctx->workers) {
        for (int i = 0; i < rctx->num_workers; i++) {
            if (!rctx->workers[i].stats_mutex_initialized) {
                continue;
            }
            pthread_mutex_lock(&rctx->workers[i].stats_mutex);
            memset(&rctx->workers[i].stats, 0, sizeof(reflector_stats_t));
            pthread_mutex_unlock(&rctx->workers[i].stats_mutex);
        }
    }
    memset(&rctx->global_stats, 0, sizeof(reflector_stats_t));
}

/* Set configuration */
int reflector_set_config(reflector_ctx_t *rctx, const reflector_config_t *config)
{
    if (!rctx || !config) {
        return -1;
    }

    /* Cannot change config while running */
    if (rctx->running) {
        reflector_log(LOG_ERROR, "Cannot change configuration while running");
        return -1;
    }

    memcpy(&rctx->config, config, sizeof(reflector_config_t));

    /* Cap num_workers to reasonable maximum */
    if (rctx->config.num_workers > MAX_WORKERS) {
        reflector_log(LOG_WARN, "Capping num_workers from %d to %d", rctx->config.num_workers,
                      MAX_WORKERS);
        rctx->config.num_workers = MAX_WORKERS;
    }
    if (rctx->config.num_workers < 1) {
        rctx->config.num_workers = 1;
    }

    return 0;
}

int reflector_update_filter(reflector_ctx_t *rctx, uint16_t port, sig_filter_t signature_filter)
{
    if (!rctx) {
        return -EINVAL;
    }
    if (rctx->running) {
        return -EBUSY;
    }
    rctx->config.ito_port   = port;
    rctx->config.sig_filter = signature_filter;
    return 0;
}

/* Get configuration */
void reflector_get_config(const reflector_ctx_t *rctx, reflector_config_t *config)
{
    if (!rctx || !config) {
        return;
    }

    memcpy(config, &rctx->config, sizeof(reflector_config_t));
}

const platform_ops_t *get_platform_ops(void)
{
    return platform_ops;
}
