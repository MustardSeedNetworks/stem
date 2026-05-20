#!/bin/sh
set -e

BINARY=/usr/bin/stem
CONFIG_DIR=/etc/stem

if [ ! -f "$CONFIG_DIR/config.yaml" ] && [ -f /usr/share/stem/config.yaml ]; then
    cp /usr/share/stem/config.yaml "$CONFIG_DIR/config.yaml"
    chown root:stem "$CONFIG_DIR/config.yaml"
    chmod 640 "$CONFIG_DIR/config.yaml"
fi

if [ ! -f "$CONFIG_DIR/environment" ]; then
    cat > "$CONFIG_DIR/environment" <<'EOF'
# Stem environment variables
# STEM_AUTH_USERNAME=admin
# STEM_AUTH_PASSWORD=changeme
# STEM_JWT_SECRET=generate-a-secure-random-string
# STEM_LICENSE_KEY=your-license-key
EOF
    chown root:stem "$CONFIG_DIR/environment"
    chmod 600 "$CONFIG_DIR/environment"
fi

if command -v setcap >/dev/null 2>&1; then
    setcap 'cap_net_raw,cap_net_admin,cap_net_bind_service=+ep' "$BINARY" || \
        echo "warning: could not set capabilities on $BINARY"
else
    echo "warning: setcap not found; install libcap/libcap2-bin for non-root packet tests"
fi

if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q "Status: active"; then
    ufw allow 8444/tcp comment 'Stem WebUI HTTPS' >/dev/null 2>&1 || true
    ufw allow 8043/tcp comment 'Stem HTTP→HTTPS redirector' >/dev/null 2>&1 || true
fi

if command -v firewall-cmd >/dev/null 2>&1 && systemctl is-active --quiet firewalld 2>/dev/null; then
    firewall-cmd --permanent --add-port=8444/tcp >/dev/null 2>&1 || true
    firewall-cmd --permanent --add-port=8043/tcp >/dev/null 2>&1 || true
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
