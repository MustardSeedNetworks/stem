# Can two Apple containers stand in for the lab? — 2026-09-14

**Question.** Lab week (standing rule 13) puts CT307, CT304, the EtherScope and
both dev servers out of reach until 2026-09-21. ST-2 clause 3 (a Playwright
suite run against a licensed real daemon) and STM-13's Ubuntu/Fedora legs both
need a Linux host with a working dataplane. The lab-free proposal was two Apple
containers on this Mac. The instruction was to prove it, not assume it.

**Answer.** Four of the five prerequisites hold. The fifth does not, and the
reason is a product defect, not a limitation of the substitute: the RFC 2544
test master transmits nothing and reports success anyway
(**[#1217](https://github.com/MustardSeedNetworks/stem/issues/1217)**). Until
that is fixed, a container run of ST-2 clause 3 would be green over a fiction,
so the clause stays open.

| Prerequisite | Result | Evidence |
| --- | --- | --- |
| Two containers share one L2 segment | **YES** | `192.168.64.33` and `.34` on one `/24`; `ping` at `ttl=64`; `ip neigh` holds the peer's own MAC `fe:2b:15:70:f6:a9` as `REACHABLE`, not the gateway's. |
| A container has `CAP_NET_RAW` / `CAP_NET_ADMIN` | **YES** | `CapEff: 00000000a80425fb` — bits 12 and 13 set. No privileged flag needed. |
| AF_PACKET capture works inside a container | **YES** | `tcpdump -i eth0` captured 5 of 5 ICMP frames, `0 packets dropped by kernel`. |
| A `CGO_ENABLED=1` linux/arm64 build exists and its dataplane loads | **YES** | See below. `/api/v1/capabilities` → `{"reflector":{"supported":true},"testMaster":{"supported":true}}`; `stem reflect -i eth0 --profile all` starts and holds a run id, where the `msn/stem` `CGO_ENABLED=0` image answers `500 … CGO dataplane not available on this platform`. |
| An RFC 2544 run moves real frames between them | **NO** | #1217: ~1 ms run, zero frames on either wire, `Iterations: 0`, `"success": true`. |

## Building the arm64 CGO binary

Never produced before: `.goreleaser.yml` builds `CGO_ENABLED=1` for linux
**amd64** only, and every other target — linux arm64 included — at
`CGO_ENABLED=0` with no dataplane at all. Two things had to be right.

**The compiler, not the libraries.** `golang:1.27.0-bookworm` carries gcc-12,
which cannot build this tree at all:

```text
gcc: error: unrecognized command-line option '-std=c23'; did you mean '-std=c2x'?
make: *** [mk/build.mk:77: dataplane] Error 1
```

`Makefile:150` sets `-std=c23`, which needs gcc-14. On
`golang:1.27.0-trixie` (gcc-14.2.0) `make dataplane` produces
`build/libreflector.a` and the link succeeds:

```text
libxdp.so.1 => /lib/aarch64-linux-gnu/libxdp.so.1
libbpf.so.1 => /lib/aarch64-linux-gnu/libbpf.so.1
```

Build deps beyond the image: `build-essential libxdp-dev libbpf-dev pkg-config`.
The UI was built on the host (Node 26.8.1, the pinned version) because
`//go:embed ui/*` at `internal/api/server.go:117` will not compile against an
`internal/api/ui/` holding only `.gitkeep`; `make go` then stamped the canonical
ldflags, and the running daemon reports a real hash rather than an empty one:

```json
{"version":"v0.24.109-2-g8d26fc7","commit":"8d26fc7",
 "buildTime":"2026-09-14T20:28:13Z",
 "uiBuildHash":"3a3100d2f1502cf1a0d2bedc42e3f853"}
```

This binary is a local proof artifact. Shipping a CGO linux/arm64 release
artifact is a product decision and is **not** proposed here — the README was
just corrected to say arm64 has no dataplane, and that remains true of what we
ship.

## Entitlement is no longer the wall

The 2026-09-14 record had ST-2 clause 3 blocked first on an expired trial, then
on the image. Neither holds now. A fresh container is a fresh device
fingerprint, so the trial starts:

```text
Success: Trial started! 14 days of full access.
Status:    Trial Mode      Days Left: 13
Tier:      Professional (full access during trial)
Device ID: 828F53EFFD5D542E      Platform:  linux
```

One wrinkle for whoever automates this: **the daemon resolves entitlement at
start**, so a trial activated after `stem web` is running is not seen — the
first RFC 2544 attempt was refused `requires a Professional licence` until the
daemon was restarted. Activate the trial before starting the daemon.

Each container also needs `STEM_AUTH_USERNAME` / `STEM_AUTH_PASSWORD` to start
`stem web` at all. Throwaway credentials were used here and the containers were
destroyed afterwards.

## What this leaves

- **ST-2 clause 3** — still open, now on #1217 rather than on entitlement or the
  image. The container pair is otherwise a viable host: two daemons, one
  licensed, on one L2 segment, both with a working dataplane.
- **STM-13's Ubuntu/Fedora legs** — the container substitute is sound for
  install, first-run and `/__version`, but the RFC 2544 and Y.1564 legs cannot
  be run honestly anywhere until #1217 is understood, including on the dev
  servers when the lab returns.
- **#1217 itself** — its second half (zero results reported as success) is
  platform-independent and is the half that matters most: it is the shape
  #1127 was meant to have closed.
