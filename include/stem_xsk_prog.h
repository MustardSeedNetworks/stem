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
 * nothing has to be found on disk at run time. It has no pointer comparison,
 * which the verifier refuses without CAP_PERFMON. The attachment is a BPF link:
 * the kernel detaches the program when the link's descriptor closes, so a
 * crashed run never leaves it on the interface.
 */

#ifndef STEM_XSK_PROG_H
#define STEM_XSK_PROG_H

#include <stddef.h>
#include <stdint.h>

#include <linux/bpf.h>

#include <bpf/bpf.h>
#include <unistd.h>

typedef struct {
    int map_fd;  /* XSKMAP: receive queue -> AF_XDP socket */
    int link_fd; /* holds the attachment; closing it detaches */
} stem_xsk_prog_t;

#define STEM_XSK_PROG_NONE ((stem_xsk_prog_t){.map_fd = -1, .link_fd = -1})

/* Load the redirect program for `queues` receive queues and attach it to
 * `ifindex`. On failure nothing stays attached and the negated errno is
 * returned (libbpf returns it). */
static inline int stem_xsk_prog_attach(stem_xsk_prog_t *prog, int ifindex, uint32_t queues)
{
    *prog = STEM_XSK_PROG_NONE;

    int map_fd = bpf_map_create(BPF_MAP_TYPE_XSKMAP, "stem_xsks", sizeof(uint32_t),
                                sizeof(uint32_t), queues, NULL);
    if (map_fd < 0) {
        return map_fd;
    }

    /* return bpf_redirect_map(&xsks, ctx->rx_queue_index, XDP_PASS); a queue
     * with no socket passes its frames to the stack. */
    const struct bpf_insn insns[] = {
        {.code    = BPF_LDX | BPF_MEM | BPF_W,
         .dst_reg = BPF_REG_2,
         .src_reg = BPF_REG_1,
         .off     = offsetof(struct xdp_md, rx_queue_index)},
        {.code    = BPF_LD | BPF_IMM | BPF_DW,
         .dst_reg = BPF_REG_1,
         .src_reg = BPF_PSEUDO_MAP_FD,
         .imm     = map_fd},
        {0},
        {.code = BPF_ALU64 | BPF_MOV | BPF_K, .dst_reg = BPF_REG_3, .imm = XDP_PASS},
        {.code = BPF_JMP | BPF_CALL, .imm = BPF_FUNC_redirect_map},
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

#endif /* STEM_XSK_PROG_H */
