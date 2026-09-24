# The AF_PACKET generator is no longer held to ~100 pps — #1239, 2026-09-24

**Question.** `STM-13_MEASURED_2026-09-15.md` made every throughput number a
measurement, and the measurement was ~100 pps on `lo` and on `eth0` alike.
[#1239](https://github.com/MustardSeedNetworks/stem/issues/1239) traced the
ceiling to the AF_PACKET receive path. This is the fix and its proof.

**Answer.** `packet_recv_batch` now receives with `MSG_DONTWAIT` and no longer
sets `SO_RCVTIMEO`. On the same two-container bench the RFC 2544 generator goes
from **110 pps to 1,072,799 pps** at 64 bytes, the sender's NIC counter agrees
with the reported rate, and at 1280 bytes the reflector's capture agrees with it
within 0.3 %. Y.1564's step loop, which calls the same `recv_batch`, goes from a
flat ~125 pps at every step to the demanded rate.

| Claim | Result | Evidence |
| --- | --- | --- |
| Before the fix, the generator is held to ~100 pps | **REPRODUCED** | `test_throughput_measured` with a new 10,000 pps floor: `measured 103 pps` → `FAIL: the generator is still held to the blocking-receive ceiling (#1239)`, `make c-test` exit 2 |
| After the fix, the floor holds on `lo` | **PASS** | same test: `measured 755290 pps`, then 736,690 / 757,390 / 783,695 on three reruns; `make c-test` exit 0 |
| RFC 2544 between two containers: ≥ 100× | **PASS — ~9,700×** | before `MaxRatePPS 110.82`, after `MaxRatePPS 1072798.74` (64 bytes, same bench, same reflector) |
| The reported rate is what left the sender | **PROVEN** | 64 bytes: master `tx_packets +5,506,667`, `tx_dropped +0`, qdisc `dropped 0`, against `MaxRatePPS 1,075,541` over the ~5.1 s trial |
| The reported rate is what reached the reflector | **PROVEN at 1280 bytes** | reflector `tcpdump`: **1,778,669 captured, 0 dropped by kernel**; master `tx_packets +1,783,607`; reported `MaxRatePPS 347,947` → 1,778,669 / 5.126 s = 346,984 pps |
| Y.1564 shows the same gain | **PASS (C probe)** | 100 Mbps CIR, 512 bytes, 2 s steps on `lo`: before 250 / 250 / 250 / 249 frames per step; after 11,749 / 23,495 / 35,243 / 46,993 |
| The daemon runs Y.1564 | **NO — filed as [#1411](https://github.com/MustardSeedNetworks/stem/issues/1411)** | `y1564_config` and `y1564_perf` both fail `-22` before sending a frame |
| C memory-safety gate | **PASS** | `make c-test-asan CC=clang`, `make c-fuzz c-fuzz-reflector FUZZ_SECONDS=20`, `make lint-c` — all exit 0 |

## The fix

`run_trial`, `run_trial_custom` and `y1564_run_step` each send one frame and
then poll `recv_batch` for returned frames. On AF_PACKET that poll was a
**blocking** `recvmsg` behind a 1 ms `SO_RCVTIMEO`, re-applied with a
`setsockopt` on every call. The kernel rounds a socket timeout up to a whole
jiffy, so an empty queue — the normal case between frames — cost 4 ms at
`CONFIG_HZ=250` and 10 ms at `CONFIG_HZ=100`, on every frame. That is the
~100 pps ceiling, and it is independent of the medium.

`recvmsg(..., MSG_DONTWAIT)` returns at once on an empty queue, and the
`setsockopt` goes with the timeout. Nothing relied on the wait: every caller
already sleeps between polls where it wants to wait (the 10 × 10 ms straggler
loops), and AF_XDP's `recv_batch` has always polled with a zero timeout.
Fixing it in the platform means all four loops inherit it; no caller changed.

**Not changed: batched transmission.** The issue also names one `sendto` per
frame. Every caller passes `count == 1`, so `sendmmsg` inside
`packet_send_batch` would have no consumer, and batching at the callers means
stamping a batch of buffers instead of the one shared template. The acceptance
did not need it, and the measured ceiling above says where the path now sits:
~1.07 M pps at 64 bytes (~7 % of 10 GbE line rate), generator-limited and
flagged as such. A host that has to reach 64-byte line rate on AF_PACKET needs
that change, measured on its own.

## Bench

Three Apple containers. The binary is a `CGO_ENABLED=1` linux/arm64 build made
in `golang:1.27.0-trixie` (gcc-14, `libxdp-dev libbpf-dev`) from `5d90c02` with
the fix; the "before" binary is the same tree with `origin/main`'s
`packet_platform.c`. The UI was built on the host first so `//go:embed` has
something to read. Two `debian:trixie-slim` containers run the daemons
(`libxdp1 libbpf1 tcpdump iproute2 procps curl jq`): `192.168.64.3` the
reflector (`stem reflect -i eth0 --profile all`), `192.168.64.4` the test master
with a fresh 14-day trial. Both have MTU 1280, which is why the large-frame run
is 1280 bytes, not 1518. AF_XDP cannot create a UMEM in these containers, so
every run takes the AF_PACKET fallback — the path this fix is about. Throwaway
credentials; all three containers were destroyed afterwards.

Runs were started with `POST /api/v1/test/start`, not `stem test`: the CLI
refuses any RFC 2544 run that includes a 64-byte frame, the default included,
because it attaches a Y.1564 block to every step and the daemon rejects that
block below 70 bytes — filed as
[#1412](https://github.com/MustardSeedNetworks/stem/issues/1412).

```text
# fixed, 64 bytes
{"FrameSize":64,"MaxRatePct":7.2092,"MaxRateMbps":549.27,"MaxRatePPS":1072798.74,
 "OfferedRatePct":50,"GeneratorLimited":true,"Iterations":1}
# fixed, 1280 bytes
{"FrameSize":1280,"MaxRatePct":36.19,"MaxRateMbps":3562.98,"MaxRatePPS":347947.41,
 "OfferedRatePct":50,"GeneratorLimited":true,"Iterations":1}
reflector: 1778669 packets captured, 1778669 received by filter, 0 dropped by kernel
master:    tx_packets +1783607, tx_dropped +0

# before (origin/main's packet_platform.c), same bench
{"FrameSize":64,"MaxRatePPS":110.82,"GeneratorLimited":true,"Iterations":1}
reflector: 580 packets captured, 0 dropped by kernel; master tx_packets +592
{"FrameSize":1280,"MaxRatePPS":110.31,"GeneratorLimited":true,"Iterations":1}
reflector: 577 packets captured, 0 dropped by kernel; master tx_packets +586
```

## Why the 64-byte reflector count is not the agreement

At 64 bytes the master's NIC sent 5,506,667 frames and the reflector's NIC
received 1,210,396 (`rx_dropped +0`, `tcpdump` 0 dropped): the virtual link
between the two Apple container VMs carried about 22 % of the offered load, and
the reflector then logged `txErrors 2,041,040` trying to return what it did
receive. Neither is Stem's generator: the sender-side counter matches the
reported rate to the frame, and at 1280 bytes, a load the link carries, the
reflector's capture matches it too. This bench can say how fast the generator
is; it cannot say what a 64-byte path between two hosts carries. The
reflector's own return-path errors at that rate are worth their own look when
STM-13 reaches real hardware.

## Y.1564: the probe, and why it is a probe

Every Y.1564 run through the daemon fails before transmitting
([#1411](https://github.com/MustardSeedNetworks/stem/issues/1411)):
`y1564_run_step` reads a platform that nothing on the daemon path has
prepared, and `ctx->config.y1564` is never filled from Go, so the step table
is all zeros (`Step 1: 0% CIR`). So the Y.1564 clause was proved with a scratch
C probe (not merged) that does both by hand — one RFC 2544 trial to prepare the
platform, `y1564_default_config` for the steps — and then calls
`y1564_config_test` on `lo` with a 100 Mbps CIR at 512 bytes and 2-second
steps, built once against each `packet_platform.c`:

```text
fixed    step 1  25%: 22.702 Mbps, frames_tx 11749
         step 2  50%: 45.114 Mbps, frames_tx 23495
         step 3  75%: 67.761 Mbps, frames_tx 35243
         step 4 100%: 90.128 Mbps, frames_tx 46993
before   step 1  25%:  0.462 Mbps, frames_tx 250
         step 2  50%:  0.461 Mbps, frames_tx 250
         step 3  75%:  0.462 Mbps, frames_tx 250
         step 4 100%:  0.463 Mbps, frames_tx 249
```

The fixed frame counts are the demanded rate (23,496 pps at 100 Mbps and
512 bytes, × 2 s = 46,993); the old ones are the ceiling at every step.

## What this does not close

- **STM-13's throughput legs** are unblocked by this defect, but on this bench
  a Y.1564 leg still needs #1411, and a CLI leg needs #1412.
- **ST-2** still needs [#1231](https://github.com/MustardSeedNetworks/stem/issues/1231)
  (D-STEM-23): the second reflector start in one daemon lifetime failed with
  `start reflector: failed to start reflector` behind a 500 on this bench too.
