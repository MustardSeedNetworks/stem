# STM-11: Stem versus the EtherScope nXG — blocked before first measurement

**Date:** 2026-09-09
**Stem build:** `v0.24.83-3-g33bef38` (`make build` on dev-srv-ubuntu, Ubuntu 26.04)
**Outcome:** no comparison was made. Stem cannot be aimed at a reflector, so
there is no stem result to compare against the tester. Filed as issue #1127.

## What this row asked for

Run RFC 2544 throughput/latency/loss and Y.1564 against the CT307 reflector from
a Linux host and from the EtherScope nXG on the same path in the same minute,
upload the tester's result to Link-Live, and record the deltas and the tolerance
the v1 plan will hold.

## Lab state at the time of the attempt

| Node | Address | Role | State |
| ---------- | ------ | ------- | ------- |
| CT307 | 10.44.40.23 (VLAN 40) | reflector | up; `/usr/bin/stem reflect --interface eth0 --profile netally --port 3842` running (pid 102) |
| EtherScope nXG | 10.44.10.31 (VLAN 10) | reference tester | up, answers ICMP |
| dev-srv-ubuntu | 10.44.30.30 (VLAN 30) | stem test host | up; `sudo` available, tcpdump present |

Both preconditions the row named were satisfied — the EtherScope was reachable
this time, unlike 2026-09-07 — so the blocker below is in the product, not the
lab.

## The blocker

`stem test` has no destination. The RFC 2544 dataplane keeps its peer in
`rfc2544_ctx_t.remote_mac` / `remote_ip` (`include/rfc2544_internal.h:54,56`),
and those fields are read in three places and **written nowhere in the tree**
(audited with `git ls-files`, not `rg -l`):

```text
$ git ls-files | xargs grep -n "remote_mac\|remote_ip\|RemoteMac\|remoteMac"
include/rfc2544_internal.h:54:    uint8_t  remote_mac[6];
include/rfc2544_internal.h:56:    uint32_t remote_ip;
src/dataplane/common/core.c:357:        memcpy(dst_mac, ctx->remote_mac, 6);
src/dataplane/common/core.c:370:        *dst_ip = ctx->remote_ip;
src/dataplane/common/core.c:837:    if (ctx->remote_mac[0] || ctx->remote_mac[1] || ctx->remote_mac[2]) {
src/dataplane/common/core.c:838:        memcpy(dst_mac, ctx->remote_mac, 6);
src/dataplane/common/core.c:1064:    if (ctx->remote_mac[0] || ctx->remote_mac[1] || ctx->remote_mac[2]) {
src/dataplane/common/core.c:1065:        memcpy(dst_mac, ctx->remote_mac, 6);
```

The context is `calloc`'d, so the guard is always false and the hardcoded
placeholders stand — `02:00:00:00:00:02` and `10.0.0.2`, carrying the comment
"Default addresses - in real use, would be configured" (`core.c:827-839`, and
identically at 1054-1065). `local_mac` _is_ filled from the NIC, so only the
destination is fictional. No entry point exposes one: not the `stem test` flag
set, not the API request types (only `TrafficGenTestConfig` has a `dstMac`), not
the standalone C test master's usage text.

Observed, twice, on the real path:

```text
$ sudo ./bin/stem test -i ens18 -t rfc2544_throughput -d 5 --frame-sizes 512 --warmup 1
The Stem v0.24.83-3-g33bef38 - Network Testing
Interface:    ens18
Tests:        rfc2544_throughput
[INFO] Platform: AF_XDP (high performance)
[xdp] Initialized on ens18 queue 0 (fd=4)
[INFO] Throughput result: 0.00% (0.00 Mbps, 0 pps)
  Max Rate:    0.00% (0.00 Mbps, 0 pps)
  Iterations:  10
  Latency:     min=0.00us avg=0.00us max=0.00us
```

Ten binary-search iterations, every trial losing everything, reported as a
**success-shaped 0.00 %** with no error and no warning that there is no peer.
A `tcpdump` on CT307's `eth0` filtered for the placeholder MAC, `10.0.0.2` and
UDP 3842 captured nothing while the test ran. The same filter on the sending
side captured nothing either, which is expected rather than informative: the
AF_XDP path bypasses the kernel capture hook.

## Second finding: "the same path" does not exist yet

The three nodes sit on three different VLANs — dev-srv-ubuntu on 30, CT307 on
40, the EtherScope on 10 — and `ip route get 10.44.40.23` from dev-srv-ubuntu
resolves `via 10.44.30.1 dev ens18`. Stem's traffic would cross the lab router;
the EtherScope's would cross a different pair of router interfaces. Even once #1127
is fixed, a delta measured this way would be measuring the router, not the
two products. Before STM-11 is retried, either a Linux stem host is placed on
the EtherScope's segment, or both testers are moved onto CT307's, and the
topology actually used is recorded with the numbers.

## What has to be true before this row can run again

1. Issue #1127 fixed: a destination on `stem test`, on the API request and on the
   module params, written through to `remote_mac`/`remote_ip`, with next-hop
   resolution for the routed case, and a typed error — never a 0.00 % result —
   when no destination is given.
2. A single L2 path shared by stem's host and the EtherScope, recorded here.
3. Only then the paired same-minute runs, the Link-Live upload of the tester's
   result, and the tolerance this plan will hold.

Nothing was run on the EtherScope or uploaded to Link-Live: with no stem number
to compare, a tester run would have produced a figure with no counterpart.
CT307 was not modified; the only commands issued against it were `tcpdump` and
`pgrep`.
