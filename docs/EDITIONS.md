# Editions

**Product:** Stem
**Status:** Current
**Owner:** Mustard Seed Networks
**Last updated:** 2026-09-08

Stem ships as **one binary on two tiers**. There is no separate build and no
edition string: what a deployment can do is decided by the key it holds, and by
nothing else. This file is the charter check the v1 plan calls for (STM-10) —
every catalog string below is a capability that exists and is enforced, and
every Free capability is ungated.

## 1. Tiers

| Tier | Price | Key required |
| --- | --- | --- |
| **Free (Reflector)** | — | Yes (a Reflector token, product code 1001) |
| **Pro** | $1,999/yr | Yes (a Professional token, product code 2001) |
| **Trial** | free, 14 days | No — `stem license --trial` grants the Pro catalog |

`Tier` is defined in [`internal/license/policy.go`](../internal/license/policy.go)
as `TierInvalid` / `TierReflector` / `TierProfessional`. The retired tier 3 (product code 3001) is rejected, with no
grandfathering. Commercial
arrangements — custom terms, volume licensing, net-30 — are not a third tier.

## 2. What each catalog string unlocks

`ProFeatures()` in `internal/license/policy.go` is the catalog. Stem prices
whole standards, and every test names its type in the body of one route, so the
gate is at the point the test type is resolved rather than at route
registration: `handleTestStart` maps the type through
[`services.FeatureForTestType`](../internal/services/features.go) and answers
402 `TIER_TOO_LOW` when the licence does not cover it.

| Feature | Tier | Capability | Test types it gates | Gate |
| --- | --- | --- | --- | --- |
| `reflector` | Free | Serving as a test endpoint | `reflect` | None, by design — see §3 |
| `rfc2544` | Pro | RFC 2544 benchmarking | the six `rfc2544_*` types | `handleTestStart` |
| `y1564` | Pro | Y.1564 service activation | `y1564`, `y1564_config`, `y1564_perf` | `handleTestStart` |
| `mef` | Pro | MEF service tests | `mef`, `mef_config`, `mef_perf` | `handleTestStart` |
| `y1731` | Pro | Y.1731 performance monitoring | the four `y1731_*` types | `handleTestStart` |
| `rfc2889` | Pro | RFC 2889 switch benchmarking | the five `rfc2889_*` types | `handleTestStart` |
| `rfc6349` | Pro | RFC 6349 TCP throughput | `rfc6349_throughput`, `rfc6349_path` | `handleTestStart` |
| `tsn` | Pro | TSN timing and isolation | `tsn`, `tsn_timing`, `tsn_isolation`, `tsn_latency` | `handleTestStart` |
| `trafficgen` | Pro | Custom traffic streams | `custom_stream` | `handleTestStart` |

The CLI carries a second, coarser check: `stem test` requires Professional (or
an active trial) for any test through `checkTestLicense` in
[`cmd/stem/cmd_testmaster.go`](../cmd/stem/cmd_testmaster.go). It is the same
verdict at a lower resolution, not a different policy.

`internal/services/features_test.go` fails the build if a Pro catalog string is
required by no test type, or if a registered test type requires no feature.
That is what stops the table above from drifting into a list of claims.

## 3. What is deliberately ungated

- **Reflection.** Free is the reflector, so `reflect` requires no entitlement.
  A gate here would make an unlicensed install useless as the thing the Free
  tier is sold as.
- **The web UI and the REST API it runs on.** Stem is operated through its API
  at every tier; gating it would gate the Free reflector's own console. Stem
  mints no programmatic tokens, so there is no separate API product to sell.
- **Stopping a run and reading a result.** Only starting a paid standard is
  sold.

## 4. Removed from the catalog (2026-09-08)

Two strings were in `proFeatures()` with nothing behind them:

- **`api`** — see §3. The API is not a separate product.
- **`multiuser`** — Stem has no second user. The `users` table was dropped
  (`migrationDropUsers`, "Drop the unused users table"); authentication is one
  operator account from `STEM_AUTH_USERNAME` / `STEM_AUTH_PASSWORD`.

Both were removed from the catalog and from the README's Pro column rather than
given gates, because gating a capability that does not exist is not
enforcement. `trafficgen` was added: `custom_stream` existed and ran unlicensed.

## 5. Validation is local

Keys are validated **offline**, against a device fingerprint, using the shared
crypto in [`foundation/pkg/license`](https://github.com/MustardSeedNetworks/foundation).
Stem does not contact a licence server, at startup or ever — a requirement, not
a preference: air-gapped industrial and government deployments are a target.

- No grace window, because there is nothing to be unreachable.
- Rebinding a key to new hardware is an operator action, not something the
  daemon negotiates.
- `/__version` reports build metadata only. It does not report the tier, and a
  client must not infer entitlement from anything it can read unauthenticated.

## 6. Open questions

| Question | Owner | Needed by |
| --- | --- | --- |
| Whether Free should work with no key at all (today an unlicensed install grants nothing and the CLI silently starts the trial) | Product | v1 |
| Whether the CLI check should be per-standard like the API's | Eng | v1 |
