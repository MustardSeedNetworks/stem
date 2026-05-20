#!/bin/sh
set -e

is_purge=0
case "$1" in
    purge|0)
        is_purge=1
        ;;
esac

if [ "$is_purge" -eq 1 ]; then
    if command -v ufw >/dev/null 2>&1; then
        ufw delete allow 8444/tcp >/dev/null 2>&1 || true
        ufw delete allow 8043/tcp >/dev/null 2>&1 || true
    fi
    if command -v firewall-cmd >/dev/null 2>&1 && systemctl is-active --quiet firewalld 2>/dev/null; then
        firewall-cmd --permanent --remove-port=8444/tcp >/dev/null 2>&1 || true
        firewall-cmd --permanent --remove-port=8043/tcp >/dev/null 2>&1 || true
        firewall-cmd --reload >/dev/null 2>&1 || true
    fi

    if getent passwd stem >/dev/null 2>&1; then
        userdel stem >/dev/null 2>&1 || true
    fi
    if getent group stem >/dev/null 2>&1; then
        groupdel stem >/dev/null 2>&1 || true
    fi

    rm -rf /etc/stem /var/lib/stem /var/log/stem
else
    echo "Stem removed. Data preserved in /var/lib/stem"
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
fi

exit 0
