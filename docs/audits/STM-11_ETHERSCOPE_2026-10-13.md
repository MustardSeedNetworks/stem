# STM-11: Stem versus the EtherScope nXG, 2026-10-13 (routed)

**Stem build:** v0.26.20. The native binary is from the released
`stem_0.26.20_amd64.deb`, extracted with `dpkg-deb -x` and not installed. It
ran as a loopback-bound daemon on dev-srv-ubuntu with a trial licence.
**Reference:** EtherScope nXG 10.44.10.31, Performance app, endpoint type
Reflector, UDP 3842.
**Reflector:** CT307 `stem-reflector` (10.44.40.23), stem 0.26.6,
daemon-owned reflector on `eth0`.

The comparison is **routed, not same-path** (owner decision 2026-10-02). Both
testers reach CT307 through the lab router, from different VLANs.

**Outcome.** At fixed rates below the path's capacity, the two testers agree
on throughput and loss. They disagree on delay and delay variation. The
EtherScope measures 0.43–0.50 ms latency and 115–226 µs jitter. Stem measures
a 0.35–0.40 ms minimum, a 1.3–1.4 ms mean and 21–26 ms FDV. Stem's RFC 2544
latency step does not run at the throughput rate. Under a shared or saturated
queue, stem gets fewer frames through than the EtherScope does. Four stem
issues are filed: #1680, #1681, #1682 and #1683. The EtherScope results are
in Link-Live; their IDs are in the last section.

## Paths and idle baselines

| Tester | Address | Hops to CT307 (10.44.40.23) | Idle baseline |
| --- | --- | --- | --- |
| stem (dev-srv-ubuntu `ens18`, virtio) | 10.44.30.30, VLAN 30 | `tracepath -n`: 10.44.30.1 (0.652–0.771 ms), then 10.44.40.23 (1.181 ms); pmtu 1500, 2 hops | `ping -c 20 -i 0.2`: min/avg/max/mdev 0.556/0.771/0.958/0.121 ms, 0 % loss |
| EtherScope nXG (wired, 1 Gb FDx) | 10.44.10.31, VLAN 10 | Path app: ICX1750-24 1/1/22 → 1/3/1 (VLAN 10, 10 Gb uplink), `msnlab` router br10 → br40, then `stem-reflector.msn.lab`; "Destination reached in 2 hops" | Path app ICMP to CT307: 82, 4, 1 ms (the first probe includes resolution); router hop 6, 11, 14 ms (control-plane ICMP) |

CT307's veth (`net0`, `bridge=vmbr0,tag=40`) has no `rate=` limit.
Subtracting the baseline changes nothing material at these rates. The router
hop answers ICMP slowly from its control plane, so ICMP is a poor baseline for
the EtherScope's side. The fixed-rate latencies below are the better floor for
it.

## The reflector had to admit both payloads

CT307's reflector is configured `profile: netally`, `signatureFilter: ["ito"]`.
That admits the EtherScope's `DATA:OT`/`PROBEOT` frames and drops stem's own
`RFC254` frames. The first stem run (512 B throughput) therefore failed with
`throughput test received no frames back from the peer`. That typed refusal
is correct behaviour, not a 0 % "success". `stem test` cannot send an ITO
payload.

For the comparison window, the reflector was switched to `profile: all`
through its own daemon API, from inside CT307. The daemon's local descriptor
was used and no credential left the container. The reflector was restored
afterwards to the byte-identical original:

```text
before: {"profile":"netally","signatureFilter":["ito"],"ouiFilter":"","portFilter":3842,"interface":"eth0","autostart":true}
during: {"profile":"all","signatureFilter":["all"],"ouiFilter":"","portFilter":3842,"interface":"eth0","autostart":true}
after:  {"profile":"netally","signatureFilter":["ito"],"ouiFilter":"","portFilter":3842,"interface":"eth0","autostart":true}
```

CT307 was not redeployed or restarted, and its stem version was unchanged.

## Same-minute pairs (512 B, 60 s, Y.1564 performance phase vs Performance app)

Stem ran Y.1564 through the daemon API with 5 s configuration steps and a
60 s performance phase. The EtherScope START was tapped when the daemon
logged `Service Performance Test: service=1`. Both testers count rates at L1
(frame + 20 B preamble and IFG): at 10 Mbps stem offered 2,347 fps and the
EtherScope 2,350 fps.

