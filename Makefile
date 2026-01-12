# The Stem Makefile
# Mustard Seed Networks - Network Performance Testing
#
# Targets:
#   make           - Build everything (UI + Go binary)
#   make ui        - Build React WebUI only
#   make go        - Build Go binary only
#   make clean     - Clean build artifacts
#   make dev       - Run development servers
#   make test      - Run all tests

# Include shared infrastructure (version, colors, timers)
include Makefile.common

# Go settings
GO := go
VERSION_PKG := github.com/krisarmstrong/stem/internal/version
LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).BuildTime=$(BUILD_TIME)
GOFLAGS := -trimpath -buildvcs=false -ldflags "$(LDFLAGS)"

# Binary name
BINARY := stem

# Platform-specific binary name (uses PLATFORM from Makefile.common)
ifeq ($(PLATFORM),darwin)
    BINARY_NAME := bin/$(BINARY)-darwin
else ifeq ($(PLATFORM),linux)
    BINARY_NAME := bin/$(BINARY)-linux
else
    BINARY_NAME := bin/$(BINARY)
endif

# ============================================================================
# Main Targets
# ============================================================================

.PHONY: all ui ui-deps go clean test dev install help lint lint-go lint-c format format-go format-c fix verify build-linux-docker update update-go update-npm version-check tools tools-go tools-frontend security security-backend security-frontend security-secrets license-check license-check-go license-check-npm license-report deb rpm packages c-build c-build-docker dataplane

# Default: build everything
all: ui go
	@echo ""
	@echo "Build complete!"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"

# Build React WebUI
ui-deps:
	@echo "Installing UI dependencies..."
	cd ui && npm install

