/*
 * test_packet_platform_init.c - the AF_PACKET worker init and its failure exits.
 *
 * On 2026-09-12 a double free in `packet_platform_init`'s failure-exit paths
 * reached CT307: the function published `pctx` on `wctx->pctx` and then, on a
 * failure, called `free(pctx)` without clearing the field, so `reflector_stop`'s
 * cleanup loop (`core.c`: `if (workers[i].pctx) ops->cleanup(&workers[i])`)
 * freed it a second time. The gate could not see it because ADR 0008 scoped
 * ASAN to `src/dataplane/common/packet.c`. ADR 0008's 2026-09-14 amendment
 * widened that scope; this is its unit half (stem#1228, plan row STM-22).
 *
 * The contract every exit must hold: after `packet_platform_init` returns < 0,
 * `wctx->pctx` is NULL, and running the cleanup loop afterwards is a no-op.
 * Each case asserts that and then runs the loop, so the regression fails twice
 * over — on the assertion without a sanitizer, and as a use-after-free read
 * plus a double free under ASAN.
 *
 * The syscalls the init path makes are wrapped (`-Wl,--wrap=socket,--wrap=bind,
 * --wrap=setsockopt`) rather than issued for real. That is not convenience:
 * AF_PACKET needs CAP_NET_RAW, which the CI runner does not have and a local
 * container does, so an unwrapped test would exercise a different exit in each
 * place and none of them on demand. `mmap` is deliberately NOT wrapped — the
 * ASAN runtime uses it — so the ring mapping fails against the substitute fd
 * and the init takes its documented simple-mode path.
 *
 * Copyright (c) 2026 Mustard Seed Networks. All rights reserved.
 */

#include <errno.h>
#include <string.h>

#include <linux/if_packet.h>
#include <sys/socket.h>

#include "reflector.h"
#include "test_framework.h"

int  packet_platform_init(reflector_ctx_t *rctx, worker_ctx_t *wctx);
void packet_platform_cleanup(worker_ctx_t *wctx);

/* ============================================================================
 * Syscall injection
 * ============================================================================ */

typedef enum {
    FAIL_NONE = 0,
    FAIL_PACKET_SOCKET,   /* socket(AF_PACKET, ...) */
    FAIL_IGNORE_OUTGOING, /* setsockopt(PACKET_IGNORE_OUTGOING) */
    FAIL_TPACKET_RING,    /* setsockopt(PACKET_RX_RING) — fails V3 and the V2 retry */
    FAIL_PACKET_BIND,     /* bind() of the AF_PACKET socket */
    FAIL_UDP_GUARD,       /* the guard socket's SO_BINDTODEVICE */
} fail_point_t;

static fail_point_t g_fail = FAIL_NONE;

int __real_socket(int domain, int type, int protocol);
int __real_setsockopt(int fd, int level, int optname, const void *val, socklen_t len);
int __real_bind(int fd, const struct sockaddr *addr, socklen_t len);

/*
 * AF_PACKET is answered with a real but harmless UDP socket: the init path
 * closes it, maps against it and hands it to setsockopt, so it must be a live
 * descriptor rather than a number.
 */
int __wrap_socket(int domain, int type, int protocol)
{
    if (domain == AF_PACKET) {
        if (g_fail == FAIL_PACKET_SOCKET) {
            errno = EPERM;
            return -1;
        }
        return __real_socket(AF_INET, SOCK_DGRAM, 0);
    }
    return __real_socket(domain, type, protocol);
}

int __wrap_setsockopt(int fd, int level, int optname, const void *val, socklen_t len)
{
    if (level == SOL_PACKET) {
        if (optname == PACKET_IGNORE_OUTGOING && g_fail == FAIL_IGNORE_OUTGOING) {
            errno = ENOPROTOOPT;
            return -1;
        }
        if (optname == PACKET_RX_RING && g_fail == FAIL_TPACKET_RING) {
            errno = EINVAL;
            return -1;
        }
        return 0; /* PACKET_VERSION, QDISC_BYPASS, FANOUT: accepted */
    }
    if (level == SOL_SOCKET && optname == SO_BINDTODEVICE) {
        if (g_fail == FAIL_UDP_GUARD) {
            errno = EPERM;
            return -1;
        }
        return 0;
    }
    if (level == SOL_SOCKET && optname == SO_ATTACH_FILTER) {
        return 0;
    }
    return __real_setsockopt(fd, level, optname, val, len);
}

