# =============================================================================
# Package Creation Targets
# =============================================================================
#
# Package creation for distribution:
#   - Debian packages (.deb) for Ubuntu/Debian
#   - RPM packages (.rpm) for RHEL/CentOS/Fedora
#   - macOS installer (.pkg)
#   - Multi-architecture support (AMD64/ARM64)
#
# =============================================================================

.PHONY: deb rpm pkg packages packages-all \
        deb-amd64 deb-arm64 rpm-amd64 rpm-arm64 _deb-arch _rpm-arch \
        container

# =============================================================================
# Package Variables
# =============================================================================

PKG_ARCH=$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
PKG_VERSION=$(shell echo $(VERSION) | sed 's/^v//;s/-dirty$$//;s/-[0-9]*-g[0-9a-f]*$$//')
DEB_ARCH=$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
RPM_ARCH=$(shell uname -m | sed 's/amd64/x86_64/;s/arm64/aarch64/')

# =============================================================================
# Debian Packages
# =============================================================================

deb: build-linux ## Build Debian package (.deb)
	@printf "$(BOLD)Building Debian package...$(RESET)\n"
	@mkdir -p dist/deb/DEBIAN
	@mkdir -p dist/deb/usr/bin
	@mkdir -p dist/deb/usr/lib/systemd/system
	@mkdir -p dist/deb/usr/share/stem
	@mkdir -p dist/deb/var/lib/stem
	@mkdir -p dist/deb/var/log/stem
	@cp bin/stem-linux dist/deb/usr/bin/stem
	@chmod 755 dist/deb/usr/bin/stem
	@if [ -f bin/iperf3 ]; then \
		cp bin/iperf3 dist/deb/usr/bin/stem-iperf3; \
		chmod 755 dist/deb/usr/bin/stem-iperf3; \
	fi
	@cp deploy/systemd/stem.service dist/deb/usr/lib/systemd/system/
	@cp deploy/config/stem.yaml dist/deb/usr/share/stem/config.yaml
	@sed 's/__VERSION__/$(PKG_VERSION)/g; s/__ARCHITECTURE__/$(DEB_ARCH)/g' \
		deploy/deb/control > dist/deb/DEBIAN/control
	@cp deploy/deb/postinst dist/deb/DEBIAN/
	@cp deploy/deb/prerm dist/deb/DEBIAN/
	@cp deploy/deb/postrm dist/deb/DEBIAN/
	@chmod 755 dist/deb/DEBIAN/postinst dist/deb/DEBIAN/prerm dist/deb/DEBIAN/postrm
	@dpkg-deb --build dist/deb dist/stem_$(PKG_VERSION)_$(DEB_ARCH).deb
	@printf "$(GREEN)Debian package: dist/stem_$(PKG_VERSION)_$(DEB_ARCH).deb$(RESET)\n"

deb-amd64: ## Build Debian package for amd64
	@$(MAKE) _deb-arch ARCH=amd64 CROSS_BINARY=stem-linux-amd64

deb-arm64: ## Build Debian package for arm64
	@$(MAKE) _deb-arch ARCH=arm64 CROSS_BINARY=stem-linux-arm64

_deb-arch:
	@printf "$(BOLD)Building Debian package for $(ARCH)...$(RESET)\n"
	@mkdir -p dist/deb-$(ARCH)/DEBIAN
	@mkdir -p dist/deb-$(ARCH)/usr/bin
	@mkdir -p dist/deb-$(ARCH)/usr/lib/systemd/system
	@mkdir -p dist/deb-$(ARCH)/usr/share/stem
	@mkdir -p dist/deb-$(ARCH)/var/lib/stem
	@mkdir -p dist/deb-$(ARCH)/var/log/stem
	@cp $(CROSS_BINARY) dist/deb-$(ARCH)/usr/bin/stem
	@chmod 755 dist/deb-$(ARCH)/usr/bin/stem
	@cp deploy/systemd/stem.service dist/deb-$(ARCH)/usr/lib/systemd/system/
	@cp deploy/config/stem.yaml dist/deb-$(ARCH)/usr/share/stem/config.yaml
	@sed 's/__VERSION__/$(PKG_VERSION)/g; s/__ARCHITECTURE__/$(ARCH)/g' \
		deploy/deb/control > dist/deb-$(ARCH)/DEBIAN/control
	@cp deploy/deb/postinst dist/deb-$(ARCH)/DEBIAN/
	@cp deploy/deb/prerm dist/deb-$(ARCH)/DEBIAN/
	@cp deploy/deb/postrm dist/deb-$(ARCH)/DEBIAN/
	@chmod 755 dist/deb-$(ARCH)/DEBIAN/postinst dist/deb-$(ARCH)/DEBIAN/prerm dist/deb-$(ARCH)/DEBIAN/postrm
	@dpkg-deb --build dist/deb-$(ARCH) dist/stem_$(PKG_VERSION)_$(ARCH).deb
	@printf "$(GREEN)dist/stem_$(PKG_VERSION)_$(ARCH).deb$(RESET)\n"

# =============================================================================
# RPM Packages
# =============================================================================