| Pair | Start (UTC) stem / ES | Offered | Stem: frames tx/rx, FLR | Stem FD avg / min / max | Stem FDV | ES throughput | ES loss | ES latency | ES jitter |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2 | 16:22:41.8 / 16:22:42 | 10 Mbps each | 140,843 / 140,843, 0 % | 1.40 / 0.399 / 21.34 ms | 20.94 ms | 9.999 Mbps | 0 % | 432 µs | 226 µs |
| 4 | 16:33:34.1 / 16:33:34 | 25 Mbps each | 351,233 / 344,930, **1.79 %** | 1.29 / 0.348 / 26.76 ms | 26.41 ms | 24.99 Mbps | 0 % | 504 µs | 115 µs |
| 3 | 16:26:13.0 / 16:26:17 | 100 Mbps each (overload) | 1,408,962 / 937,750, 33.4 % | 230.2 / 210.1 / 264.0 ms | 53.9 ms | 73.52 Mbps | 26.3 % | 26 ms | 28 µs |

Pair 1 (16:15:29 / 16:15:33, 10 Mbps) is excluded. The CLI's `-d 60` also set
each of the four Y.1564 configuration steps to 60 s. Only the 2.5 Mbps step
overlapped the EtherScope's minute, so that pair measured different rates. In
that pair the EtherScope reported 9.999 Mbps, 0 %, 401 µs and 231 µs.

Pair 3's EtherScope test ran about 86 s, not 60. A mis-tap earlier had set
the test duration to 30 minutes, so the test was stopped by hand. Pair 3 is
included only to show contention: two testers each offering 100 Mbps into a
path that carries about 74–88 Mbps measure each other.

## Solo runs (each tester alone, back to back)

| Run | Time (UTC) | Result |
| --- | --- | --- |
| EtherScope, 100 Mbps, 60 s | 16:28:49 | 87.68 Mbps, 12.2 % loss, 85 ms latency, 19 µs jitter; "thresholds exceeded (6)" |
| stem Y.1564, CIR 100 Mbps | 16:30:08 | perf phase: 1,409,414 tx / 1,013,590 rx (≈ 71.9 Mbps L1), FLR 28.1 %, FD 238.4 ms. Config steps: 23.5 Mbps 0 %, 46.4 Mbps 0 %, 68.3 Mbps 0 % with FD avg 19.1 ms, 94.3 Mbps 26.0 % |
| stem Y.1564, CIR 25 Mbps | 16:35:03 | 350,963 / 350,963, FLR 0 %, FD 1.41 / 0.423 / 24.59 ms, FDV 24.16 ms |
| stem `rfc2544_throughput`, 512 B, 10 s trials, 1 % resolution | 16:37:01 | 18,055 pps = **73.95 Mbps (L2)**, 76.8 Mbps at L1; 7 iterations; `GeneratorLimited: false` |
| stem `rfc2544_latency`, 512 B | same run | 10–100 % of an assumed 10 Gbps line rate (the virtio NIC reports no speed); every step ~210–260 ms, all queueing. See #1680 |
| stem `rfc2544_frame_loss`, 512 B | same run | 96.7–97.3 % loss at 10–100 % of 10 Gbps. RFC 2544 §26.3 starts at the medium's maximum rate, so this is expected on a 74 Mbps path |

The EtherScope Performance app is a fixed-rate service test. The unit has no
RFC 2544 binary search, so stem's RFC 2544 throughput is compared with the
EtherScope's solo capacity point. At 100 Mbps the EtherScope got 87.68 Mbps
through. Stem's zero-loss rate is 76.8 Mbps (L1), and under the same overload
stem got 71.9 Mbps through.

## Deltas and where each one lands

