# ADR-0012: In-place update — verification, install and rollback

**Status:** Proposed
**Date:** 2026-10-16

## Context

The owner decided on 2026-09-12, and amended on 2026-09-17, that stem updates
itself in place "just like iOS, macOS, Android". The update is opt-in and off by
default, the installer offers it, a Settings toggle can change it later, and an
air-gapped install never attempts a callout. It is in v1.0 (STM-21, #1796). The
work has two slices:

1. an operator-invoked `stem update` verb that verifies, stages, restarts and
   rolls back;
2. a scheduled check, default off, built on top of slice 1.

Three constraints shape it:

- **Fail closed.** seed's old updater checked a downloaded binary against a
  checksum served by the same host. That proves nothing, and it was deleted
  (seed#2354). Trust has to come from the release's signing identity, never from
  where the bytes came from.
- **The upgrade path has to be trustworthy first.** niac-go#2092 was a stale
  state file that survived an upgrade and broke every later start while
  deploy-validate passed. An updater would have applied that silently. STM-23
  proved stem's upgrade path, and the updater must not add a new way to strand
  state.
- **Stem is a measurement instrument.** An update between two runs makes them
  non-comparable. A result must record the version that produced it, and no
  update may start while a test is running.

Every release already publishes what verification needs (checked on v0.27.12):

- a sigstore bundle (`application/vnd.dev.sigstore.bundle.v0.3+json`) for
  `checksums.txt` and for each artifact, signed keylessly with the certificate
  identity
  `https://github.com/MustardSeedNetworks/stem/.github/workflows/release.yml@refs/tags/vX.Y.Z`,
  issuer `https://token.actions.githubusercontent.com`;
- `stem-slsa-provenance.intoto.jsonl`, one DSSE bundle of SLSA v1 provenance.
  Its nine subjects are the archives and packages. Its workflow parameters name
  the repository, `.github/workflows/release.yml` and `refs/tags/vX.Y.Z`.

## Decision

### Verification: sigstore-go, offline, with an embedded trust root

Bundles are verified in-process with `github.com/sigstore/sigstore-go`. The
Sigstore public-good trusted root (`trusted_root.json`: the Fulcio CAs and the
Rekor keys) is embedded in the binary. Verification makes **no network call**:
the bundle carries the certificate, the Rekor inclusion proof and the
integrated timestamp. A probe built against v1.3.0 verified the real v0.27.12
`checksums.txt` bundle and refused the same file with one byte appended.
`strace -e trace=connect,socket` saw zero socket or connect calls.

An artifact is accepted only when **all** of these checks pass. A missing file
counts as a failure. There is no flag that skips a check.

1. The `checksums.txt` bundle verifies with the exact certificate identity for
   the **target** tag (SAN equality, not a prefix or regex), the GitHub Actions
   issuer, one transparency-log entry and one integrated timestamp.
2. The artifact's own `.cosign.bundle` verifies with the same policy.
3. The artifact's SHA-256 equals its line in `checksums.txt`.
4. The provenance bundle verifies with the same identity, its predicate type is
   `https://slsa.dev/provenance/v1`, one subject has the artifact's exact name
   and SHA-256, and its workflow parameters name this repository,
   `.github/workflows/release.yml` and the target tag.

Pinning the SAN to the target tag means a source cannot pass off one release's
artifact as another's. Pinning the release workflow means an artifact signed by
any other workflow in the repository is refused.

The embedded root changes only when Sigstore rotates its keys. A scheduled
fleet CI check fetches the TUF-published root and fails when it differs from the
embedded copy, so a rotation becomes a PR. A stem built before a rotation
cannot verify releases signed after it, and the error says so and points to a
manual reinstall. An operator-supplied root is not accepted: it would be a
fail-open switch.

**Downgrade is refused.** Any older signed release verifies, so a source could
otherwise freeze or roll a site back. `stem update` installs only a version
newer than the running one. Rollback (below) is internal and never goes through
the resolver.

### Sources

The `--source` flag sets where releases come from. It is persisted for the
scheduled layer in slice 2.

- `github` (the default) uses the GitHub releases API for this repository.
- An `https://` base URL is a mirror. An absolute directory path is an
  air-gapped drop. Both use one layout: `<base>/vX.Y.Z/<asset>`, with the assets
  kept under their release names. For a directory, "latest" is the highest
  semver subdirectory. HTTPS has no portable directory listing, so a mirror
  names its newest tag in a `<base>/latest` file instead (amended 2026-10-17,
  as built in `foundation/pkg/release` v0.10.0).

Verification does not depend on the source. A mirror can withhold a release,
but it cannot alter one.

### Who installs: a privileged verb, never the daemon

The daemon runs as `stem` under `NoNewPrivileges`, `ProtectSystem=strict` and a
capability set. It cannot replace `/usr/bin/stem`, and it should not be able
to. `stem update` runs as root, or Administrator on Windows, and refuses
otherwise. Slice 2's scheduled layer is a root-owned timer that runs the same
verb: a systemd timer, a launchd `StartCalendarInterval` daemon, or a Windows
scheduled task. It exits before any network call unless the persisted opt-in is
on. The Settings toggle changes that persisted setting, and the daemon stays
unprivileged.

The scheduled layer, amended 2026-10-17 as built:

- The opt-in is `<state dir>/update.json`, `{"automatic": false}` when the
  file is missing. It is given the data directory's owner, so the Settings
  toggle writes it without root, and `stem update --automatic on|off
  [--source …]` writes it from the CLI. The timer reads it inside the data
  directory only (a symlink out of it is refused), as strict JSON of at most
  4 KiB, because it runs as root over a file the service account can write.
  That account can therefore turn updates on or name a source. It gains
  nothing by it: only a newer release that passes every check is installed.
- `stem update --scheduled` loads the opt-in before anything else and exits
  0 while it is off. It takes no `--source`; the saved one applies. A held run
  lock is a skip, not a failure: the next day's check tries again.
- The timers fire daily around 03:00 local time: `stem-update.timer`
  (`RandomizedDelaySec=1h`, `Persistent=true`), enabled by the `.deb` and
  `.rpm`; the `com.stem.update` LaunchDaemon, loaded by the macOS `.pkg`; and
  on Windows, where the zip has no installer, a scheduled task the operator
  registers once (`docs/DEPLOYMENT.md`).
- The `.deb` offers the opt-in at install with a debconf question,
  `stem/automatic-updates`, at priority high and default false. Debconf falls
  back to its noninteractive frontend without a terminal, so an unattended
  install takes the default and never waits for an answer; it can preseed
  `true` instead. A yes writes the opt-in only when `update.json` does not
  exist yet, so a choice saved from Settings or the CLI outlives every upgrade.
  The `.rpm` cannot prompt (rpm scriptlets have no terminal) and prints the
  command that turns it on. The macOS `.pkg` offers it as an "Automatic
  updates" choice that starts unselected: a payload-less component whose
  postinstall turns the opt-in on, again only when `update.json` does not
  exist. `installer -pkg` from a script takes the default, and
  `-applyChoiceChangesXML` selects it, the `.pkg`'s preseed.

The install method follows how the running binary was installed. The
operator never chooses it.

| Install | Detected by | Apply | Restart |
| ------- | ----------- | ----- | ------- |
| `.deb` | `dpkg -S` owns the binary | `dpkg -i` | `systemctl restart stem` |
| `.rpm` | `rpm -qf` owns the binary | `rpm -U` | `systemctl restart stem` |
| macOS tarball | `/usr/local/stem` prefix | replace the binary from the tarball | `launchctl kickstart -k` |
| Windows zip | anything else on windows | rename the running `.exe` aside, place the new one | none: Windows ships no service |

An install that matches none of these, such as a binary run from a build tree,
is refused with a message that names the reason.

Two refinements, amended 2026-10-17 as built in `internal/update`:

- On macOS `/usr/local/stem` is also the daemon's data directory, so only the
  binary is replaced (written beside it and renamed into place). The state
  copy covers the rest of the tree.
- On Linux the unit allows three starts a minute, so a refused start cannot
  respin forever. An update that rolls back starts the service four times:
  the package's postinst and the updater's own restart, twice. The updater
  therefore runs `systemctl reset-failed stem` before each deliberate restart.
  A crash loop is still caught by the restart-count check.

### Sequence and rollback

1. Take the run lock (below), or refuse if a test holds it.
2. Resolve the target, then download and verify it.
3. Download and verify the **running** version's artifact too. With no
   verified rollback artifact, nothing is touched. Rollback must never depend
   on the network after the point of no return.
4. Stop the service and copy the state directory (`/var/lib/stem`, or the
   platform equivalent) to a staging directory beside it. A database migration
   made by the new version must be undoable along with the binary. This is the
   niac-go#2092 lesson.
5. Apply and restart.
6. Health check, with a 60 s deadline: the service is active, its restart count
   is unchanged, and `https://127.0.0.1:<port>/__version` reports the target
   version with a non-empty `uiBuildHash`. On Windows, `stem version` from the
   new binary reports the target.
7. On failure: stop the service, restore the state copy, reapply the previous
   artifact (`dpkg -i`, `rpm -U --oldpackage`, or the tree or binary swap),
   restart, and confirm the previous version is healthy. Exit non-zero, naming
   the failed check.
8. On success, remove the staging copy and the rollback artifact.

Every step is logged as a structured event (`event=update.*`) that names the
from- and to-versions.

### Never mid-run, and every result names its version

While any test, session or reflector run is active, the daemon holds an
advisory lock on `<state dir>/run.lock`: `flock` on Unix, `LockFileEx` on
Windows. `stem update` takes the lock non-blocking and refuses while the daemon
holds it. It keeps holding the lock through the install, so no run can start
against a binary that is being replaced. A daemon that the update restarts
therefore finds the lock held at boot: a reflector set to start at boot waits
for the lock in the background while the API serves, because the health check
reads the API (amended 2026-10-17). Slice 1 also records stem's version
and commit with every stored result. That is what makes results from either
side of an update distinguishable.

### Where the code lives

Verification, source resolution and the trust root go in a new
`foundation/pkg/release`. That is the same split as `pkg/license`: the security
core lives once in foundation, and each product supplies its policy. Stem's
policy is its repository and workflow identity, its asset names and its install
table. seed (S-U1) and niac (U1) have the same row for v1.1 and should not
re-implement bundle verification. Stem's `internal/update` holds the install,
rollback and lock logic, and `cmd/stem` holds the verb.

## Alternatives considered

- **Shell out to `cosign`.** Rejected. Customer hosts do not have it, and it
  would be one more binary for the update path to trust.
- **Verify the bundle by hand** (the Fulcio chain, the Rekor SET, DSSE).
  Rejected. That is security-critical parsing we would own forever, and
  sigstore-go is the reference implementation.
- **Checksums only.** Rejected for the reason seed's updater was deleted.
- **The daemon updates itself.** Rejected. The daemon would need root, which
  undoes the hardened unit.
- **Rollback by re-downloading the old version after a failure.** Rejected. A
  failed upgrade on a site whose source just went away would be stranded.

## Consequences

- **Dependency cost.** In the probe module, sigstore-go v1.3.0 pulls about 290
  extra modules into the graph (368 against stem's 79 today), and the probe
  binary alone is 27 MB. govulncheck reported no reachable vulnerability in the
  probe. This is the largest single dependency addition stem would make, and it
  is the main thing for the owner to accept or reject.
- **Hard Rule 3 still holds.** An install makes no update callout until an
  operator runs `stem update` or turns the schedule on. Slice 2 proves that with
  a network-blocked install that makes zero outbound attempts.
- **Doc change before release.** `NIAC_V1_SUPPORT_AND_COMMERCIAL.md` §3.2 has to
  be reworded before a release ships the verb (support row D-DOC-3).
- **Sigstore rotation.** A stem built before a Sigstore key rotation needs one
  manual reinstall after it.
- **Unknown install methods.** An install that matches no row in the table is
  refused, never guessed at.
