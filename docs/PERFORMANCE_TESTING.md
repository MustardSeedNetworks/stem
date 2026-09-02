# End-to-end performance testing with Stem

This guide covers the test workflows Stem actually ships. It deliberately does
not describe Netperf-style request/response testing, Flowgrind, HTTP or DNS
load generation, or hardware-timestamped latency — Stem does not do those, and
the issues proposing them were reviewed and closed; see issues 537, 540, 542,
543, 544 and 545 for the reasoning in each case.

Everything below was checked against `stem version 0.24.48`. Where a command
could not be executed on the machine this was written on, that is said plainly
rather than implied.

---

## 1. What Stem measures

Six modules, 29 test types. `stem list-tests` is the authoritative list; this
table is the map.

| Module | Standard | Test types |
| --- | --- | --- |
| Benchmark | RFC 2544 | `rfc2544_throughput`, `rfc2544_latency`, `rfc2544_frame_loss`, `rfc2544_back_to_back`, `rfc2544_system_recovery`, `rfc2544_reset` |
| ServiceTest | ITU-T Y.1564 / MEF | `y1564_config`, `y1564_perf`, `y1564`, `mef_config`, `mef_perf`, `mef` |
| TrafficGen | custom | `custom_stream` |
| Measure | ITU-T Y.1731 | `y1731_delay`, `y1731_loss`, `y1731_slm`, `y1731_loopback` |
| Certify | RFC 2889 / RFC 6349 / TSN | `rfc2889_forwarding`, `rfc2889_caching`, `rfc2889_learning`, `rfc2889_broadcast`, `rfc2889_congestion`, `rfc6349_throughput`, `rfc6349_path`, `tsn_timing`, `tsn_isolation`, `tsn_latency`, `tsn` |
| Reflector | loopback/echo | `reflect` |

**Test names are exact.** `-t throughput` is rejected; the type is
`rfc2544_throughput`. Aggregate names (`y1564`, `mef`, `tsn`) run their
module's sequence.

```bash
stem list-tests
```

---

## 2. Requirements

### The test master and the reflector are two machines

Every measurement is a round trip. One host generates and measures
(`stem test`), the other returns frames unchanged (`stem reflect`). Pointing a
test master at a host that is not reflecting produces 100% loss, not an error.

### Linux with CGO, on both ends

The dataplane is C, built with build tags, and Linux-only. On macOS or
Windows, or on a Linux build without CGO, every test command fails at
initialisation:

```text
Error: Failed to initialize dataplane: dataplane unavailable: dataplane operations require Linux with CGO enabled

Platform: darwin/arm64
Missing requirements:
  - Linux required (current: darwin)
```

This is a hard gate, not a degraded mode. The WebUI, TUI, licensing, and
`list-tests` all work everywhere; **running a test does not.**

### Licence tier

| Tier | Grants |
| --- | --- |
| Reflector | `reflect` only |
| Professional | `reflect` plus rfc2544, rfc2889, rfc6349, mef, tsn, y1564, y1731, api |

A first `stem test` with no licence starts a 14-day Professional trial
automatically and says so. Check and clear it with:

```bash
stem license --status
stem license --deactivate
```

---

## 3. A first run

On the reflector host:

```bash
stem reflect -i eth0 --profile all
```

On the test master:

```bash
stem test -i eth0 -t rfc2544_throughput -d 60
```

Defaults that matter, all visible in the run banner before traffic starts:

| Flag | Default | Meaning |
| --- | --- | --- |
| `--duration` / `-d` | 60 | seconds per trial |
| `--frame-sizes` | `64,128,256,512,1024,1280,1518` | RFC 2544's frame-size ladder |
| `--resolution` | 0.1 | binary-search resolution, % of line rate |
| `--max-loss` | 0.0 | loss ratio a rate must stay under to pass |
| `--warmup` | 2 | seconds discarded before measuring |
| `--trials` | 3 | repetitions per frame size |

Several tests in one run, comma-separated, in order:

```bash
stem test -i eth0 -t rfc2544_throughput,rfc2544_latency,rfc2544_frame_loss,rfc2544_back_to_back
```