| Metric | Stem vs EtherScope | Verdict |
| --- | --- | --- |
| Achieved rate at a fixed CIR | 2,347 vs 2,350 fps (−0.13 %); 9.999 / 24.99 Mbps on both | **agree** |
| Loss at 10 Mbps, same minute | 0 vs 0 | **agree** |
| Loss at 25 Mbps, same minute | 1.79 % vs 0; stem alone at 25 Mbps 0 % | **stem gap**: only stem's frames die at a shared queue, #1683 |
| Capacity under overload (solo) | 71.9 vs 87.68 Mbps through | **stem gap**, #1683 |
| Latency floor | stem FD min 0.399 / 0.348 ms vs ES 0.432 / 0.504 ms | **agree** within 0.16 ms |
| Mean latency | stem FD avg 1.40 / 1.29 ms vs ES 0.432 / 0.504 ms (×2.6–3.2) | **stem gap**: a 20–27 ms tail that ICMP never shows, #1681 |
| Delay variation | stem FDV (max − min) 20.9 / 26.4 ms vs ES jitter 0.226 / 0.115 ms | **stem gap**: the same tail plus a max − min definition, #1681. The definitions also differ, so this is not a like-for-like delta |
| RFC 2544 latency | ~230 ms at every load | **stem gap**: not run at the throughput rate (§26.2), #1680 |
| CLI output | a successful run prints no measurements; `--json` is prefixed with the banner | **stem gap**, #1682 |

Nothing here was filed against the EtherScope or NetAlly. Every disagreement
above traces to stem.

## Tolerance this plan holds

These are measured same minute, routed, 512 B, against CT307, with combined
offered load at or below 25 % of the path's measured RFC 2544 throughput
(about 18 Mbps on this path). Anything above that measures contention, not
the product.

| Metric | Tolerance | Status at v0.26.20 |
| --- | --- | --- |
| Achieved rate at a fixed CIR | within ±1 % of the EtherScope | PASS (−0.13 %) |
| FLR | 0 whenever the EtherScope reports 0 | PASS at 10 Mbps |
| FD minimum | within ±0.2 ms of the EtherScope's latency | PASS (0.033 and 0.156 ms) |
| FD mean | ≤ EtherScope latency + 0.25 ms | **FAIL** (+0.97 ms), #1681 |
| FDV | ≤ EtherScope jitter + 1 ms, once stem defines FDV the way MEF 10.3 does | **FAIL** (+20.7 ms), #1681 |
| Overload capacity, solo, back to back | within 10 % of the EtherScope's through-rate at the same offered load | **FAIL** (−18 %), #1683 |

STM-12 and STM-13 reuse this table. A row passes when its rerun meets every
line.

## Link-Live results

The EtherScope uploaded each Performance run to Link-Live (organization
Gotlabs.net, unit `Kris's EtherScope nXG - 536204`). The IDs were read on
2026-10-14 through the Link-Live API (`/v1/admin/results`,
`resultType: performance`). Each one is matched to a run above by its
`testStartTime`, to the second. Every result's average latency, jitter and
loss equals the value recorded above.

| Run | Link-Live result ID | `testStartTime` (UTC) | File | Grade | Avg latency / jitter / loss | Max latency |
| --- | --- | --- | --- | --- | --- | --- |
| Pair 1 (excluded), 10 Mbps | `6ace5964429e02da5bf89013` | 16:15:33.807 | `20261013-161635.lrpt` | green | 401 µs / 231 µs / 0 | 9.4 ms |
| Pair 2, 10 Mbps | `6ace5b10429e02da5bfaaca2` | 16:22:42.590 | `20261013-162343.lrpt` | green | 432 µs / 226 µs / 0 | 10.5 ms |
| Solo, 100 Mbps | `6ace5c80429e02da5bfc6ff5` | 16:28:49.929 | `20261013-162950.lrpt` | red | 85.2 ms / 19 µs / 12.2 % | 160.4 ms |
| Pair 4, 25 Mbps | `6ace5d9c429e02da5bfdb8e1` | 16:33:34.715 | `20261013-163435.lrpt` | green | 504 µs / 115 µs / 0 | 18.6 ms |

Pair 3 (16:26:17, stopped by hand) has no Link-Live result: no performance
result exists between 16:24 and 16:28.
The EtherScope's Path run to CT307 is Link-Live path analysis
`6ace5fe8429e02da5b008b43` (16:44:22 UTC, destination 10.44.40.23).

The EtherScope's own maximum latency was 10.5 ms in pair 2 and 18.6 ms in
pair 4. So the path itself has a tail of the order seen in stem's FD maximum
(21.3 and 26.8 ms), although its mean stays at 0.43–0.50 ms. The FD-mean and
FDV tolerances above compare means and variation, not maxima, so they are
unchanged.
