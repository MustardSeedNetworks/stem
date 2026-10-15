/*
 * Privileges an AF_XDP socket needs, checked before anything is loaded.
 *
 * Stem loads its own XDP program (stem_xsk_prog.h): creating the XSKMAP and
 * loading the program need CAP_BPF and CAP_NET_ADMIN, the socket needs
 * CAP_NET_RAW. CAP_SYS_ADMIN stands in for the first two, as in the kernel.
 * Without the check the first refusal is a bare EPERM from the map create,
 * which names nothing (stem#1747).
 */

#ifndef STEM_XDP_CAPS_H
#define STEM_XDP_CAPS_H

#include <errno.h>
#include <stdbool.h>
#include <stdint.h>

#include <linux/capability.h>
#include <sys/syscall.h>

#include <unistd.h>

#define STEM_XDP_CAPS_NEEDED "AF_XDP needs CAP_BPF, CAP_NET_ADMIN and CAP_NET_RAW"

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
    bool sys_admin = stem_cap_effective(data, CAP_SYS_ADMIN);
    bool bpf       = sys_admin || stem_cap_effective(data, CAP_BPF);
    bool net_admin = sys_admin || stem_cap_effective(data, CAP_NET_ADMIN);
    return bpf && net_admin && stem_cap_effective(data, CAP_NET_RAW) ? 0 : -EPERM;
}

#endif /* STEM_XDP_CAPS_H */
