Name:       stem
Version:    __VERSION__
Release:    1%{?dist}
Summary:    The Stem - Network Performance Testing Tool by Mustard Seed Networks
License:    Proprietary
URL:        https://github.com/krisarmstrong/stem
BuildArch:  __ARCHITECTURE__

Requires:   systemd, libcap
Requires(pre): shadow-utils
Provides: user(stem)
Provides: group(stem)

%description
The Stem is a high-performance network testing tool supporting:
- RFC 2544 benchmarking (throughput, latency, frame loss)
- ITU-T Y.1564 service activation testing
- ITU-T Y.1731 Ethernet OAM
- Packet reflection for remote testing
- WebUI, TUI, and CLI interfaces

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/lib/systemd/system
mkdir -p %{buildroot}/etc/stem
mkdir -p %{buildroot}/usr/share/stem
mkdir -p %{buildroot}/var/lib/stem
mkdir -p %{buildroot}/var/log/stem

# Copy binary (single binary with embedded assets)
install -m 755 %{_repo_root}/bin/stem-linux %{buildroot}/usr/bin/stem

# Copy systemd service file
install -m 644 %{_repo_root}/deploy/systemd/stem.service %{buildroot}/usr/lib/systemd/system/stem.service

# Copy default configuration
install -m 640 %{_repo_root}/deploy/config/stem.yaml %{buildroot}/usr/share/stem/config.yaml

%files
%attr(755, root, root) /usr/bin/stem
%attr(644, root, root) /usr/lib/systemd/system/stem.service
%attr(640, root, stem) /usr/share/stem/config.yaml
%dir %attr(750, root, stem) /etc/stem
%dir %attr(750, stem, stem) /var/lib/stem
%dir %attr(750, stem, stem) /var/log/stem

%pre
# Create service user and group
getent group stem >/dev/null || groupadd -r stem
getent passwd stem >/dev/null || \
    useradd -r -g stem -d /var/lib/stem -s /sbin/nologin \
    -c "The Stem Network Performance Testing" stem
exit 0

%post
# Set ownership of directories
chown -R stem:stem /var/lib/stem /var/log/stem
chown root:stem /etc/stem

# Install default config if not present
if [ ! -f /etc/stem/config.yaml ]; then
    cp /usr/share/stem/config.yaml /etc/stem/config.yaml
    chown root:stem /etc/stem/config.yaml
    chmod 640 /etc/stem/config.yaml
fi

# Create environment file if not present
if [ ! -f /etc/stem/environment ]; then
    cat > /etc/stem/environment << 'EOF'
# Stem environment variables
# STEM_AUTH_USERNAME=<your-admin-username>
# STEM_AUTH_PASSWORD=<choose-a-strong-unique-password>
# STEM_JWT_SECRET=generate-a-secure-random-string
# STEM_LICENSE_KEY=your-license-key
EOF
    chown root:stem /etc/stem/environment
    chmod 600 /etc/stem/environment
fi

# Set capabilities for raw socket access
# - CAP_NET_RAW: Required for raw packet I/O
# - CAP_NET_ADMIN: Required for interface control
# - CAP_NET_BIND_SERVICE: Required for binding to privileged ports
/usr/sbin/setcap 'cap_net_raw,cap_net_admin,cap_net_bind_service=+ep' /usr/bin/stem || true

# Configure firewall if firewalld is running
if systemctl is-active --quiet firewalld 2>/dev/null; then
    firewall-cmd --permanent --add-port=8444/tcp 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
    echo "Firewall configured for Stem service (port 8444)"
fi

%systemd_post stem.service

%preun
%systemd_preun stem.service

%postun
%systemd_postun_with_restart stem.service

# On complete removal (not upgrade), clean up
if [ $1 -eq 0 ]; then
    # Remove firewall rules
    if systemctl is-active --quiet firewalld 2>/dev/null; then
        firewall-cmd --permanent --remove-port=8444/tcp 2>/dev/null || true
        firewall-cmd --permanent --remove-port=8043/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
    fi

    # Remove user/group
    userdel stem 2>/dev/null || true
    groupdel stem 2>/dev/null || true
fi

%changelog
* Thu Jan 23 2025 Kris Armstrong <kris@mustardseednetworks.com>
- Converted to binary package format for CI/CD release workflow
- Added capabilities (NET_RAW, NET_ADMIN, NET_BIND_SERVICE)
- Added firewalld integration
- Added environment file for secrets management
