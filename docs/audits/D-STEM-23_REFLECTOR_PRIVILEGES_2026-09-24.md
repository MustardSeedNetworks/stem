# The reflector no longer drops the daemon's privileges — #1231, 2026-09-24

**Question.** [#1231](https://github.com/MustardSeedNetworks/stem/issues/1231):
the first reflector run in a daemon's lifetime worked. Every later reflector
start, and every RFC 2544 run, then failed with a bare `500` until the service
was restarted.

**Answer.** `reflector_start` called `drop_privileges()`, which `setuid()`s to
`nobody`. Since STM-20 the reflector runs inside the long-lived `stem web`
process, and a `setuid()` changes the credentials of every thread, so the whole
daemon became uid 65534 and lost `CAP_NET_RAW` for good. The call and the
function are deleted: which user the daemon runs as, and which capabilities it
holds, is the service manager's decision. The shipped unit already runs as
`User=stem` with `AmbientCapabilities=CAP_NET_RAW`, where `drop_privileges()`
was a no-op because it only acted for uid 0. The only daemons it ever changed
were ones started as root: containers and `sudo stem web`.

A raw-socket failure now names its cause. `reflector_start` returns a negative
errno for every failure it can report, `Dataplane.Start` wraps it as a
`syscall.Errno`, and `MapTestError` answers a permission denial with `503
SERVICE_UNAVAILABLE`, "Raw socket access denied: the daemon needs
CAP_NET_RAW", in place of `500 INTERNAL_ERROR`, "Failed to execute test".

| Claim | Result | Evidence |
| --- | --- | --- |
| One root daemon: reflector start, stop, start, then a test run | **PASS** (before: second start `500`) | both reflector runs report `Run … started on UDP 3842`; RFC 2544 at 1280 B `success: true`, `MaxRatePPS 301631.65` against a peer container |
| The daemon's threads keep their UID | **PASS** (before: `9 65534`) | `/proc/PID/task/*/status`: `9 0` before any run, after each reflector run and after the test run |
| A raw-socket failure returns its cause | **PASS** (before: `500 INTERNAL_ERROR`) | daemon under `setpriv --reuid=65534 --inh-caps=-all`, `CapEff: 0000000000000000`: `status=503 code=SERVICE_UNAVAILABLE`, internal error `operation not permitted` |
| Init failures report the syscall's errno | **PASS** (before: `-1`) | `test_packet_platform_init`: `init_failure_exits_report_the_syscall_errno`, `Expected: -19 Actual: -1` on main |
| Two starts keep the credentials, in Go | **PASS** (before: `start 2: … operation not permitted`) | `TestStartKeepsTheDaemonsCredentials`, run as root in the container |

## Bench

Two Apple containers from `golang:1.27.0-trixie` (gcc-14; gcc-12 cannot compile
this tree): `192.168.64.9` runs the daemon under test and `192.168.64.10` runs
a peer daemon with its reflector on `eth0`. Both binaries are `CGO_ENABLED=1`
linux/arm64 `make go` builds with the UI built on the host first: one from
`origin/main` (`66e97ee`), one from this branch. Each daemon had a fresh 14-day
trial and throwaway credentials. Both containers were destroyed afterwards.

```text
######## BEFORE: origin/main
  before any run: 9 threads Uid 0
== reflector start 1   Run stem-6c320bfb9ceab713-1 started on UDP 3842.
  after reflector run 1: 9 threads Uid 65534
== reflector start 2   Error: daemon returned 500
== RFC 2544 run        exit=1, "status": "failed"
[INFO] Dropped privileges to nobody (uid=65534, gid=65534)
[ERROR] Failed to create AF_PACKET socket: Operation not permitted
level=ERROR msg="API error" status=500 code=INTERNAL_ERROR
  message="Failed to execute test"
  internal_error="start reflector: failed to start reflector"

######## AFTER: this branch
  before any run: 9 threads Uid 0
== reflector start 1   Run stem-4cea217a5ce4f48b-1 started on UDP 3842.
  after reflector run 1: 9 threads Uid 0
== reflector start 2   Run stem-4cea217a5ce4f48b-2 started on UDP 3842.
  after reflector run 2: 9 threads Uid 0
== RFC 2544 run        exit=0, "success": true, "MaxRatePPS": 301631.65
  after the test run: 9 threads Uid 0
```

The daemon with no capabilities, before and after:

```text
BEFORE  status=500 code=INTERNAL_ERROR message="Failed to execute test"
        internal_error="start reflector: failed to start reflector"
AFTER   status=503 code=SERVICE_UNAVAILABLE
        message="Raw socket access denied: the daemon needs CAP_NET_RAW"
        internal_error="start reflector: failed to start reflector:
        operation not permitted"
```

## Why `-1` had to go as well

`reflector_start` returned `-1` for most failures and `-ENOMEM` for the rest.
`-1` is `-EPERM`, so any errno-reading caller would have reported "operation
not permitted" for a failed mutex or thread. Every exit now returns a real
negative errno: the socket, `PACKET_IGNORE_OUTGOING`, ring and bind failures
return their syscall's errno, captured before cleanup's `close()` overwrites it.
The UDP guard returns its errno, a missing guard port returns `-EINVAL`, and
`pthread_mutex_init` / `pthread_create` return their own error numbers. This is
also why the Go test run without capabilities passes on `origin/main`: there
the socket failure's `-1` reads as `EPERM` by coincidence. The C unit test is
what tells the two apart.

## What this does not close

- **The CLI still prints only `Error: daemon returned 503`.** It discards the
  daemon's error body, which is
  [#1235](https://github.com/MustardSeedNetworks/stem/issues/1235) (D-STEM-26).
- **The test master's own socket failures** go through `src/dataplane`, not the
  reflector, and are not mapped here. Once the daemon keeps its credentials
  they have no cause to fail.
- **`src/reflector/main.c`** was the standalone reflector that privilege
  dropping was written for. It is excluded from every build (`Makefile:158`).
