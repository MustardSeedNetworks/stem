/*
 * The XDP program that hands each receive queue's frames to its AF_XDP socket.
 *
 * Stem loads and attaches this program itself and gives libxdp
 * XSK_LIBXDP_FLAGS__INHIBIT_PROG_LOAD. libxdp's own loader attaches a program
 * and then looks it up by id (BPF_PROG_GET_FD_BY_ID), which the kernel allows
 * only with CAP_SYS_ADMIN; holding the descriptors from the load needs no
 * lookup, so CAP_BPF, CAP_NET_ADMIN and CAP_NET_RAW are enough (stem#1747).
 *
 * The program is built from instructions rather than a compiled object, so
 * nothing has to be found on disk at run time. Its only pointer comparison is
 * the packet bounds check, the one kind the verifier allows without
 * CAP_PERFMON. The attachment is a BPF link: the kernel detaches the program
 * when the link's descriptor closes, so a crashed run never leaves it on the
 * interface.
 *
 * A frame redirected to a socket never reaches the kernel stack, so the
 * program redirects only unfragmented IPv4 UDP with a 20-byte IP header to one
 * port: the frames a reflector answers, or the replies a test master counts.
 * ARP, SSH, DHCP and the rest go to the stack. Redirecting everything took the
 * host's own traffic and cut it off the network (stem#1767).
 */

#ifndef STEM_XSK_PROG_H
#define STEM_XSK_PROG_H

#include <stddef.h>
#include <stdint.h>

#include <arpa/inet.h>
#include <linux/bpf.h>

#include <bpf/bpf.h>
#include <unistd.h>

typedef struct {
    int map_fd;  /* XSKMAP: receive queue -> AF_XDP socket */
    int link_fd; /* holds the attachment; closing it detaches */
} stem_xsk_prog_t;

#define STEM_XSK_PROG_NONE ((stem_xsk_prog_t){.map_fd = -1, .link_fd = -1})

#define STEM_XSK_LDX(size, dst, src, offset) \
    {.code = BPF_LDX | BPF_MEM | (size), .dst_reg = (dst), .src_reg = (src), .off = (offset)}
#define STEM_XSK_JNE(dst, value, to) \
    {.code = BPF_JMP | BPF_JNE | BPF_K, .dst_reg = (dst), .off = (to), .imm = (value)}

/* Load the redirect program for `queues` receive queues and attach it to
 * `ifindex`, redirecting the test frames addressed to `udp_port`, which must
 * not be 0. On failure nothing stays attached and the negated errno is
 * returned (libbpf returns it). */