Service activation, where the rates are the point:

```bash
stem test -i eth0 -t y1564 --cir 100 --eir 50
```

`--cir` and `--eir` are Mbps. The pass/fail thresholds are `--fd-threshold`
(frame delay, ms, default 10), `--fdv-threshold` (delay variation, ms,
default 5) and `--flr-threshold` (loss ratio, %, default 0.01).

---

## 4. Safety

Stem generates traffic at up to line rate. That is the point of it, and it is
also why it does not belong on a production path without arrangement.

- **Run on a link you own, or one whose owner has agreed.** An RFC 2544
  throughput search deliberately drives the link to the point of loss; on a
  shared path it is indistinguishable from an outage.
- **`--max-loss 0.0` is a pass criterion, not a limiter.** It decides which
  rates count as passing; it does not stop the generator from exceeding them
  during the search.
- **Bound the run before you bound the rate.** `-d` and `--trials` multiply:
  seven frame sizes × three trials × sixty seconds is 21 minutes of traffic
  per test type, before the binary search adds its own iterations.
- **Start small.** One frame size, short duration, then widen:
  `stem test -i eth0 -t rfc2544_throughput --frame-sizes 512 -d 10 --trials 1`

---

## 5. Reading the results

Human-readable by default; machine-readable on request:

```bash
stem test -i eth0 -t rfc2544_throughput --json
stem test -i eth0 -t rfc2544_throughput --csv
```

What each RFC 2544 sub-test answers:

| Test | The question | Read it as |
| --- | --- | --- |
| `rfc2544_throughput` | fastest rate with loss at or under `--max-loss` | the number people quote; per frame size, not one figure |
| `rfc2544_latency` | delay at a given load | latency at line rate and latency at idle are different claims |
| `rfc2544_frame_loss` | loss across a load sweep | the shape matters — a cliff is a buffer, a ramp is a policer |
| `rfc2544_back_to_back` | burst absorbed before loss | buffer depth |
| `rfc2544_system_recovery` | time to recover after overload | how the device behaves after it is pushed past its limit |
| `rfc2544_reset` | throughput recovery after a reset | availability, not speed |

Throughput below line rate at 64-byte frames and at line rate for 1518-byte
frames is the normal signature of a packet-rate limit rather than a bandwidth
limit. That is a result, not a fault.

---

## 6. Troubleshooting

**`Unknown test type 'throughput'`** — use the exact name from
`stem list-tests` (`rfc2544_throughput`).

**`dataplane unavailable: dataplane operations require Linux with CGO enabled`**
— the host cannot run tests. Section 2.

**100% loss on every frame size** — almost always the far end. Confirm
`stem reflect` is running there, on the interface facing the link under test,
and that the profile is not filtering your traffic out (`--profile all` is the
widest).

**Throughput far below expectation, no loss** — check the negotiated speed and
duplex of both interfaces before suspecting the path.

**The WebUI is unreachable at `http://host:8444`** — Stem is HTTPS-only. Use
`https://`. A browser with no scheme typed will get connection refused, which
is intended. The certificate is self-signed unless you have installed one;
`stem install-ca` adds Stem's root to the OS trust store.

---

## 7. Interfaces

Three, over the same engine:

```bash
stem test -i eth0 -t rfc2544_throughput   # CLI, scriptable, --json / --csv
stem tui                                  # terminal dashboard
stem web -p 8444                          # HTTPS WebUI (default :8444)
```

---

## Verification note

The commands in sections 1, 2 and 3 that do **not** require the dataplane —
`list-tests`, `license --status`, `license --deactivate`, argument parsing and
the run banner for every command shown — were executed against
`stem 0.24.48` and behave as documented, including the exact error text quoted
in section 2 and section 6.

The measurement runs themselves were **not** executed for this guide: the
dataplane requires Linux with CGO on both the test master and the reflector,
and this was written on macOS, where the gate above fires before any traffic
is generated. Numbers and result shapes in section 5 describe what each test
computes, not a captured run. Validating them end to end belongs on the Linux
dev servers with a real link, which is where platform-specific behaviour is
meant to be verified.