int __wrap_bind(int fd, const struct sockaddr *addr, socklen_t len)
{
    if (addr->sa_family == AF_PACKET) {
        if (g_fail == FAIL_PACKET_BIND) {
            errno = ENODEV;
            return -1;
        }
        return 0;
    }
    if (addr->sa_family == AF_INET) {
        return 0; /* the UDP guard, bound to a port the test does not own */
    }
    return __real_bind(fd, addr, len);
}

/* ============================================================================
 * Fixtures
 * ============================================================================ */

static reflector_config_t g_config;
static reflector_ctx_t    g_rctx;
static worker_ctx_t       g_wctx;

static void fixture_reset(void)
{
    memset(&g_config, 0, sizeof(g_config));
    memset(&g_rctx, 0, sizeof(g_rctx));
    memset(&g_wctx, 0, sizeof(g_wctx));

    snprintf(g_config.ifname, sizeof(g_config.ifname), "lo");
    g_config.ifindex    = 1;
    g_config.ito_port   = 3842;
    g_config.sig_filter = SIG_FILTER_ALL;

    g_rctx.config      = g_config;
    g_rctx.num_workers = 1;
    g_wctx.config      = &g_rctx.config;
}

/*
 * What `reflector_stop` does with a worker after a failed start. On the
 * contract this is a no-op; on the CT307 regression it is a use-after-free
 * read followed by a second free.
 */
static void run_reflector_stop_cleanup_loop(void)
{
    if (g_wctx.pctx) {
        packet_platform_cleanup(&g_wctx);
    }
}

static void assert_failure_exit_is_clean(fail_point_t where, const char *what)
{
    fixture_reset();
    g_fail = where;

    int rc = packet_platform_init(&g_rctx, &g_wctx);
    g_fail = FAIL_NONE;

    printf("  %s\n", what);

    /*
     * Order matters. The cleanup loop runs before the assertion so that the
     * regression is caught by ASAN as well as by the assertion: an assertion
     * that returned first would leave the second free unexecuted, and the
     * sanitizer half of this gate would prove nothing.
     */
    const void *published = g_wctx.pctx;
    run_reflector_stop_cleanup_loop();

    ASSERT_TRUE(rc < 0);
    /* The CT307 double free: a failed init that leaves pctx published. */
    ASSERT_NULL(published);
}

/* ============================================================================
 * Tests
 * ============================================================================ */

TEST(init_failure_exits_leave_no_dangling_context)
{
    assert_failure_exit_is_clean(FAIL_PACKET_SOCKET, "socket(AF_PACKET) refused");
    assert_failure_exit_is_clean(FAIL_IGNORE_OUTGOING, "PACKET_IGNORE_OUTGOING refused");
    assert_failure_exit_is_clean(FAIL_TPACKET_RING, "neither TPACKET_V3 nor V2 ring available");
    assert_failure_exit_is_clean(FAIL_PACKET_BIND, "bind to the interface refused");
    assert_failure_exit_is_clean(FAIL_UDP_GUARD, "UDP guard port unavailable");
}

/*
 * The guard socket is only opened for worker 0 under a signature filter that
 * reflects UDP, and an unset port there is a configuration error, not a
 * syscall failure — so it is the one exit reached with no injection at all.
 */
TEST(missing_guard_port_is_a_clean_failure)
{
    fixture_reset();
    g_config.ito_port      = 0;
    g_rctx.config.ito_port = 0;

    int rc = packet_platform_init(&g_rctx, &g_wctx);

    const void *published = g_wctx.pctx;
    run_reflector_stop_cleanup_loop();

    ASSERT_TRUE(rc < 0);
    ASSERT_NULL(published);
}

TEST(successful_init_publishes_a_context_that_cleanup_releases)
{
    fixture_reset();

    int rc = packet_platform_init(&g_rctx, &g_wctx);

    ASSERT_EQ(0, rc);
    ASSERT_NOT_NULL(g_wctx.pctx);

    packet_platform_cleanup(&g_wctx);
    ASSERT_NULL(g_wctx.pctx);

    /* Idempotent: the same loop runs again on a reflector that already stopped. */
    run_reflector_stop_cleanup_loop();
}

int main(void)
{
    TEST_SUITE("AF_PACKET worker init (stem#1228)");
    RUN_TEST(init_failure_exits_leave_no_dangling_context);
    RUN_TEST(missing_guard_port_is_a_clean_failure);
    RUN_TEST(successful_init_publishes_a_context_that_cleanup_releases);
    TEST_SUMMARY();
    return TEST_EXIT_STATUS();
}
