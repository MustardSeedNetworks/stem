#!/bin/sh
set -e

BINARY=/usr/bin/stem
CONFIG_DIR=/etc/stem
AF_XDP_DROPIN_DIR=/etc/systemd/system/stem.service.d
AF_XDP_DROPIN="$AF_XDP_DROPIN_DIR/10-af-xdp-capability.conf"

if [ ! -f "$CONFIG_DIR/config.yaml" ] && [ -f /usr/share/stem/config.yaml ]; then
    cp /usr/share/stem/config.yaml "$CONFIG_DIR/config.yaml"
    chown root:stem "$CONFIG_DIR/config.yaml"
    chmod 640 "$CONFIG_DIR/config.yaml"
fi

if [ ! -f "$CONFIG_DIR/environment" ]; then
    cat > "$CONFIG_DIR/environment" <<'EOF'
# Stem environment variables
# STEM_AUTH_USERNAME=<your-admin-username>
# STEM_AUTH_PASSWORD=<choose-a-strong-unique-password>
# STEM_JWT_SECRET=generate-a-secure-random-string
# STEM_LICENSE_KEY=your-license-key
EOF
    chown root:stem "$CONFIG_DIR/environment"
    chmod 600 "$CONFIG_DIR/environment"
fi

if command -v setcap >/dev/null 2>&1; then
    capabilities=cap_net_raw,cap_net_admin,cap_net_bind_service
    if ldd "$BINARY" 2>/dev/null | grep -q libxdp && \
        command -v capsh >/dev/null 2>&1 && capsh --has-b=cap_bpf >/dev/null 2>&1; then
        capabilities="$capabilities,cap_bpf"
        install -d -m 755 "$AF_XDP_DROPIN_DIR"
        printf '%s\n' '[Service]' 'AmbientCapabilities=CAP_BPF' > "$AF_XDP_DROPIN"
    else
        rm -f "$AF_XDP_DROPIN"
    fi
    setcap "$capabilities=+ep" "$BINARY" || \
        echo "warning: could not set capabilities on $BINARY"
else
    rm -f "$AF_XDP_DROPIN"
    echo "warning: setcap not found; install libcap/libcap2-bin for non-root packet tests"
fi

if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q "Status: active"; then
    ufw allow 8444/tcp comment 'Stem WebUI HTTPS' >/dev/null 2>&1 || true
fi

if command -v firewall-cmd >/dev/null 2>&1 && systemctl is-active --quiet firewalld 2>/dev/null; then
    firewall-cmd --permanent --add-port=8444/tcp >/dev/null 2>&1 || true
    firewall-cmd --reload >/dev/null 2>&1 || true
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    systemctl enable stem.service >/dev/null 2>&1 || true
    if systemctl is-active --quiet stem.service 2>/dev/null; then
        systemctl restart stem.service || true
    else
        systemctl start stem.service || true
    fi
fi

cat <<'EOF'

==============================================
  The Stem installed successfully
==============================================

Web interface: https://localhost:8444 (self-signed certificate)
  Trust the cert: sudo stem install-ca

Quick start:
  1. Edit /etc/stem/environment to set credentials
  2. Restart: sudo systemctl restart stem

Commands:
  View logs:  journalctl -u stem -f
  CLI help:   stem --help

EOF

exit 0
