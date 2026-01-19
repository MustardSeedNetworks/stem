# =============================================================================
# Package Creation Targets
# =============================================================================
#
# Package creation for distribution:
#   - Debian packages (.deb) for Ubuntu/Debian
#   - RPM packages (.rpm) for RHEL/CentOS/Fedora
#   - Systemd service installation
#
# =============================================================================

.PHONY: deb rpm packages install-service

# =============================================================================
# Debian Package
# =============================================================================

deb: build ## Build Debian package (.deb)
	@echo "Building Debian package..."
	@mkdir -p pkg/deb/DEBIAN
	@mkdir -p pkg/deb/usr/local/bin
	@mkdir -p pkg/deb/etc/systemd/system
	@cp bin/stem-linux pkg/deb/usr/local/bin/stem
	@chmod 755 pkg/deb/usr/local/bin/stem
	@echo "Package: stem" > pkg/deb/DEBIAN/control
	@echo "Version: $(VERSION)" >> pkg/deb/DEBIAN/control
	@echo "Section: net" >> pkg/deb/DEBIAN/control
	@echo "Priority: optional" >> pkg/deb/DEBIAN/control
	@echo "Architecture: amd64" >> pkg/deb/DEBIAN/control
	@echo "Maintainer: Mustard Seed Networks" >> pkg/deb/DEBIAN/control
	@echo "Description: Network Performance Testing Tool" >> pkg/deb/DEBIAN/control
	@dpkg-deb --build pkg/deb stem-$(VERSION)-amd64.deb
	@rm -rf pkg/deb
	$(call success,Debian package built: stem-$(VERSION)-amd64.deb)

# =============================================================================
# RPM Package
# =============================================================================

rpm: build ## Build RPM package (.rpm)
	@echo "Building RPM package..."
	@echo "Note: Requires rpmbuild to be installed"
	@mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	@cp bin/stem-linux ~/rpmbuild/SOURCES/stem
	@echo "Name: stem" > ~/rpmbuild/SPECS/stem.spec
	@echo "Version: $(VERSION)" >> ~/rpmbuild/SPECS/stem.spec
	@echo "Release: 1" >> ~/rpmbuild/SPECS/stem.spec
	@echo "Summary: Network Performance Testing Tool" >> ~/rpmbuild/SPECS/stem.spec
	@echo "License: Proprietary" >> ~/rpmbuild/SPECS/stem.spec
	@echo "%description" >> ~/rpmbuild/SPECS/stem.spec
	@echo "The Stem - Network Performance Testing Tool" >> ~/rpmbuild/SPECS/stem.spec
	@echo "%install" >> ~/rpmbuild/SPECS/stem.spec
	@echo "mkdir -p %{buildroot}/usr/local/bin" >> ~/rpmbuild/SPECS/stem.spec
	@echo "cp %{SOURCE0} %{buildroot}/usr/local/bin/stem" >> ~/rpmbuild/SPECS/stem.spec
	@echo "%files" >> ~/rpmbuild/SPECS/stem.spec
	@echo "/usr/local/bin/stem" >> ~/rpmbuild/SPECS/stem.spec
	rpmbuild -bb ~/rpmbuild/SPECS/stem.spec
	$(call success,RPM package built)

# =============================================================================
# Combined Targets
# =============================================================================

packages: deb rpm ## Build all packages (deb + rpm)

# =============================================================================
# Systemd Service Installation
# =============================================================================

install-service: build ## Install as systemd service (requires root)
	@echo "Installing systemd service..."
	install -D -m 0755 bin/stem-linux /usr/bin/stem
	install -D -m 0644 deploy/systemd/stem.service /lib/systemd/system/stem.service
	install -D -m 0640 deploy/config/stem.yaml /etc/stem/config.yaml
	@if ! getent group stem >/dev/null; then groupadd -r stem; fi
	@if ! getent passwd stem >/dev/null; then \
		useradd -r -g stem -d /var/lib/stem -s /sbin/nologin stem; \
	fi
	install -d -m 0750 -o stem -g stem /var/lib/stem
	install -d -m 0750 -o stem -g stem /var/log/stem
	systemctl daemon-reload
	@echo "Service installed. Run: systemctl enable --now stem"