ui: ui-deps
	@echo "Building React WebUI..."
	cd ui && npm run build
	@echo "Copying dist to pkg/web for embedding..."
	mkdir -p internal/web/dist
	cp -r ui/dist/* internal/web/dist/

# Build Go binary
go:
	@echo "Building $(BINARY)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/stem/
	@echo "Built: $(BINARY_NAME)"

# Quick build (Go only, assumes UI is already built)
quick:
	@echo "Quick build (Go only)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/stem/

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf ui/dist/
	rm -rf ui/node_modules/
	rm -rf internal/web/dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# ============================================================================
# Development
# ============================================================================

# Run UI dev server
ui-dev:
	cd ui && npm run dev

# Run Go backend
go-dev:
	$(GO) run ./cmd/stem/ web -p 8080

# Development mode (run both)
dev:
	@echo "Starting development servers..."
	@echo "UI: http://localhost:3000"
	@echo "API: http://localhost:8080"
	@echo ""
	@echo "Run in separate terminals:"
	@echo "  make ui-dev    # React dev server"
	@echo "  make go-dev    # Go backend"

# ============================================================================
# Linting & Formatting
# ============================================================================

# Run all linters
lint: lint-go
	@echo "✓ All linters passed"

# Run Go linters (golangci-lint v2)
lint-go:
	@echo "Running Go linter (golangci-lint)..."
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi; \
	$$GOLANGCI_LINT run --allow-parallel-runners ./...
	@echo "✓ Go lint passed"

# Run C linter (clang-tidy) - Linux only
lint-c:
ifeq ($(UNAME),Linux)
	@echo "Running C linter (clang-tidy)..."
	@if ! command -v clang-format >/dev/null 2>&1; then \
		echo "clang-format not found; install it to enforce formatting."; \
		exit 1; \
	fi
	@if ! command -v clang-tidy >/dev/null 2>&1; then \
		echo "clang-tidy not found; install it to enforce linting."; \
		exit 1; \
	fi
	@if [ -f build/compile_commands.json ]; then \
		clang_tidy_db=build; \
	elif [ -f compile_commands.json ]; then \
		clang_tidy_db=.; \
	else \
		echo "compile_commands.json not found. Generate with: bear -- make dataplane c-test"; \
		exit 1; \
	fi; \
	find src include tests -type f \( -name '*.c' -o -name '*.h' \) | xargs clang-format --dry-run --Werror; \
	find src include tests -type f -name '*.c' | xargs clang-tidy -p $$clang_tidy_db -warnings-as-errors=*
	@echo "✓ C lint complete"
else
	@echo "C linting requires Linux"
endif

# Format all code
format: format-go
	@echo "✓ All code formatted"

# Format Go code
format-go:
	@echo "Formatting Go code..."
	@gofmt -w -s .
	@echo "✓ Go code formatted"

# Format C code - Linux only
format-c:
ifeq ($(UNAME),Linux)
	@echo "Formatting C code..."
	@if ! command -v clang-format >/dev/null 2>&1; then \
		echo "clang-format not found; install it to format C code."; \
		exit 1; \
	fi
	find src include tests -type f \( -name '*.c' -o -name '*.h' \) | xargs clang-format -i
	@echo "✓ C code formatted"
else
	@echo "C formatting requires Linux"
endif

# Auto-fix linting issues
fix:
	@echo "Auto-fixing Go code..."
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi; \
	$$GOLANGCI_LINT run --fix ./...
	@gofmt -w -s .
	@echo "✓ Auto-fix complete"

# ============================================================================
# Verification (CI/CD pipeline)
# ============================================================================

# Full verification: lint, test, security, build
verify: ## Full verification pipeline (lint, test, security, build)
	@printf "\n$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  VERIFICATION PIPELINE$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n\n"
	$(call timer-start,verify)
	@printf "$(BOLD)[1/5]$(RESET) Running linters...\n"
	@$(MAKE) lint
	@printf "$(BOLD)[2/5]$(RESET) Running tests...\n"
	@$(MAKE) test
	@printf "$(BOLD)[3/5]$(RESET) Running security scans...\n"
	@$(MAKE) security
	@printf "$(BOLD)[4/5]$(RESET) Building binary...\n"
	@$(MAKE) build
	@printf "$(BOLD)[5/5]$(RESET) Checking licenses...\n"
	@$(MAKE) license-check || true
	$(call timer-end,verify,Verification pipeline)
	@printf "\n$(BOLD)$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BOLD)$(GREEN)  ✓ VERIFICATION COMPLETE$(RESET)\n"
	@printf "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "    Binary: $(BINARY_NAME)\n"
	@printf "    Version: $(VERSION)\n\n"

# ============================================================================
# Testing
# ============================================================================

# Run Go tests
test:
	@echo "Running Go tests..."
	$(GO) test -v -race ./internal/... ./cmd/...

# Run tests with coverage
test-coverage:
	@echo "Running Go tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/...
	$(GO) tool cover -func=coverage.out

# Generate HTML coverage report
test-coverage-html: test-coverage
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ============================================================================
# Installation
# ============================================================================

# Install to /usr/local/bin
install: all
	@echo "Installing $(BINARY) to /usr/local/bin..."
	install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY)
	@echo "Installed: /usr/local/bin/$(BINARY)"

# Uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY)
	@echo "Uninstalled: /usr/local/bin/$(BINARY)"

# ============================================================================
# C/Dataplane Build Configuration
# ============================================================================
# The C dataplane code has multiple build modes for different environments.
#
# Build Profiles:
#   c-build           Build dataplane for current platform (Linux only)
#   c-build-stub      Build stub for non-Linux platforms (for testing)
#   c-test            Run C unit tests
#
# Platform Support:
#   Linux:   Full support with AF_PACKET/AF_XDP backends
#   macOS:   Stub mode only (for development/testing)
#   Docker:  Use 'make c-build-docker' for Linux build from any platform
#
# DPDK Support (Optional):
#   DPDK provides maximum performance but adds complexity.
#   Enable with: make c-build DPDK=1
#
#   Requirements for DPDK:
#   - DPDK 23.11 LTS installed
#   - Hugepages configured
#   - DPDK_DIR environment variable set
#
#   Without DPDK (default):
#   - Uses AF_PACKET or AF_XDP
#   - Easier to deploy
#   - Good performance for most use cases
# ============================================================================

# C compiler settings
CC := gcc
CFLAGS := -Wall -Wextra -O3 -march=native -pthread -Iinclude
C_LDFLAGS := -pthread -lm

# C sources - both dataplane and reflector (excluding main.c)
C_DATAPLANE_SRCS := $(wildcard src/dataplane/common/*.c)
C_REFLECTOR_SRCS := $(filter-out src/reflector/main.c,$(wildcard src/reflector/*.c))
C_ALL_SRCS := $(C_DATAPLANE_SRCS) $(C_REFLECTOR_SRCS)
C_ALL_OBJS := $(C_ALL_SRCS:.c=.o)

# Build C dataplane library (Linux only)
dataplane:
ifeq ($(UNAME),Linux)
	@echo "Building C dataplane + reflector library..."
	@for src in $(C_ALL_SRCS); do \
		$(CC) $(CFLAGS) -c $$src -o $${src%.c}.o; \
	done
	mkdir -p build
	ar rcs build/libreflector.a $(C_ALL_OBJS)
	cp build/libreflector.a librfc2544.a
	@echo "Built: build/libreflector.a"
else
	@echo "Dataplane requires Linux (uses AF_PACKET/AF_XDP)"
endif

# Build C code using Docker (cross-platform)
c-build-docker: ## Build C code using Docker (cross-platform)
	@echo "Building C dataplane in Docker..."
	docker run --rm -v $(PWD):/src -w /src gcc:latest make c-build

# Alias for dataplane target
c-build: dataplane

# ============================================================================
# C Tests (Linux only)
# ============================================================================

# C test sources
C_TEST_SRCS := $(wildcard tests/c/*.c)
C_TEST_BINS := $(patsubst tests/c/%.c,bin/test_%,$(C_TEST_SRCS))

# Sources for pacing unit tests (minimal dependencies)
C_PACING_SRCS := src/dataplane/common/pacing.c

# Sources for protocol tests (with stub dependencies)
C_PROTO_SRCS := src/dataplane/common/pacing.c \
	src/dataplane/common/y1564.c \
	src/dataplane/common/y1731.c \
	src/dataplane/common/tsn.c \
	src/dataplane/common/mef.c \
	src/dataplane/common/rfc2889.c \
	src/dataplane/common/rfc6349.c \
	tests/c/test_stubs.c

# Build C tests
c-test:
ifeq ($(UNAME),Linux)
	@echo "Building C tests..."
	mkdir -p bin
	$(CC) $(CFLAGS) -o bin/test_pacing tests/c/test_pacing.c $(C_PACING_SRCS) $(C_LDFLAGS)
	$(CC) $(CFLAGS) -o bin/test_protocols tests/c/test_protocols.c $(C_PROTO_SRCS) $(C_LDFLAGS)
	@echo "Running C tests..."
	./bin/test_pacing
	./bin/test_protocols
else
	@echo "C tests require Linux"
endif

# Run smoke tests (requires root)
smoke-test:
ifeq ($(UNAME),Linux)
	@echo "Running smoke tests..."
	sudo tests/smoke/run_smoke_tests.sh
else
	@echo "Smoke tests require Linux"
endif

# ============================================================================
# Combined Build Targets
# ============================================================================

# Build binary (creates symlink for convenience)
build: go
	mkdir -p bin
	ln -sf $(notdir $(BINARY_NAME)) bin/stem 2>/dev/null || cp $(BINARY_NAME) bin/stem

# Full test suite
test-all: test c-test
	@echo "All tests complete"

# ============================================================================
# Package Generation
# ============================================================================
# Build distributable packages for Linux systems

.PHONY: deb rpm packages

# Build Debian package
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

# Build RPM package (requires rpmbuild)
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

# Build all packages
packages: deb rpm ## Build all packages (deb + rpm)

# Install systemd service (requires root)
install-service: build
	@echo "Installing systemd service..."
	install -D -m 0755 bin/stem-linux /usr/bin/stem
	install -D -m 0644 packaging/systemd/stem.service /lib/systemd/system/stem.service
	install -D -m 0640 packaging/config/stem.yaml /etc/stem/config.yaml
	@if ! getent group stem >/dev/null; then groupadd -r stem; fi
	@if ! getent passwd stem >/dev/null; then \
		useradd -r -g stem -d /var/lib/stem -s /sbin/nologin stem; \
	fi
	install -d -m 0750 -o stem -g stem /var/lib/stem
	install -d -m 0750 -o stem -g stem /var/log/stem
	systemctl daemon-reload
	@echo "Service installed. Run: systemctl enable --now stem"

# ============================================================================
# Docker Build (Cross-Platform)
# ============================================================================

# Build Linux binary via Docker (for cross-compilation from macOS)
build-linux-docker: ## Build Linux binary using Docker (cross-platform)
	@printf "$(BOLD)$(CYAN)Building Linux binary via Docker...$(RESET)\n"
	docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		golang:1.25 \
		go build -trimpath -buildvcs=false \
			-ldflags "-s -w \
				-X $(VERSION_PKG).Version=$(VERSION) \
				-X $(VERSION_PKG).Commit=$(COMMIT) \
				-X $(VERSION_PKG).BuildTime=$(BUILD_TIME)" \
			-o bin/stem-linux-amd64 \
			./cmd/stem/
	@printf "$(GREEN)✓ Linux binary built: bin/stem-linux-amd64$(RESET)\n"

# ============================================================================
# Dependency Management
# ============================================================================

# Update all dependencies
update: update-go update-npm ## Update all dependencies
	@printf "\n$(GREEN)✓ All dependencies updated$(RESET)\n"
	@printf "$(YELLOW)Remember to test before committing!$(RESET)\n"

# Update Go dependencies
update-go: ## Update Go modules
	@printf "$(BOLD)$(CYAN)Updating Go dependencies...$(RESET)\n"
	$(call timer-start,update-go)
	go get -u ./...
	go mod tidy
	$(call timer-end,update-go,Go dependencies update)

# Update npm dependencies
update-npm: ## Update npm packages
	@printf "$(BOLD)$(CYAN)Updating npm dependencies...$(RESET)\n"
	$(call timer-start,update-npm)
	cd ui && npm update
	cd ui && npm audit fix || true
	$(call timer-end,update-npm,npm dependencies update)

# Show version information
version-check: ## Show version info and outdated packages
	@printf "$(BOLD)$(CYAN)Version Information$(RESET)\n"
	@printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	@printf "$(BOLD)Runtime:$(RESET)\n"
	@printf "  Go:              $$(go version | awk '{print $$3}')\n"
	@printf "  Node.js:         $$(node --version)\n"
	@printf "  npm:             $$(npm --version)\n"
	@printf "\n$(BOLD)Go Tools:$(RESET)\n"
	@printf "  golangci-lint:   $$(golangci-lint --version 2>/dev/null | head -1 || echo 'not installed')\n"
	@printf "  govulncheck:     $$(govulncheck -version 2>/dev/null || echo 'not installed')\n"
	@printf "  staticcheck:     $$(staticcheck -version 2>/dev/null || echo 'not installed')\n"
	@printf "\n$(BOLD)Dependencies:$(RESET)\n"
	@printf "  Go modules:      $$(go list -m all 2>/dev/null | wc -l | tr -d ' ') packages\n"
	@printf "  npm packages:    $$(cd ui && npm ls --depth=0 2>/dev/null | wc -l | tr -d ' ') packages\n"
	@printf "\n$(BOLD)Outdated:$(RESET)\n"
	@GO_OUTDATED=$$(go list -u -m all 2>/dev/null | grep '\[' | wc -l | tr -d ' '); \
	printf "  Go outdated:     $$GO_OUTDATED packages\n"
	@cd ui && npm outdated 2>/dev/null | tail -n +2 | wc -l | xargs -I {} printf "  npm outdated:    {} packages\n"

# ============================================================================
# Developer Tools
# ============================================================================

# Install all Go development tools
tools-go: ## Install Go development tools
	@printf "$(BOLD)$(CYAN)Installing Go development tools...$(RESET)\n"
	$(call timer-start,tools-go)
	@printf "  Installing golangci-lint v2...\n"
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@printf "  Installing govulncheck...\n"
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@printf "  Installing gosec...\n"
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@printf "  Installing gofumpt...\n"
	go install mvdan.cc/gofumpt@latest
	@printf "  Installing goimports...\n"
	go install golang.org/x/tools/cmd/goimports@latest
	@printf "  Installing staticcheck...\n"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@printf "  Installing gitleaks...\n"
	go install github.com/zricethezav/gitleaks/v8@latest
	@printf "  Installing gotestsum...\n"
	go install gotest.tools/gotestsum@latest
	$(call timer-end,tools-go,Tool installation)
	@printf "\n$(GREEN)✓ All Go tools installed to $$(go env GOPATH)/bin$(RESET)\n"
	@printf "\nInstalled tools:\n"
	@printf "  • golangci-lint  - Comprehensive Go linter\n"
	@printf "  • govulncheck    - Go vulnerability checker\n"
	@printf "  • gosec          - Go security scanner\n"
	@printf "  • gofumpt        - Stricter gofmt\n"
	@printf "  • goimports      - Import management\n"
	@printf "  • staticcheck    - Static analysis\n"
	@printf "  • gitleaks       - Secret detection\n"
	@printf "  • gotestsum      - Better test output\n"

# Install frontend tools
tools-frontend: ## Install frontend development tools
	@printf "$(BOLD)$(CYAN)Installing frontend tools...$(RESET)\n"
	cd ui && npm install
	@printf "$(GREEN)✓ Frontend tools installed$(RESET)\n"

# Install all tools
tools: tools-go tools-frontend ## Install all development tools
	@printf "\n$(GREEN)✓ All development tools installed$(RESET)\n"

# ============================================================================
# Security Scanning
# ============================================================================

# Run all security checks
security: security-backend security-frontend security-secrets ## Run all security scans
	@printf "\n$(GREEN)✓ All security scans complete$(RESET)\n"

# Backend security (Go)
security-backend: ## Run Go security scans
	@printf "$(BOLD)$(CYAN)Running Go security scans...$(RESET)\n"
	$(call timer-start,security-backend)
	@printf "  [1/3] Running govulncheck...\n"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ govulncheck not installed (run: make tools-go)$(RESET)\n"; \
	fi
	@printf "  [2/3] Running gosec...\n"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ gosec not installed (run: make tools-go)$(RESET)\n"; \
	fi
	@printf "  [3/3] Running staticcheck...\n"
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ staticcheck not installed (run: make tools-go)$(RESET)\n"; \
	fi
	$(call timer-end,security-backend,Go security scan)

# Frontend security (npm)
security-frontend: ## Run npm security audit
	@printf "$(BOLD)$(CYAN)Running npm security audit...$(RESET)\n"
	$(call timer-start,security-frontend)
	cd ui && npm audit --audit-level=high || true
	$(call timer-end,security-frontend,npm security audit)

# Secret scanning
security-secrets: ## Scan for secrets in codebase
	@printf "$(BOLD)$(CYAN)Scanning for secrets...$(RESET)\n"
	$(call timer-start,security-secrets)
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --source . --verbose || true; \
	else \
		printf "$(YELLOW)⚠ gitleaks not installed (run: make tools-go)$(RESET)\n"; \
	fi
	$(call timer-end,security-secrets,Secret scan)

# ============================================================================
# License Compliance
# ============================================================================

# Check all licenses
license-check: license-check-go license-check-npm ## Check dependency licenses
	@printf "\n$(GREEN)✓ License check complete$(RESET)\n"

# Check Go dependency licenses
license-check-go: ## Check Go module licenses
	@printf "$(BOLD)$(CYAN)Checking Go dependency licenses...$(RESET)\n"
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		printf "$(YELLOW)Installing go-licenses...$(RESET)\n"; \
		go install github.com/google/go-licenses@latest; \
	fi
	@go-licenses check ./... \
		--disallowed_types=forbidden,restricted \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

# Check npm dependency licenses
license-check-npm: ## Check npm package licenses
	@printf "$(BOLD)$(CYAN)Checking npm dependency licenses...$(RESET)\n"
	@cd ui && npx license-checker --summary --onlyAllow \
		"MIT;Apache-2.0;BSD-2-Clause;BSD-3-Clause;ISC;CC0-1.0;Unlicense;0BSD" \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

# Generate license report
license-report: ## Generate full license report
	@printf "$(BOLD)$(CYAN)Generating license report...$(RESET)\n"
	@mkdir -p reports
	@printf "Go Licenses:\n" > reports/licenses.txt
	@printf "============\n" >> reports/licenses.txt
	@go-licenses csv ./... 2>/dev/null >> reports/licenses.txt || true
	@printf "\n\nnpm Licenses:\n" >> reports/licenses.txt
	@printf "=============\n" >> reports/licenses.txt
	@cd ui && npx license-checker --csv 2>/dev/null >> ../reports/licenses.txt || true
	@printf "$(GREEN)✓ License report: reports/licenses.txt$(RESET)\n"

# ============================================================================
# Help
# ============================================================================

help: ## Show this help
	@printf "$(BOLD)The Stem$(RESET) - Mustard Seed Networks\n"
	@printf "Network Performance Testing Tool\n"
	@printf "\n"
	@printf "$(BOLD)Usage:$(RESET) make [target]\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Build:$(RESET)\n"
	@printf "  all              Build everything (UI + Go binary)\n"
	@printf "  build            Build Go binary with symlink\n"
	@printf "  ui               Build React WebUI\n"
	@printf "  go               Build Go binary\n"
	@printf "  quick            Quick build (Go only, skip UI)\n"
	@printf "  clean            Clean all build artifacts\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Quality:$(RESET)\n"
	@printf "  verify           Full verification (lint + test + security + build)\n"
	@printf "  lint             Run all linters\n"
	@printf "  lint-go          Run Go linter (golangci-lint)\n"
	@printf "  lint-c           Run C linter (clang-tidy, Linux only)\n"
	@printf "  format           Format all code\n"
	@printf "  fix              Auto-fix linting issues\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Testing:$(RESET)\n"
	@printf "  test             Run Go unit tests\n"
	@printf "  test-coverage    Run tests with coverage report\n"
	@printf "  test-all         Run all tests (Go + C)\n"
	@printf "  c-test           Build and run C unit tests (Linux)\n"
	@printf "  smoke-test       Run C smoke tests (Linux, root)\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Security:$(RESET)\n"
	@printf "  security         Run all security scans\n"
	@printf "  security-backend Run Go security scans\n"
	@printf "  security-frontend Run npm audit\n"
	@printf "  security-secrets Scan for secrets (gitleaks)\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Licenses:$(RESET)\n"
	@printf "  license-check    Check dependency licenses\n"
	@printf "  license-check-go Check Go module licenses\n"
	@printf "  license-check-npm Check npm package licenses\n"
	@printf "  license-report   Generate full license report\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Development:$(RESET)\n"
	@printf "  dev              Show dev server instructions\n"
	@printf "  ui-dev           Run React dev server (port 3000)\n"
	@printf "  go-dev           Run Go backend (port 8080)\n"
	@printf "  tools            Install all development tools\n"
	@printf "  tools-go         Install Go development tools\n"
	@printf "  tools-frontend   Install frontend development tools\n"
	@printf "  update           Update all dependencies\n"
	@printf "  update-go        Update Go modules\n"
	@printf "  update-npm       Update npm packages\n"
	@printf "  version-check    Show version info and outdated packages\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Installation:$(RESET)\n"
	@printf "  install          Install to /usr/local/bin\n"
	@printf "  uninstall        Remove from /usr/local/bin\n"
	@printf "  install-service  Install as systemd service (Linux, root)\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Packaging:$(RESET)\n"
	@printf "  rpm              Build RPM package (Fedora/RHEL)\n"
	@printf "  deb              Build DEB package (Debian/Ubuntu)\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)C Dataplane:$(RESET)\n"
	@printf "  c-build          Build C dataplane (alias for dataplane)\n"
	@printf "  c-build-docker   Build C code using Docker (cross-platform)\n"
	@printf "  dataplane        Build C dataplane library (Linux only)\n"
	@printf "\n"
	@printf "$(BOLD)$(CYAN)Docker:$(RESET)\n"
	@printf "  build-linux-docker Build Linux binary via Docker\n"
