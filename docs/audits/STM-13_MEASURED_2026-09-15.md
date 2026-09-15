# A throughput result is a measurement — #1233 on the wire, 2026-09-15

**Question.** `STM-13_FALLBACK_2026-09-15.md` proved an unmodified binary can
run an RFC 2544 throughput search (#1232), and the same run reported
**9990.23 Mbps / 14,866,419 pps / 99.90 % of line rate over a link that
carried ~6.7 k frames in ~70 s**. Whether any throughput number this product
prints is a measurement was filed as
[#1233](https://github.com/MustardSeedNetworks/stem/issues/1233). This is the
fix and its proof.

**Answer.** The three `max_rate_*` fields are now the best trial's own achieved
figures, the rate the binary search settled on is reported separately as
`offered_rate_pct`, and a trial that falls more than 10 % short of its offered
load is flagged `generator_limited` and stops the search. The reported rate now
agrees with a packet capture of the same run.

| Claim | Result | Evidence |
| --- | --- | --- |
| Before the fix, the reported rate is the offered rate | **REPRODUCED** | scratch `red_probe` (the three reporting lines restored to their old derivation, not merged): `measured 7440476 pps / 5000.00 Mbps / 50.0000%` for a trial that achieved 103 pps → `FAIL: reported pps is the offered load, not a measurement`, `make c-test` exit 2 |
| After the fix, the reported rate is the achieved rate | **PASS** | `test_throughput_measured`: `offered 50.00% (7440476 pps demanded), measured 103 pps / 0.05 Mbps / 0.0007%, generator_limited=true, iterations=1`, exit 0 |
| The whole C suite still passes | **PASS** | `make c-test` green in `golang:1.27.0-trixie` (6 tests) |
| The cgo boundary carries the new fields | **PROVEN** | a real `CGO_ENABLED=1` run through `(*Context).RunThroughputTest` returns `"MaxRatePPS": 101.03, "OfferedRatePct": 50, "GeneratorLimited": true` |
| The reported rate matches the wire | **PROVEN** | same run: `tcpdump -i eth0 -nn "udp port 3842"` counted **125 frames, 0 dropped by kernel**, over the ~1.24 s the trial plus its straggler wait occupied — **≈101 pps**, against `MaxRatePPS: 101.03` |
| The generator can reach the offered load | **NO — filed as [#1239](https://github.com/MustardSeedNetworks/stem/issues/1239)** | the ceiling is ~100 pps on `lo` as well as on `eth0`, so it is not the medium |

## The fix

`rfc2544_throughput_test` judged each trial on loss alone. A generator that
cannot reach the offered load still gets back every frame it managed to send,
so loss reads 0.000 % at every step and the binary search walks the rate to the
top of its range on evidence it does not have. `max_rate_pct` was then the
_configured_ rate it settled on, and `max_rate_mbps` / `max_rate_pps` were
derived from that rate and the line rate — none of the three ever looked at
what left the interface, although `run_trial` had been computing
`achieved_pps` / `achieved_mbps` correctly all along and the caller discarded
them.

Three changes, all in `rfc2544_throughput_test`:

- the reported rate is the best trial's `achieved_pps` / `achieved_mbps`, and
  `max_rate_pct` is that rate over the theoretical maximum for the frame size
  (both sides wire-size based, so the percentage and the pps agree);
- `offered_rate_pct` carries the search's own answer, so demanded and achieved
  are comparable rather than conflated;
- a trial whose achieved load falls more than `RFC2544_GENERATOR_TOLERANCE`
  (10 %) short of the demanded load sets `generator_limited` and **breaks**.
  Continuing would spend nine more iterations re-measuring the same ceiling,
  and RFC 2544 §26.1 requires the offered load actually be generated, so the
  honest answer is already in hand after that trial.

**Why clamp-and-flag rather than failing the run.** Refusing a
generator-limited run would make every honest low-rate host red — including
the only lab-free bench that exists — and would tell an operator less than the
measured number does. The numbers are the truth; the flag is the warning. What
must never happen again is printing a rate the wire did not carry, and that is
what the flag plus the achieved figures prevent. The tolerance is deliberately
generous and is not tied to `resolution_pct` (0.1–1 %): no measurement of how
close AF_XDP on production hardware comes to its offered load exists yet, and
a tight threshold would flag genuine runs.

## What the reported numbers look like now

The same result shape the issue quoted, from a real cgo run on `eth0` against
an unreflected peer:

```json
{
  "FrameSize": 64,
  "MaxRatePct": 0.0006789538431448653,
  "MaxRateMbps": 0.05172981529627786,
  "MaxRatePPS": 101.0347955005427,
  "OfferedRatePct": 50,
  "GeneratorLimited": true,
  "Iterations": 1
}
```

`stem test` renders whatever fields a result carries, so `OfferedRatePct` and
`GeneratorLimited` reach the text, JSON and CSV output without a CLI change;
`rfc2544_print_results` gained an Offered column and a `GENERATOR LIMITED`
marker for the same reason.

## The new defect this exposes, and why it is worse than it looks

The generator tops out at **~100 pps on loopback**, which rules out the
medium: `test_throughput_measured` runs on `lo` and reports 103–104 pps, the
`eth0` run reports 101. Diagnosis: `packet_recv_batch`
(`src/dataplane/linux_packet/packet_platform.c:298`) sets `SO_RCVTIMEO` to
1 ms and then calls a **blocking** `recvmsg` — the socket is never
`O_NONBLOCK` — and `run_trial`'s hot loop calls it after every single frame.
The kernel rounds `SO_RCVTIMEO` up to whole jiffies, and this kernel is
`CONFIG_HZ=250`, so each timed-out call costs 4 ms; two per iteration is the
~9.7 ms that 103 pps implies. Transmission is also one `sendto` per frame with
no batching, so the AF_PACKET path could not approach 10 Gbps even with the
blocking receive removed.

This matters beyond throughput: `run_trial_custom` shares the same loop, so
Y.1564, Y.1731, MEF and TSN inherit the ceiling. And since #1232 made
AF_PACKET the runtime fallback, it is the path taken on every host without
working AF_XDP — CT307's LXC guest included. AF_XDP's own `recv_batch` polls
with a zero timeout and is not subject to this.

Filed as **[#1239](https://github.com/MustardSeedNetworks/stem/issues/1239)**
rather than folded in: it changes hot-loop receive behaviour,
it is measured by the `C Performance` A/B gate, and folding it in would make
this PR's wire proof a conflation of two changes.

## Bench

One Apple container, `golang:1.27.0-trixie` plus
`libxdp-dev libbpf-dev tcpdump iproute2 procps curl` — gcc-14, because gcc-12
cannot compile this tree (`-std=c23`). `CGO_ENABLED=1` linux/arm64,
`make dataplane` then a throwaway probe in-module (internal packages are not
importable from outside, so it cannot live in `/tmp`; removed afterwards).
Adding the two fields meant editing `include/rfc2544.h` **and** the cgo
preamble, which redeclares that struct rather than including the header — a
silent layout mismatch one unpaired edit away, filed as
[#1240](https://github.com/MustardSeedNetworks/stem/issues/1240).
AF_XDP cannot create a UMEM in the container, so every run here takes the
AF_PACKET fallback from #1232. The container was destroyed afterwards.

## What this does not close

**ST-2 clause 3** (a Playwright suite run against a licensed real daemon) is
still open, and #1233 was its second blocker. One remains:
[#1231](https://github.com/MustardSeedNetworks/stem/issues/1231) — the first
reflector run setuids the whole daemon to `nobody`, so a harness that starts
and stops a reflector twice gets `EPERM` behind a bare 500. The clause's own
wording still needs the amendment #1078 asked about, but the shape of it has
changed: a per-step measurement assertion is now _possible_, because a step
that measures nothing reports a measured ~0 with `generatorLimited` set rather
than 99.90 % of line rate.

**STM-13's RFC 2544 / Y.1564 legs** remain blocked. A run now reports honestly,
but reporting ~100 pps honestly is not an RFC 2544 result; the generator defect
above has to be fixed before any per-platform throughput leg is worth taking,
on the dev servers or anywhere.

`Latency.Count: 0` in a successful throughput result is untouched here: it is a
different mechanism (`measure_latency` unset in the config the caller builds),
noted on the issue rather than fixed in a reporting change.
