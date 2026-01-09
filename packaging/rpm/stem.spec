# Disable debug package (Go binary is stripped)
%define debug_package %{nil}

Name:           stem
Version:        0.2.8
Release:        1%{?dist}
Summary:        The Stem - Network Performance Testing Tool

License:        Proprietary
URL:            https://github.com/krisarmstrong/stem
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  gcc
BuildRequires:  make
BuildRequires:  systemd-rpm-macros
BuildRequires:  libbpf-devel
BuildRequires:  libxdp-devel
BuildRequires:  elfutils-libelf-devel
BuildRequires:  zlib-devel

Requires:       systemd
Requires:       libbpf
Requires:       libxdp
Recommends:     firewalld

%description
The Stem is a high-performance network testing tool supporting:
- RFC 2544 benchmarking (throughput, latency, frame loss)
- ITU-T Y.1564 service activation testing
- ITU-T Y.1731 Ethernet OAM
- Packet reflection for remote testing
- WebUI, TUI, and CLI interfaces

%prep
%setup -q

%build
# Build C dataplane library first (required for CGO)
make dataplane
# Build Go binary with CGO linking to C dataplane
make build VERSION=%{version}

%install
# Binary
install -D -m 0755 bin/stem-linux %{buildroot}%{_bindir}/stem

# Systemd service
install -D -m 0644 packaging/systemd/stem.service %{buildroot}%{_unitdir}/stem.service

# Configuration
install -D -m 0640 packaging/config/stem.yaml %{buildroot}%{_sysconfdir}/stem/config.yaml
install -d -m 0750 %{buildroot}%{_sysconfdir}/stem

# Environment file (for secrets)
cat > %{buildroot}%{_sysconfdir}/stem/environment << 'EOF'
# Stem environment variables
# STEM_AUTH_USERNAME=admin
# STEM_AUTH_PASSWORD=changeme
# STEM_JWT_SECRET=generate-a-secure-random-string
# STEM_LICENSE_KEY=your-license-key
EOF
chmod 0600 %{buildroot}%{_sysconfdir}/stem/environment

# State and log directories
install -d -m 0750 %{buildroot}%{_sharedstatedir}/stem
install -d -m 0750 %{buildroot}%{_localstatedir}/log/stem

# Firewalld service definition
install -d -m 0755 %{buildroot}%{_prefix}/lib/firewalld/services
cat > %{buildroot}%{_prefix}/lib/firewalld/services/stem.xml << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<service>
  <short>Stem</short>
  <description>The Stem - Network Performance Testing WebUI</description>
  <port protocol="tcp" port="8080"/>
  <port protocol="tcp" port="8443"/>
</service>
EOF
chmod 0644 %{buildroot}%{_prefix}/lib/firewalld/services/stem.xml

%pre
# Create stem user/group if they don't exist
getent group stem >/dev/null || groupadd -r stem
getent passwd stem >/dev/null || \
    useradd -r -g stem -d %{_sharedstatedir}/stem -s /sbin/nologin \
    -c "The Stem Network Testing" stem
exit 0

%post
%systemd_post stem.service

# Find available port if default is in use
find_available_port() {
    local start_port=$1
    local port=$start_port
    while [ $port -lt $((start_port + 100)) ]; do
        if ! ss -tlnp | grep -q ":${port} "; then
            echo $port
            return 0
        fi
        port=$((port + 1))
    done
    echo $start_port
}

# Check and update ports if needed
CONFIG_FILE="%{_sysconfdir}/stem/config.yaml"
if [ -f "$CONFIG_FILE" ]; then
    # Check HTTP port
    HTTP_PORT=$(grep -E '^\s+port:\s*8080' "$CONFIG_FILE" | head -1)
    if [ -n "$HTTP_PORT" ] && ss -tlnp | grep -q ":8080 "; then
        NEW_PORT=$(find_available_port 8080)
        if [ "$NEW_PORT" != "8080" ]; then
            sed -i "s/port: 8080/port: $NEW_PORT/" "$CONFIG_FILE"
            echo "Note: Port 8080 in use, configured to use port $NEW_PORT"
        fi
    fi

    # Check HTTPS port
    HTTPS_PORT=$(grep -E '^\s+tls_port:\s*8443' "$CONFIG_FILE" | head -1)
    if [ -n "$HTTPS_PORT" ] && ss -tlnp | grep -q ":8443 "; then
        NEW_PORT=$(find_available_port 8443)
        if [ "$NEW_PORT" != "8443" ]; then
            sed -i "s/tls_port: 8443/tls_port: $NEW_PORT/" "$CONFIG_FILE"
            echo "Note: Port 8443 in use, configured to use port $NEW_PORT"
        fi
    fi
fi

# Configure firewall if firewalld is running
if systemctl is-active --quiet firewalld; then
    firewall-cmd --reload 2>/dev/null || true
    firewall-cmd --permanent --add-service=stem 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
    echo "Firewall configured for Stem service"
fi

# Set permissions
chown -R stem:stem %{_sharedstatedir}/stem
chown -R stem:stem %{_localstatedir}/log/stem
chown root:stem %{_sysconfdir}/stem
chown root:stem %{_sysconfdir}/stem/config.yaml
chown root:stem %{_sysconfdir}/stem/environment

echo ""
echo "=============================================="
echo "The Stem has been installed!"
echo ""
echo "Quick start:"
echo "  1. Edit /etc/stem/environment to set credentials"
echo "  2. systemctl enable --now stem"
echo "  3. Access WebUI at http://localhost:8080"
echo ""
echo "For CLI usage: stem --help"
echo "=============================================="

%preun
%systemd_preun stem.service

%postun
%systemd_postun_with_restart stem.service

# Remove firewall rule on uninstall
if [ $1 -eq 0 ]; then
    if systemctl is-active --quiet firewalld; then
        firewall-cmd --permanent --remove-service=stem 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
    fi
fi

%files
%license LICENSE
%doc README.md
%{_bindir}/stem
%{_unitdir}/stem.service
%dir %attr(0750,root,stem) %{_sysconfdir}/stem
%config(noreplace) %attr(0640,root,stem) %{_sysconfdir}/stem/config.yaml
%config(noreplace) %attr(0600,root,stem) %{_sysconfdir}/stem/environment
%dir %attr(0750,stem,stem) %{_sharedstatedir}/stem
%dir %attr(0750,stem,stem) %{_localstatedir}/log/stem
%{_prefix}/lib/firewalld/services/stem.xml

%changelog
* Wed Jan 08 2025 Kris Armstrong <kris@mustardseednetworks.com> - 0.2.8-1
- Added complete configuration documentation to help system
- Implemented i18n infrastructure with English and Spanish support
- All tests pass with lint and format compliance

* Wed Jan 08 2025 Kris Armstrong <kris@mustardseednetworks.com> - 0.2.7-1
- C23 lint compliance (clang-tidy)
- Fixed data race in test executor
- Added systemd service support
- Added firewalld integration
