#!/bin/sh
set -e

# Stop the service only when the package is actually going away.
#
# RPM runs the OLD package's %preun *after* the NEW package's %post, so an
# unconditional stop here undid the upgrade: postinstall had just restarted
# stem, this stopped and disabled it, and the upgrade finished with the service
# down while the installer said "installed and running" (#1445). dpkg runs prerm
# before the new postinst, which is why only the .rpm broke.
#
# RPM passes the number of versions that will remain (0 on removal, 1 or more
# during an upgrade); dpkg passes a word. postremove.sh keys off the same
# convention. niac-go fixed the identical script in niac-go#2085.
case "${1:-}" in
    0 | remove | purge) ;;
    *) exit 0 ;;
esac

if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet stem.service 2>/dev/null; then
        systemctl stop stem.service || true
    fi
    if systemctl is-enabled --quiet stem.service 2>/dev/null; then
        systemctl disable stem.service || true
    fi
fi

exit 0
