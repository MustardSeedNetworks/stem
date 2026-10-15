/*
 * Privileges an AF_XDP socket needs, checked before libxdp runs.
 *
 * libxdp attaches its program with CAP_NET_ADMIN and then looks the attached
 * program up by id (BPF_PROG_GET_FD_BY_ID), which the kernel allows only with
 * CAP_SYS_ADMIN. CAP_BPF and CAP_PERFMON do not cover that lookup: with them
 * libxdp fails its dispatcher check, attaches its default program without a
 * dispatcher, leaves it attached, and xsk_socket__create() reports ENOENT,
 * which names nothing (stem#1747). Checking first gives the operator the real
 * cause and attaches nothing.
 */

#ifndef STEM_XDP_CAPS_H
#define STEM_XDP_CAPS_H

#include <errno.h>
#include <stdbool.h>
#include <stdint.h>

#include <linux/capability.h>
#include <sys/syscall.h>

#include <unistd.h>

#define STEM_XDP_CAPS_NEEDED "AF_XDP needs CAP_SYS_ADMIN, CAP_NET_ADMIN and CAP_NET_RAW"

static inline bool stem_cap_effective(const struct __user_cap_data_struct *data, unsigned cap)
{
    return (data[cap / 32].effective & (UINT32_C(1) << (cap % 32))) != 0;
}

/* 0 when this thread may open an AF_XDP socket, -EPERM when a capability is
 * missing, or the negated capget() errno. */
static inline int stem_xdp_caps_check(void)
{
    struct __user_cap_header_struct hdr = {.version = _LINUX_CAPABILITY_VERSION_3, .pid = 0};
    struct __user_cap_data_struct   data[_LINUX_CAPABILITY_U32S_3] = {0};

    if (syscall(SYS_capget, &hdr, data) != 0) {
        return -errno;
    }
    return stem_cap_effective(data, CAP_SYS_ADMIN) && stem_cap_effective(data, CAP_NET_ADMIN) &&
                   stem_cap_effective(data, CAP_NET_RAW)
               ? 0
               : -EPERM;
}

#endif /* STEM_XDP_CAPS_H */