rpm: build-linux ## Build RPM package (.rpm)
	@printf "$(BOLD)Building RPM package...$(RESET)\n"
	@mkdir -p dist/rpm/BUILD dist/rpm/RPMS dist/rpm/SOURCES dist/rpm/SPECS dist/rpm/SRPMS
	@mkdir -p dist/rpm/SOURCES/stem-$(PKG_VERSION)
	@cp bin/stem-linux dist/rpm/SOURCES/stem-$(PKG_VERSION)/stem
	@if [ -f bin/iperf3 ]; then \
		cp bin/iperf3 dist/rpm/SOURCES/stem-$(PKG_VERSION)/stem-iperf3; \
	fi
	@cp deploy/systemd/stem.service dist/rpm/SOURCES/stem-$(PKG_VERSION)/
	@cp deploy/config/stem.yaml dist/rpm/SOURCES/stem-$(PKG_VERSION)/
	@sed 's/__VERSION__/$(PKG_VERSION)/g; s/__ARCHITECTURE__/$(RPM_ARCH)/g; s|%{_repo_root}|$(CURDIR)|g' \
		deploy/rpm/stem.spec > dist/rpm/SPECS/stem.spec
	@rpmbuild --define "_topdir $(CURDIR)/dist/rpm" \
		--define "_repo_root $(CURDIR)" \
		-bb dist/rpm/SPECS/stem.spec
	@mv dist/rpm/RPMS/$(RPM_ARCH)/*.rpm dist/ 2>/dev/null || true
	@printf "$(GREEN)RPM package: dist/stem-$(PKG_VERSION)-1.*.$(RPM_ARCH).rpm$(RESET)\n"

rpm-amd64: ## Build RPM package for amd64
	@$(MAKE) _rpm-arch ARCH=x86_64 CROSS_BINARY=stem-linux-amd64

rpm-arm64: ## Build RPM package for arm64
	@$(MAKE) _rpm-arch ARCH=aarch64 CROSS_BINARY=stem-linux-arm64

_rpm-arch:
	@printf "$(BOLD)Building RPM package for $(ARCH)...$(RESET)\n"
	@mkdir -p dist/rpm-$(ARCH)/BUILD dist/rpm-$(ARCH)/RPMS dist/rpm-$(ARCH)/SOURCES dist/rpm-$(ARCH)/SPECS dist/rpm-$(ARCH)/SRPMS
	@mkdir -p dist/rpm-$(ARCH)/SOURCES/stem-$(PKG_VERSION)
	@cp $(CROSS_BINARY) dist/rpm-$(ARCH)/SOURCES/stem-$(PKG_VERSION)/stem
	@cp deploy/systemd/stem.service dist/rpm-$(ARCH)/SOURCES/stem-$(PKG_VERSION)/
	@cp deploy/config/stem.yaml dist/rpm-$(ARCH)/SOURCES/stem-$(PKG_VERSION)/
	@sed 's/__VERSION__/$(PKG_VERSION)/g; s/__RPM_ARCH__/$(ARCH)/g; s|%{_repo_root}|$(CURDIR)|g' \
		deploy/rpm/stem.spec > dist/rpm-$(ARCH)/SPECS/stem.spec
	@rpmbuild --define "_topdir $(CURDIR)/dist/rpm-$(ARCH)" \
		--define "_repo_root $(CURDIR)" \
		--target $(ARCH) \
		-bb dist/rpm-$(ARCH)/SPECS/stem.spec
	@mv dist/rpm-$(ARCH)/RPMS/$(ARCH)/*.rpm dist/ 2>/dev/null || true
	@printf "$(GREEN)dist/stem-$(PKG_VERSION)-1.$(ARCH).rpm$(RESET)\n"

# =============================================================================
# macOS Package
# =============================================================================

pkg: build-darwin ## Build macOS installer package (.pkg)
	@if [ "$$(uname -s)" != "Darwin" ]; then \
		printf "$(RED)ERROR: macOS .pkg can only be built on macOS$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BOLD)Building macOS .pkg package...$(RESET)\n"
	@chmod +x deploy/macos/build-pkg.sh
	@./deploy/macos/build-pkg.sh ./bin/stem-darwin $(PKG_VERSION)
	@printf "$(GREEN)macOS package: dist/stem-$(PKG_VERSION)-$$(uname -m | sed 's/x86_64/amd64/').pkg$(RESET)\n"

# =============================================================================
# Multi-Package Targets
# =============================================================================

packages: deb rpm ## Build both .deb and .rpm packages
	@printf "$(GREEN)All packages built in dist/$(RESET)\n"
	@ls -la dist/*.deb dist/*.rpm 2>/dev/null || true

packages-all: ## Build .deb and .rpm for both amd64 and arm64
	@printf "$(BOLD)Building packages for all architectures...$(RESET)\n"
	@$(MAKE) deb-amd64
	@$(MAKE) rpm-amd64
	@$(MAKE) deb-arm64
	@$(MAKE) rpm-arm64
	@printf "$(GREEN)All packages built:$(RESET)\n"
	@ls -la dist/*.deb dist/*.rpm 2>/dev/null || true

# =============================================================================
# Container Images (Pack/Buildpacks) - LOCAL DEV ONLY
# =============================================================================
# NOTE: No public registry pushing during development.
# License validation required before commercial distribution.

CONTAINER_IMAGE := stem

container: ## Build container image locally (Pack/Buildpacks)
	@printf "$(BOLD)Building container with Pack (local only)...$(RESET)\n"
	@pack build $(CONTAINER_IMAGE):$(VERSION) \
		--builder paketobuildpacks/builder-jammy-base \
		--env BP_GO_TARGETS="./cmd/stem" \
		--env BP_GO_BUILD_LDFLAGS="-s -w -X $(VERSION_PKG).semver=$(VERSION) -X $(VERSION_PKG).commit=$(COMMIT)"
	@printf "$(GREEN)Container: $(CONTAINER_IMAGE):$(VERSION) (local)$(RESET)\n"
