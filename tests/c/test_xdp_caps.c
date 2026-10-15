/*
 * test_xdp_caps.c - the AF_XDP capability check names what is missing (#1747)
 *
 * Stem loads its own XDP program, so CAP_BPF, CAP_NET_ADMIN and CAP_NET_RAW
 * are enough, and CAP_SYS_ADMIN stands in for the first two as in the kernel.
 * stem_xdp_caps_check() must accept exactly the sets the kernel accepts. Root
 * runs (`sudo make c-test`) exercise each set by lowering the effective
 * capabilities; an unprivileged run checks the refusal only.
 */

#include <stdint.h>

#include "../../include/stem_errno.h"
#include "../../include/stem_xdp_caps.h"
#include "test_framework.h"

typedef struct __user_cap_data_struct cap_data_t[_LINUX_CAPABILITY_U32S_3];

static struct __user_cap_header_struct cap_header(void)
{
    struct __user_cap_header_struct hdr = {.version = _LINUX_CAPABILITY_VERSION_3, .pid = 0};
    return hdr;
}

static int read_caps(cap_data_t data)
{
    struct __user_cap_header_struct hdr = cap_header();
    return (int)syscall(SYS_capget, &hdr, data);
}

static int write_caps(const cap_data_t data)
{
    struct __user_cap_header_struct hdr = cap_header();
    return (int)syscall(SYS_capset, &hdr, data);
}

/* Lower the effective set to `caps` (each must be permitted) and run the check. */
static int check_with(const cap_data_t permitted, const unsigned *caps, int ncaps)
{
    cap_data_t data;
    memcpy(data, permitted, sizeof(data));
    data[0].effective = 0;
    data[1].effective = 0;
    for (int i = 0; i < ncaps; i++) {
        data[caps[i] / 32].effective |= UINT32_C(1) << (caps[i] % 32);
    }
    if (write_caps(data) != 0) {
        return -1000 - errno;
    }
    int ret = stem_xdp_caps_check();
    (void)write_caps(permitted);
    return ret;
}

static cap_data_t original;

/* The set the packaged service runs with. */
TEST(accepts_bpf_without_sys_admin)
{
    const unsigned caps[] = {CAP_NET_RAW, CAP_NET_ADMIN, CAP_BPF};
    ASSERT_EQ(0, check_with(original, caps, 3));
}

TEST(accepts_sys_admin_for_bpf_and_net_admin)
{
    const unsigned caps[] = {CAP_NET_RAW, CAP_SYS_ADMIN};
    ASSERT_EQ(0, check_with(original, caps, 2));
}

TEST(refuses_missing_bpf)
{
    const unsigned caps[] = {CAP_NET_RAW, CAP_NET_ADMIN, CAP_PERFMON};
    ASSERT_EQ(-EPERM, check_with(original, caps, 3));
}

TEST(refuses_missing_net_admin)
{
    const unsigned caps[] = {CAP_NET_RAW, CAP_BPF};
    ASSERT_EQ(-EPERM, check_with(original, caps, 2));
}

TEST(refuses_missing_net_raw)
{
    const unsigned caps[] = {CAP_NET_ADMIN, CAP_BPF, CAP_SYS_ADMIN};
    ASSERT_EQ(-EPERM, check_with(original, caps, 3));
}

TEST(refuses_unprivileged)
{
    ASSERT_EQ(-EPERM, stem_xdp_caps_check());
}

int main(void)
{
    TEST_SUITE("AF_XDP capability check");
    if (read_caps(original) != 0) {
        printf("capget: %s\n", stem_strerror(errno));
        return EXIT_FAILURE;
    }
    bool privileged = stem_cap_effective(original, CAP_SYS_ADMIN) &&
                      stem_cap_effective(original, CAP_BPF) &&
                      stem_cap_effective(original, CAP_PERFMON);
    if (privileged) {
        RUN_TEST(accepts_bpf_without_sys_admin);
        RUN_TEST(accepts_sys_admin_for_bpf_and_net_admin);
        RUN_TEST(refuses_missing_bpf);
        RUN_TEST(refuses_missing_net_admin);
        RUN_TEST(refuses_missing_net_raw);
    } else {
        printf("not root: checking the unprivileged refusal only\n");
        RUN_TEST(refuses_unprivileged);
    }
    TEST_SUMMARY();
    return TEST_EXIT_STATUS();
}
