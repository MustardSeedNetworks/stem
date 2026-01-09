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

# Version (injected via ldflags at build time)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

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

# Platform detection
UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
    BINARY_NAME := bin/$(BINARY)-darwin
else ifeq ($(UNAME),Linux)
    BINARY_NAME := bin/$(BINARY)-linux
else
    BINARY_NAME := bin/$(BINARY)
endif

# ============================================================================
# Main Targets
# ============================================================================

.PHONY: all ui ui-deps go clean test dev install help lint lint-go lint-c format format-go format-c fix verify

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

# Full verification: lint, test, build
verify: lint test build
	@echo ""
	@echo "════════════════════════════════════════════════════════════════════════════"
	@echo "  ✓ VERIFICATION COMPLETE"
	@echo "    - Linters passed"
	@echo "    - Tests passed"
	@echo "    - Build successful: $(BINARY_NAME)"
	@echo "    - Version: $(VERSION)"
	@echo "════════════════════════════════════════════════════════════════════════════"

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
# C Dataplane (Linux only)
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
# Packaging
# ============================================================================

# Build RPM package (Fedora/RHEL/CentOS)
rpm: build
	@echo "Building RPM package..."
	@if ! command -v rpmbuild >/dev/null 2>&1; then \
		echo "Error: rpmbuild not found. Install rpm-build package."; \
		exit 1; \
	fi
	@mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	@cd .. && tar czf ~/rpmbuild/SOURCES/stem-$(VERSION).tar.gz \
		--transform 's,^stem,stem-$(VERSION),' \
		--exclude='stem/.git' --exclude='stem/node_modules' --exclude='stem/bin' \
		--exclude='stem/ui/node_modules' --exclude='stem/build' \
		stem
	@sed 's/Version:.*/Version:        $(VERSION)/' packaging/rpm/stem.spec > ~/rpmbuild/SPECS/stem.spec
	@rpmbuild -bb ~/rpmbuild/SPECS/stem.spec
	@echo "RPM built: ~/rpmbuild/RPMS/x86_64/stem-$(VERSION)*.rpm"

# Build DEB package (Debian/Ubuntu)
deb: build
	@echo "Building DEB package..."
	@if ! command -v dpkg-buildpackage >/dev/null 2>&1; then \
		echo "Error: dpkg-buildpackage not found. Install devscripts package."; \
		exit 1; \
	fi
	@mkdir -p debian
	@cp -r packaging/debian/* debian/
	@dpkg-buildpackage -us -uc -b
	@echo "DEB built: ../stem_$(VERSION)*.deb"

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
# Help
# ============================================================================

help:
	@echo "The Stem - Mustard Seed Networks"
	@echo "Copyright (c) 2025 Mustard Seed Networks"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Main targets:"
	@echo "  all            Build everything (UI + Go binary)"
	@echo "  build          Build Go binary with symlink"
	@echo "  ui             Build React WebUI"
	@echo "  go             Build Go binary"
	@echo "  quick          Quick build (Go only, skip UI)"
	@echo "  clean          Clean all build artifacts"
	@echo "  verify         Full verification (lint + test + build)"
	@echo ""
	@echo "Linting & Formatting:"
	@echo "  lint           Run all linters"
	@echo "  lint-go        Run Go linter (golangci-lint)"
	@echo "  lint-c         Run C linter (clang-tidy, Linux)"
	@echo "  format         Format all code"
	@echo "  format-go      Format Go code (gofmt)"
	@echo "  format-c       Format C code (clang-format, Linux)"
	@echo "  fix            Auto-fix linting issues"
	@echo ""
	@echo "Development:"
	@echo "  ui-dev         Run React dev server (port 3000)"
	@echo "  go-dev         Run Go backend (port 8080)"
	@echo "  dev            Show dev instructions"
	@echo ""
	@echo "Testing:"
	@echo "  test           Run Go unit tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  c-test         Build and run C unit tests (Linux)"
	@echo "  smoke-test     Run smoke tests with veth (Linux, root)"
	@echo "  test-all       Run all tests (Go + C)"
	@echo ""
	@echo "Installation:"
	@echo "  install        Install to /usr/local/bin"
	@echo "  uninstall      Remove from /usr/local/bin"
	@echo "  install-service Install as systemd service (Linux, root)"
	@echo ""
	@echo "Packaging:"
	@echo "  rpm            Build RPM package (Fedora/RHEL)"
	@echo "  deb            Build DEB package (Debian/Ubuntu)"
	@echo ""
	@echo "C Dataplane (Linux):"
	@echo "  dataplane      Build C dataplane library"
