# The test master runs where AF_XDP does not — #1232 on the wire, 2026-09-15

**Question.** `STM-13_WIRE_2026-09-14.md` proved #1217 on the wire, but only
with a scratch line forcing AF_PACKET: `select_platform` picks AF_XDP on every
Linux build and never falls back, so an _unmodified_ binary could not run
RFC 2544 in a container at all. That was filed as
[#1232](https://github.com/MustardSeedNetworks/stem/issues/1232). This is the
fix and its proof.

**Answer.** The test master now degrades to AF_PACKET at runtime, as the
reflector already did. An **unmodified** binary completes an RFC 2544
throughput search between two containers and the reflector returns its frames.
No scratch change of any kind.

| Claim | Result | Evidence |
| --- | --- | --- |
| Before the fix, an unforced run reaches no platform | **REPRODUCED** | `red_probe` built from `origin/main`'s dataplane: `Failed to create UMEM` → `Trial failed: -5`, `throughput_test ret=-5 platform=NONE`, exit 1 |
| After the fix, the same call reaches a platform | **PASS** | `test_platform_fallback`: `platform after an unforced run: fallback`, exit 0 |
| An unmodified binary runs RFC 2544 where AF_XDP fails | **PROVEN** | `Iterations: 10` from `stem test -t rfc2544_throughput`; daemon log: `Platform: AF_XDP` → `Failed to initialize platform` → `continuing with AF_PACKET` |
| The frames are real | **PROVEN** | 6,686 frames on the reflector's wire, 13,371 on the sender's, `0 packets dropped by kernel` on both |
| The reflector returns them | **PROVEN** | `192.168.64.78.3842 > 192.168.64.79.12345: UDP, length 18` in the sender's capture |
| The reported throughput is a measurement | **NO — still [#1233](https://github.com/MustardSeedNetworks/stem/issues/1233)** | 9990.23 Mbps / 14,866,419 pps / 99.90 % over a link that carried ~6.7 k frames in ~70 s. Untouched by this change. |

## The fix

`select_platform` chose a platform and `prepare_platform` failed the run when
its `init()` failed. The choice is now two named steps over one worker-init
helper:

- `rfc2544_preferred_platform()` — AF_XDP where the build saw
  `<linux/if_xdp.h>`, AF_PACKET on any other Linux build.
- `rfc2544_fallback_platform(tried)` — what to try after `tried` failed on this
  host. AF_XDP degrades to AF_PACKET; AF_PACKET is the last resort and returns
  `NULL`.

`prepare_platform` initializes the preferred platform, and on failure retries
once with the fallback. That is the reflector's own behaviour
(`src/reflector/core.c`), not a second policy.

`config.force_packet` is **not** carried across the Go boundary and should not
be. The issue's "suggested shape" asked for both halves, but once the runtime
fallback exists, a knob to pin AF_PACKET has no named operator who needs it on
a host where AF_XDP works — it would be a config knob standing in for a
decision, plus a permanent test surface. Recorded on #1232.

## Two things worth not rediscovering

**`struct platform_ops` has four independent definitions** — `core.c:61`,
`y1564.c:42`, `platform_stub.h:21`, and an anonymous copy inside each of
`xdp_platform.c` and `packet_platform.c`. The two `get_dataplane_*_ops()`
functions return `const void *`; `core.c` declared them as returning
`const platform_ops_t *` in a local `extern`, a type mismatch that compiles
only because no translation unit sees both. Declaring those getters in a shared
header therefore fails the build outright. That is why the new seam exposes
`rfc2544_preferred_platform()` rather than the getters. The four-way duplication
is a real defect and is **not** fixed here.

**`make c-test` on macOS does not build this test.** The Linux branch builds it;
the Darwin branch builds three tests against `-DSTUB_PLATFORM`. A clean local
`make c-test` on the Mac is not a clear for the dataplane.

## Bench

Two Apple containers on one L2 segment (`192.168.64.78` reflector,
`192.168.64.79` test master), image `golang:1.27.0-trixie` plus
`libxdp-dev libbpf-dev tcpdump iproute2 curl procps` — gcc-14, because gcc-12
cannot compile this tree (`-std=c23`). The binary is a `CGO_ENABLED=1`
linux/arm64 `make go` build of this branch with the UI built on the host first,
`uiBuildHash 3a3100d2f1502cf1a0d2bedc42e3f853`. A fresh container is a fresh
device fingerprint, so `stem license --trial` grants Professional; it must be
activated **before** `stem web` starts, because the daemon resolves entitlement
at start. Throwaway credentials; both containers destroyed afterwards.

```text
$ stem test -i eth0 -t rfc2544_throughput -d 5 --frame-sizes 64 --trials 1 \
    --peer 192.168.64.78 --json
"data": { "FrameSize": 64, "Iterations": 10,
          "MaxRateMbps": 9990.234375, "MaxRatePPS": 14866419 }

daemon log:
[210.651] [INFO]  Platform: AF_XDP (high performance)
[210.651] [ERROR] Failed to initialize platform
[210.651] [WARN]  Platform initialization failed; continuing with AF_PACKET (reduced performance)

reflector 192.168.64.78 : 6686 captured, 0 dropped by kernel
sender    192.168.64.79 : 13371 captured, 0 dropped by kernel
05:31:34.503790 IP 192.168.64.79.12345 > 192.168.64.78.3842: UDP, length 18
05:31:34.503908 IP 192.168.64.78.3842 > 192.168.64.79.12345: UDP, length 18
```

`--peer` is required and is documented nowhere ([#1235](https://github.com/MustardSeedNetworks/stem/issues/1235));
`stem --help` still lists neither `--peer` nor `--peer-port` under TEST OPTIONS.

## What this does not close

ST-2 clause 3 (a Playwright suite run against a licensed real daemon) is still
open. #1232 was its first blocker; two remain:
[#1231](https://github.com/MustardSeedNetworks/stem/issues/1231) (the daemon
setuids itself on the first reflector run, so a harness that starts and stops a
reflector twice gets `EPERM` behind a bare 500) and
[#1233](https://github.com/MustardSeedNetworks/stem/issues/1233) (the offered
rate is reported as measured, so a step-transition assertion would go green over
a fiction — the clause needs a credible per-step measurement, the amendment
that #1078 asked about).
