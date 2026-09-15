#!/usr/bin/env bash
# check-service-restart-policy.sh — a misconfigured install must stop, not spin.
#
# The shipped unit had Restart=on-failure with RestartSec=5 and no start limit.
# systemd's default limit is 5 starts in 10 s, which a 5 s cadence can never
# fill, so a daemon that exits 1 on missing credentials respun forever — 34,763
# times on CT307 — while the installer printed "installed successfully"
# (stem#1249). Two things stop that, and both are one line someone can move or
# delete without noticing:
#
#   - StartLimitIntervalSec / StartLimitBurst must be in [Unit]. In [Service]
#     systemd logs "Unknown key name" and IGNORES them, so the unit looks fixed
#     and behaves exactly as it did before. That is the regression this guards.
#   - postinstall must look at whether the service is actually running before it
#     claims anything, so the install output cannot lie about a dead unit.

set -euo pipefail

cd "$(dirname "$0")/.."

UNIT=deploy/systemd/stem.service
POSTINSTALL=deploy/nfpm/postinstall.sh

status=0
fail() {
	printf 'FAIL: %s\n' "$1" >&2
	status=1
}

for f in "${UNIT}" "${POSTINSTALL}"; do
	[ -f "${f}" ] || fail "${f} is missing"
done
[ "${status}" -eq 0 ] || exit 1

# The [Unit] section is everything from the [Unit] header to the next header.
unit_section="$(awk '/^\[Unit\]/{inside=1; next} /^\[/{inside=0} inside' "${UNIT}")"

for key in StartLimitIntervalSec StartLimitBurst; do
	if ! grep -q "^${key}=" "${UNIT}"; then
		fail "${UNIT} sets no ${key}; an unconfigured install will respin forever"
	elif ! grep -q "^${key}=" <<<"${unit_section}"; then
		fail "${UNIT} sets ${key} outside [Unit]; systemd ignores it there"
	fi
done

burst="$(awk -F= '/^StartLimitBurst=/{print $2}' <<<"${unit_section}" | tr -d '[:space:]')"
if [ -n "${burst}" ] && [ "${burst}" -gt 5 ]; then
	fail "${UNIT} StartLimitBurst=${burst} is looser than systemd's own default of 5"
fi

if grep -q '^Restart=' "${UNIT}" && ! grep -q '^Restart=no' "${UNIT}"; then
	if ! grep -q '^RestartSec=' "${UNIT}"; then
		fail "${UNIT} restarts without a RestartSec; the limit window is unpredictable"
	fi
fi

# An is-active grep alone proves nothing: the pre-fix script already called it,
# to choose between `restart` and `start`. What matters is an is-active AFTER
# the last start attempt -- that is the one whose answer can gate the banner.
last_start="$(grep -n '^[[:space:]]*systemctl \(re\)\?start stem.service' "${POSTINSTALL}" |
	tail -1 | cut -d: -f1)"
if [ -z "${last_start}" ]; then
	fail "${POSTINSTALL} never starts stem.service"
elif ! awk -v n="${last_start}" 'NR > n && /is-active/ {found=1} END {exit !found}' "${POSTINSTALL}"; then
	fail "${POSTINSTALL} checks is-active only before starting; it cannot know what it is claiming"
fi

# The banner said "installed successfully" whatever the unit was doing. Judge
# the output, not the prose: comment lines may name the old string to explain it.
if grep -v '^[[:space:]]*#' "${POSTINSTALL}" | grep -qi 'installed successfully'; then
	fail "${POSTINSTALL} claims success unconditionally; say what is actually true"
fi

if ! grep -q '/etc/stem/environment' "${POSTINSTALL}"; then
	fail "${POSTINSTALL} does not name /etc/stem/environment on the failure path"
fi

if [ "${status}" -eq 0 ]; then
	printf '✅ service restart policy: start limit in [Unit], postinstall verifies\n'
fi

exit "${status}"
