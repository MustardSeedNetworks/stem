#!/bin/bash
#
# Upgrade an installed Stem on macOS from an unpacked release archive:
#   sudo ./upgrade.sh [path/to/stem]
#
# Replaces the binary and the launchd plists only. stem.yaml,
# credentials.json, certs/ and data/ are never touched. The previous binary
# is kept as stem.bak, and restored if the new one does not start.
#
set -euo pipefail

INSTALL_DIR="/usr/local/stem"
DAEMON_PLIST="/Library/LaunchDaemons/com.stem.plist"
UPDATE_PLIST="/Library/LaunchDaemons/com.stem.update.plist"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${1:-$SCRIPT_DIR/stem}"

die() { echo "error: $*" >&2; exit 1; }
running() { pgrep -f "$INSTALL_DIR/stem web" >/dev/null; }

[[ "${1:-}" == "--help" || "${1:-}" == "-h" ]] && { echo "usage: sudo $0 [path/to/stem]"; exit 0; }
[[ "$(uname)" == "Darwin" ]] || die "this script is for macOS; use the .deb or .rpm on Linux"
[[ $EUID -eq 0 ]] || die "run as root: sudo $0"
[[ -f "$BINARY" ]] || die "stem binary not found at $BINARY"
[[ -x "$INSTALL_DIR/stem" && -f "$DAEMON_PLIST" ]] || die "stem is not installed; run ./install.sh"

echo "Upgrading $("$INSTALL_DIR/stem" version 2>/dev/null || echo unknown) to $("$BINARY" version 2>/dev/null || echo unknown)"

launchctl bootout system/com.stem 2>/dev/null || true
launchctl bootout system/com.stem.update 2>/dev/null || true
for _ in 1 2 3 4 5; do running || break; sleep 1; done
running && pkill -f "$INSTALL_DIR/stem web" || true

cp -p "$INSTALL_DIR/stem" "$INSTALL_DIR/stem.bak"
install -m 755 -o root -g wheel "$BINARY" "$INSTALL_DIR/stem"
for name in com.stem com.stem.update; do
    [[ -f "$SCRIPT_DIR/launchd/$name.plist" ]] \
        && install -m 644 -o root -g wheel "$SCRIPT_DIR/launchd/$name.plist" "/Library/LaunchDaemons/$name.plist"
done

launchctl bootstrap system "$DAEMON_PLIST"
[[ -f "$UPDATE_PLIST" ]] && launchctl bootstrap system "$UPDATE_PLIST"

sleep 3
if ! running; then
    echo "error: the new binary did not start; restoring the previous one" >&2
    launchctl bootout system/com.stem 2>/dev/null || true
    install -m 755 -o root -g wheel "$INSTALL_DIR/stem.bak" "$INSTALL_DIR/stem"
    launchctl bootstrap system "$DAEMON_PLIST"
    exit 1
fi

echo "Now running $("$INSTALL_DIR/stem" version 2>/dev/null || echo unknown). Previous binary: $INSTALL_DIR/stem.bak"
