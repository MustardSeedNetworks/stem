# A forwarding rate is a measurement — #1242, 2026-09-24

**Question.** [#1233](https://github.com/MustardSeedNetworks/stem/issues/1233)
fixed RFC 2544 throughput reporting the offered rate as measured
(`STM-13_MEASURED_2026-09-15.md`). The same defect in the RFC 2889
forwarding-rate search was filed as
[#1242](https://github.com/MustardSeedNetworks/stem/issues/1242). This is the
fix and its proof.

**Answer.** `max_rate_fps`, `max_rate_pct` and `aggregate_rate_mbps` are now
derived from the best trial's achieved frame rate, the rate the binary search
settled on is reported separately as `offered_rate_pct`, and a trial more than
`RFC2544_GENERATOR_TOLERANCE` (10 %) short of its demanded load sets
`generator_limited` and stops the search. The cgo preambles were also found to
declare the result struct eight bytes short of the header, so the C side had
been writing past the result Go allocated.

| Claim | Result | Evidence |
| --- | --- | --- |
| Before the fix, the reported rate is the offered rate | **REPRODUCED** | header fields added, `rfc2889.c` unchanged: `Forwarding Rate: 99.22% (14764695 fps, 9921.87 Mbps)` for a search that transmitted `frames_tx=6161879` over seven 1 s trials → `FAIL: reported fps exceeds the frames the search transmitted`, `make c-test` exit 2 |
| After the fix, the reported rate is the achieved rate | **PASS** | `test_rfc2889_forwarding_measured`: `offered 50.00% (7440476 fps demanded), measured 737735 fps / 495.76 Mbps / 4.9576%, generator_limited=true, frames_tx=840700` |
| The test catches the defect, not just a build break | **PASS** | two mutants, each applied alone: dropping the `break` → `FAIL: the search must stop at the first generator-limited trial`; restoring `max_rate_fps = max_pps * best_rate / 100.0` → `FAIL: reported fps exceeds the frames the search transmitted`; both `make c-test` exit 2 |
| The whole C suite still passes | **PASS** | `make c-test` exit 0 in `golang:1.27.0-trixie` (9 tests) |
| The cgo boundary carries the new fields | **PROVEN** | a real `CGO_ENABLED=1` run through `(*Context).RunRFC2889ForwardingTest` on `eth0` returns `"MaxRateFps": 719858.05, "OfferedRatePct": 50, "GeneratorLimited": true, "FramesTx": 823748` |
| The reported frame count matches the wire | **PROVEN** | same run: `tcpdump -i eth0 -nn "udp port 3842"`, 0 dropped by kernel, counted **847,534** frames after the 2 s warmup against `FramesTx` **823,748** (+2.9 %) |
| The preamble layout matched the header | **NO — fixed here** | `sizeof` of the struct as the cgo preambles declared it on `origin/main`: **56 bytes**; as `include/rfc2544.h` declared it: **64 bytes** (`loss_pct` missing) |

## The fix

`rfc2889_forwarding_test` (`src/dataplane/common/rfc2889.c`) judged each trial
on loss alone. Nothing in it read `trial.achieved_pps`, although `run_trial`
computes it, so a generator that could not reach the offered load read 0 %
loss at every step and the search walked its rate to the top of the range. All
three rate fields were then derived from that configured rate.

The change mirrors #1241 exactly: the reported frame rate is the best trial's
`achieved_pps`; `max_rate_pct` is that over the theoretical maximum for the
frame size; `aggregate_rate_mbps` keeps its existing wire-size definition
(`(frame_size + 20) * 8`), now applied to the measured rate; and a
generator-limited trial breaks the search, because every further iteration
re-measures the same ceiling. Clamp-and-flag rather than fail, for the reasons
`STM-13_MEASURED_2026-09-15.md` gives.

## Reading the wire comparison

Compare frame **counts**, not rates. `MaxRateFps` is `FramesTx` divided by the
trial's elapsed time, and elapsed includes the straggler wait after the send
loop — 1.14 s here, the same accounting #1241's audit recorded at 1.24 s. So
719,858 fps sits below the 811–837 k frames per wall-second the capture shows,
by construction and in the conservative direction. The frame count is the
like-for-like figure. The capture's warmup cut is taken at the first captured
frame plus 2 s and is approximate to the timer boundary, which is the +2.9 %.

The default 2 s warmup is in the run because the Go config treats a zero
`WarmupSec` as "use the default"; that is existing behaviour, untouched here.

## The layout defect

Two cgo preambles (`dataplane.go` and `dataplane_rfc2889.go`) redeclare
`rfc2889_fwd_result_t` rather than including the header, and both omitted the
header's final `loss_pct`. `rfc2889_forwarding_test` begins with
`memset(result, 0, sizeof(*result))` using the header's 64 bytes and later
stores `loss_pct` at offset 56, both into a 56-byte object Go allocated. Both
preambles now match the header field for field. Replacing the redeclaration
with an include, so this cannot recur silently, is
[#1240](https://github.com/MustardSeedNetworks/stem/issues/1240) (D-STEM-25).

## Bench

One Apple container, `golang:1.27.0-trixie` (gcc 14) plus
`libxdp-dev libbpf-dev tcpdump iproute2 libpcap-dev`, `CGO_ENABLED=1`
linux/arm64. AF_XDP cannot create a UMEM in the container, so every run took
the AF_PACKET fallback. The cgo run used a throwaway in-package test probe,
removed afterwards; the peer was the container's gateway, which does not
reflect, so `FramesRx` is 0 and the generator shortfall is the only thing the
result can report. The container was destroyed afterwards.
