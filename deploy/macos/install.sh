#!/bin/bash
#
# Install Stem on macOS as a launchd daemon. Run from the unpacked release
# archive:  sudo ./install.sh [path/to/stem]
#
set -euo pipefail

INSTALL_DIR="/usr/local/stem"
DAEMON_PLIST="/Library/LaunchDaemons/com.stem.plist"
UPDATE_PLIST="/Library/LaunchDaemons/com.stem.update.plist"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${1:-$SCRIPT_DIR/stem}"

die() { echo "error: $*" >&2; exit 1; }

[[ "${1:-}" == "--help" || "${1:-}" == "-h" ]] && { echo "usage: sudo $0 [path/to/stem]"; exit 0; }
[[ "$(uname)" == "Darwin" ]] || die "this script is for macOS; use the .deb or .rpm on Linux"
[[ $EUID -eq 0 ]] || die "run as root: sudo $0"
[[ -f "$BINARY" ]] || die "stem binary not found at $BINARY"
[[ -f "$SCRIPT_DIR/launchd/com.stem.plist" ]] || die "launchd/com.stem.plist is missing from the archive"
[[ ! -e "$INSTALL_DIR/stem" ]] || die "stem is already installed in $INSTALL_DIR; run ./upgrade.sh"

mkdir -p "$INSTALL_DIR/logs"
install -m 755 "$BINARY" "$INSTALL_DIR/stem"
# A config left behind by an earlier install or uninstall is the operator's.
[[ -e "$INSTALL_DIR/stem.yaml" || ! -f "$SCRIPT_DIR/stem.yaml" ]] \
    || install -m 644 "$SCRIPT_DIR/stem.yaml" "$INSTALL_DIR/stem.yaml"
chown -R root:wheel "$INSTALL_DIR"

install -m 644 -o root -g wheel "$SCRIPT_DIR/launchd/com.stem.plist" "$DAEMON_PLIST"
# The daily update check exits before any network call until an operator
# runs `stem update --automatic on`.
install -m 644 -o root -g wheel "$SCRIPT_DIR/launchd/com.stem.update.plist" "$UPDATE_PLIST"

launchctl bootstrap system "$DAEMON_PLIST"
launchctl bootstrap system "$UPDATE_PLIST"

sleep 3
pgrep -f "$INSTALL_DIR/stem web" >/dev/null \
    || die "service loaded but is not running; see $INSTALL_DIR/logs/stem.error.log"

echo "Stem $("$INSTALL_DIR/stem" version 2>/dev/null || echo unknown) is running."
echo "Open https://localhost:8444 and finish first-run setup with the token in"
echo "$INSTALL_DIR/.setup-token (sudo cat it)."
echo "Logs: $INSTALL_DIR/logs/"
