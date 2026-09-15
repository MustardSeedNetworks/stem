# A fresh `.deb` crash-loops while the installer reports success — #1249

**Question.** `deploy/nfpm/postinstall.sh` writes `/etc/stem/environment` with
every credential commented out, the daemon exits 1 without them, and
`deploy/systemd/stem.service` had `Restart=on-failure` with `RestartSec=5` and
no start limit. That is the mechanism behind the 34,763 restarts found on CT307
(memory `stem-ct307-deploy-found-three-defects`) and was filed as
[#1249](https://github.com/MustardSeedNetworks/stem/issues/1249). Does a fresh
unconfigured install actually respin forever, and does the fix stop it?

**Answer.** Yes, and yes. Reproduced on the released `v0.24.117` `.deb`:
**11 starts in 45 seconds**, `NRestarts=10`, the unit never leaving
`activating`, and `apt` exiting 0 under a banner reading "The Stem installed
successfully". With the fix the unit reaches `failed` in about twenty seconds
after exactly **three** starts, and the install output says the service is not
running and names the file to edit.

| Claim | Before | After |
| --- | --- | --- |
| Unit state ~50 s after install | `activating` (forever) | **`failed`** |
| `journalctl -u stem \| grep -c "Starting on"` | **11** and climbing | **3** |
| `systemctl show -p NRestarts` | **10** | **3** |
| `systemctl show -p StartLimitBurst` | `5` (systemd's default) | **`3`** |
| `systemctl show -p StartLimitIntervalUSec` | default 10 s | **`1min`** |
| Install output | "The Stem installed successfully" | **"The Stem is installed but NOT RUNNING"** + the three variables to set |
| `apt-get install` exit | 0 | 0 (deliberately — see below) |
| `ConfigurationDirectory` mode warning per start | 1 | **0** |

## The fix

**`StartLimitIntervalSec=60` / `StartLimitBurst=3`, in `[Unit]`.** systemd's
default limit is five starts in ten seconds, which a five-second restart
cadence can never fill — the loop never throttles. Three attempts in a minute
ends it. The section matters: these keys are `[Unit]` keys, and in `[Service]`
systemd logs "Unknown key name" and **ignores them**, so a unit could look
fixed and behave exactly as before. `systemctl show -p StartLimitBurst`
reading `3` rather than `5` is the proof they took effect.

**`postinstall.sh` reports what is true.** `systemctl start` returns as soon as
a `Type=simple` process forks, so an immediate `is-active` reads `active` for
the moment before an unconfigured daemon exits. The script now polls
`is-active` for five seconds after the start attempt and picks one of three
banners — running / not running / no systemd on this host — with the
credential instructions and `journalctl` pointer on the failure path only.

**Why it still exits 0.** A fresh install with no credentials _cannot_ start
the daemon, and that is the expected first-run state, not an installation
failure. Exiting non-zero would fail every first install and leave dpkg
half-configured. The defect was never the exit code — it was a banner that
claimed success over a dead unit. So the install succeeds and says so
accurately.

**Also fixed, one line in the same unit:** `ConfigurationDirectoryMode=0750`.
nfpm creates `/etc/stem` at `0750` (`.goreleaser.yml` `contents:`) while
`ConfigurationDirectory=stem` defaults to `0755`, so systemd warned on every
single start that the mode differed. The unit was the side that was wrong.

## The regression guard

The row's acceptance cannot run in CI — it needs systemd and a real package
install. `scripts/check-service-restart-policy.sh` guards the part that can
drift silently, and is wired into the Quality Checks job. Each of its checks
was proven to fail on the specific regression it exists for:

```text
RED 1: start-limit keys absent (pre-fix unit)
  FAIL: sets no StartLimitIntervalSec; an unconfigured install will respin forever
  FAIL: sets no StartLimitBurst; an unconfigured install will respin forever
RED 2: keys present but in [Service] (silently ignored by systemd)
  FAIL: sets StartLimitIntervalSec outside [Unit]; systemd ignores it there
  FAIL: sets StartLimitBurst outside [Unit]; systemd ignores it there
RED 3: burst looser than systemd's own default
  FAIL: StartLimitBurst=9 is looser than systemd's own default of 5
RED 4: pre-fix postinstall
  FAIL: checks is-active only before starting; it cannot know what it is claiming
  FAIL: claims success unconditionally; say what is actually true
GREEN: the branch as it stands
  ✅ service restart policy: start limit in [Unit], postinstall verifies
```

**RED 4 is the one worth keeping.** The first version of this gate grepped the
postinstall for `is-active` and went **green against the pre-fix script**,
because that script already called `is-active` — to choose between `restart`
and `start`. A gate that a known-broken input passes is worth nothing. The
check now requires an `is-active` occurrence _after_ the last start attempt,
which is the only one whose answer can gate a banner, and separately forbids
the literal unconditional success claim (judging the script's output, not its
comments — a comment may quote the old string to explain it).

## Bench, and four facts that cost the time

Ubuntu 26.04.1 LTS (`systemd 259`) as **PID 1 in an Apple container**. This
works and is the lab-free bench for every packaging row:

- **`--cap-add ALL` is required.** Without it `/sbin/init` exits immediately
  with no output; verified both ways.
- **`touch` is not a cgroup2 writability test.** Creating a plain file on
  cgroupfs always fails with `EPERM`; `mkdir` is the test. An earlier probe
  read "cgroup is read-only", which was wrong — `/sys/fs/cgroup` is `rw` and
  `mkdir` succeeds.
- **systemd remounts `/tmp` after boot**, so a `container cp` to `/tmp` lands
  under the pre-boot mount and vanishes. Use `/root`.
- **`container logs` is empty for a systemd PID 1** — read the journal over
  `container exec` instead.

The package under test was the released `stem_0.24.117_arm64.deb` with the two
changed files swapped in (`dpkg-deb -R`, replace, regenerate `md5sums`,
`dpkg-deb -b`) rather than a fresh goreleaser run: `DEBIAN/postinst` is
`deploy/nfpm/postinstall.sh` byte for byte, so the repack is faithful to what
goreleaser produces from this branch. macOS has no `dpkg-deb`, so the repack
ran in a `debian:trixie-slim` container. The RED and GREEN installs were in
**separate** containers, not the same one reset.

## Two things this run also settled, and one it did not

**The happy path is intact** — re-running the install with credentials set
reports "The Stem is installed and running", `is-active` is `active`,
`NRestarts=0`, and `/__version` answers `0.24.117` with a non-empty
`uiBuildHash` (`bc5b74e0…`). That is incidentally **STM-13's install /
first-run / `/__version` leg on Ubuntu 26.04**, taken lab-free.

**Out of scope, and it is the real fix.** The issue's own closing note is
right: the daemon should start in a setup state without credentials, matching
seed's and niac's first-run flow and the fleet's "first-run forces setup"
invariant. Then a fresh install would come up and wait to be configured
instead of failing at all. That is a product change, not a packaging one, and
it is not in this row.

`deploy/deb/*` and `deploy/rpm/stem.spec` carry their own copies of the unit
and were deliberately left alone: nothing ships them (`.goreleaser.yml`'s
`nfpms:` block is the only packaging path) and **D-STEM-9** deletes them.
