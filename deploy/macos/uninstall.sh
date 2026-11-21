#!/bin/bash
#
# Remove Stem from macOS:  sudo ./uninstall.sh [--purge]
#
# Without --purge the launchd services and the binary are removed, and
# stem.yaml, credentials.json, certs/, data/ and logs/ stay in /usr/local/stem
# so a later install picks them up. --purge deletes /usr/local/stem entirely.
#
set -euo pipefail

INSTALL_DIR="/usr/local/stem"
PURGE=false

die() { echo "error: $*" >&2; exit 1; }

case "${1:-}" in
    "") ;;
    --purge) PURGE=true ;;
    --help|-h) echo "usage: sudo $0 [--purge]"; exit 0 ;;
    *) die "unknown option: $1" ;;
esac
[[ "$(uname)" == "Darwin" ]] || die "this script is for macOS"
[[ $EUID -eq 0 ]] || die "run as root: sudo $0"

launchctl bootout system/com.stem 2>/dev/null || true
launchctl bootout system/com.stem.update 2>/dev/null || true
pkill -f "$INSTALL_DIR/stem web" 2>/dev/null || true
rm -f /Library/LaunchDaemons/com.stem.plist /Library/LaunchDaemons/com.stem.update.plist

if [[ "$PURGE" == true ]]; then
    rm -rf "$INSTALL_DIR"
    echo "Removed $INSTALL_DIR, including configuration, credentials and data."
else
    rm -f "$INSTALL_DIR/stem" "$INSTALL_DIR/stem.bak"
    echo "Removed the service and binary. Configuration, credentials and data are"
    echo "kept in $INSTALL_DIR; run 'sudo $0 --purge' to delete them."
fi