static inline int stem_xsk_prog_attach(stem_xsk_prog_t *prog, int ifindex, uint32_t queues,
                                       uint16_t udp_port)
{
    *prog = STEM_XSK_PROG_NONE;

    int map_fd = bpf_map_create(BPF_MAP_TYPE_XSKMAP, "stem_xsks", sizeof(uint32_t),
                                sizeof(uint32_t), queues, NULL);
    if (map_fd < 0) {
        return map_fd;
    }

    /* Loads read the packet's network-order bytes in host order, so each
     * field is compared with its htons() value. Every jump lands on `pass`,
     * the last two instructions, and counts from the instruction after it. */
    const struct bpf_insn insns[] = {
        STEM_XSK_LDX(BPF_W, BPF_REG_7, BPF_REG_1, offsetof(struct xdp_md, data)),
        STEM_XSK_LDX(BPF_W, BPF_REG_8, BPF_REG_1, offsetof(struct xdp_md, data_end)),
        {.code = BPF_ALU64 | BPF_MOV | BPF_X, .dst_reg = BPF_REG_2, .src_reg = BPF_REG_7},
        /* BPF_ADD and BPF_K are both 0, so the immediate form is just this. */
        {.code = BPF_ALU64 | BPF_ADD, .dst_reg = BPF_REG_2, .imm = 14 + 20 + 8},
        {.code = BPF_JMP | BPF_JGT | BPF_X, .dst_reg = BPF_REG_2, .src_reg = BPF_REG_8, .off = 17},
        STEM_XSK_LDX(BPF_H, BPF_REG_2, BPF_REG_7, 12), /* EtherType */
        STEM_XSK_JNE(BPF_REG_2, htons(0x0800), 15),
        STEM_XSK_LDX(BPF_B, BPF_REG_2, BPF_REG_7, 14), /* version 4, 20-byte header */
        STEM_XSK_JNE(BPF_REG_2, 0x45, 13),
        STEM_XSK_LDX(BPF_B, BPF_REG_2, BPF_REG_7, 23), /* protocol */
        STEM_XSK_JNE(BPF_REG_2, IPPROTO_UDP, 11),
        STEM_XSK_LDX(BPF_H, BPF_REG_2, BPF_REG_7, 20), /* a later fragment has no UDP header */
        {.code = BPF_ALU64 | BPF_AND | BPF_K, .dst_reg = BPF_REG_2, .imm = htons(0x1fff)},
        STEM_XSK_JNE(BPF_REG_2, 0, 8),
        STEM_XSK_LDX(BPF_H, BPF_REG_2, BPF_REG_7, 36), /* UDP destination port */
        STEM_XSK_JNE(BPF_REG_2, htons(udp_port), 6),
        /* return bpf_redirect_map(&xsks, ctx->rx_queue_index, XDP_PASS); a
         * queue with no socket passes its frames to the stack. */
        STEM_XSK_LDX(BPF_W, BPF_REG_2, BPF_REG_1, offsetof(struct xdp_md, rx_queue_index)),
        {.code    = BPF_LD | BPF_IMM | BPF_DW,
         .dst_reg = BPF_REG_1,
         .src_reg = BPF_PSEUDO_MAP_FD,
         .imm     = map_fd},
        {0},
        {.code = BPF_ALU64 | BPF_MOV | BPF_K, .dst_reg = BPF_REG_3, .imm = XDP_PASS},
        {.code = BPF_JMP | BPF_CALL, .imm = BPF_FUNC_redirect_map},
        {.code = BPF_JMP | BPF_EXIT},
        /* pass: */
        {.code = BPF_ALU64 | BPF_MOV | BPF_K, .dst_reg = BPF_REG_0, .imm = XDP_PASS},
        {.code = BPF_JMP | BPF_EXIT},
    };
    int prog_fd = bpf_prog_load(BPF_PROG_TYPE_XDP, "stem_xsk", "Proprietary", insns,
                                sizeof(insns) / sizeof(insns[0]), NULL);
    if (prog_fd < 0) {
        close(map_fd);
        return prog_fd;
    }

    /* No mode flag: the kernel attaches in driver mode where the driver has
     * XDP and generic mode where it does not. */
    int link_fd = bpf_link_create(prog_fd, ifindex, BPF_XDP, NULL);
    close(prog_fd);
    if (link_fd < 0) {
        close(map_fd);
        return link_fd;
    }

    prog->map_fd  = map_fd;
    prog->link_fd = link_fd;
    return 0;
}

/* Deliver `queue`'s frames to the AF_XDP socket `xsk_fd`. */
static inline int stem_xsk_prog_add(const stem_xsk_prog_t *prog, uint32_t queue, int xsk_fd)
{
    uint32_t fd = (uint32_t)xsk_fd;
    return bpf_map_update_elem(prog->map_fd, &queue, &fd, BPF_ANY);
}

static inline void stem_xsk_prog_detach(stem_xsk_prog_t *prog)
{
    if (prog->link_fd >= 0) {
        close(prog->link_fd);
    }
    if (prog->map_fd >= 0) {
        close(prog->map_fd);
    }
    *prog = STEM_XSK_PROG_NONE;
}

#undef STEM_XSK_LDX
#undef STEM_XSK_JNE

#endif /* STEM_XSK_PROG_H */
