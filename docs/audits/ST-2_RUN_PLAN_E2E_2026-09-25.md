# A two-step run plan from the UI, on the wire — ST-2 clause 3, 2026-09-25

**Question.** ST-2's last acceptance clause (#1078): Playwright against a real,
licensed daemon selects RFC 2544 throughput and Y.1564, starts, and the run
plan reports both steps with the second running after the first passes. The
clause names `GET /api/v1/test/status`, which does not exist; the run plan is
published on `GET /api/v1/stats` (`steps[]`), so that is what is asserted.

**Answer.** Met. `ui/e2e/run-plan.spec.ts` drives the settings drawer, starts
from the Benchmark page, and asserts on the daemon's own answer: the plan is
`[rfc2544_throughput, y1564_config]`, one snapshot reads `[passed, running]`
with a non-zero measured rate on the first step, the page renders the same two
statuses, and the run finishes under the same `suiteId` with all four Y.1564
rates sent and reflected. Green on Chromium and WebKit.

| Claim | Result | Evidence |
| --- | --- | --- |
| The plan the daemon runs is the UI's selection, in order | **PASS** | `steps[].testType` = `[rfc2544_throughput, y1564_config]` |
| The second step runs after the first passes | **PASS** | one `/api/v1/stats` snapshot: `[passed, running]`, `testStatus running` |
| The passed step measured something | **PASS** | `MaxRatePPS` ≈ 107k in that snapshot |
| The page shows the daemon's plan | **PASS** | `run-plan-step-0` `data-status=passed`, `run-plan-step-1` `running` |
| Both steps carry real traffic | **PROVEN** | sender capture: 412,074 frames to the reflector, 412,072 back |
| The spec fails when nothing is reflected | **RED** | reflector stopped: fails on `MaxRatePPS` = 0 |

## Bench

dev-srv-ubuntu (Ubuntu 26.04, x86_64), as in
`ST-2_Y1564_DAEMON_2026-09-25.md`: two unprivileged Docker containers
(`ubuntu:26.04` plus `libxdp1 libbpf1 tcpdump`, `--cap-add NET_RAW
--cap-add NET_ADMIN`, so the AF_PACKET fallback runs) on one bridge, the repo
mounted at `/src`, both running the host's `CGO_ENABLED=1` `make build` through
`scripts/e2e-daemon.sh /run/stem 18644`.

```text
reflector    172.18.0.2  stem reflect -i eth0            (UDP 3842)
test master  172.18.0.3  stem license --trial, then the daemon
             -p 127.0.0.1:18644:18644 --mac-address 02:42:ac:12:00:33

STEM_E2E_RUN_PLAN_INTERFACE=eth0 STEM_E2E_RUN_PLAN_PEER=172.18.0.2 \
E2E_BASE_URL=https://localhost:18644 \
  npx playwright test --workers=1 e2e/run-plan.spec.ts

  ✓  1 [chromium] › e2e/run-plan.spec.ts › run plan › runs RFC 2544 throughput then Y.1564, in order, each measuring (23.7s)
  ✓  2 [webkit] › e2e/run-plan.spec.ts › run plan › runs RFC 2544 throughput then Y.1564, in order, each measuring (27.7s)
  2 passed (54.6s)
```

**Pin the test master's MAC.** The licence is bound to a device fingerprint
that includes the primary MAC, and Docker assigns a fresh MAC on
`docker restart`. Without `--mac-address` the restarted daemon logs
`license file unusable … cipher: message authentication failed` and every Pro
standard is refused.

The spec leaves frame sizes at their defaults and does not assert the Y.1564
verdict, for the reasons below.

## Found, not fixed here

- [#1463](https://github.com/MustardSeedNetworks/stem/issues/1463) — a Y.1564
  step is `passed` whatever the service measured. The executor sets `Success`
  without reading `ServicePass`; on this bench every step fails FDV (≈10 ms
  against the default 5 ms), and with the reflector stopped every step reads
  FLR 100 %, and both report `passed`.
- [#1464](https://github.com/MustardSeedNetworks/stem/issues/1464) — RFC 2544
  and Y.1564 measure only `FrameSizes[0]`; `Params["frame_sizes"]` is written
  and read by nothing, while the progress estimate counts every size.
- [#1465](https://github.com/MustardSeedNetworks/stem/issues/1465) — frame-size
  checkbox changes never reach the run config: `useConfigForm` forwards only
  `type === 'change'` watch events and `setValue` emits none.
- [#1466](https://github.com/MustardSeedNetworks/stem/issues/1466) — with the
  reflector stopped, the RFC 2544 search runs ten trials at 100 % loss and
  reports `passed` at 0 pps. The spec's measured-rate assertion is what fails
  on that bench.
