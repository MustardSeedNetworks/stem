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

# Terminal colors (ANSI escape codes)
BOLD := \033[1m
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RESET := \033[0m

# Step display helper
# Usage: $(call step,current,total,description)
define step
	@printf "\n$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BOLD)$(CYAN)[$(1)/$(2)]$(RESET) $(BOLD)$(3)$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
endef

# Success message helper
# Usage: $(call success,message)
define success
	@printf "$(GREEN)✓ $(1)$(RESET)\n"
endef

# Warning message helper
# Usage: $(call warn,message)
define warn
	@printf "$(YELLOW)⚠ $(1)$(RESET)\n"
endef

# Error message helper
# Usage: $(call error,message)
define error
	@printf "$(RED)✗ $(1)$(RESET)\n"
endef

# Timer: Start a named timer
# Usage: $(call timer-start,name)
define timer-start
	@date +%s > /tmp/make-timer-$(1)
endef

# Timer: Show elapsed time for a named timer
# Usage: $(call timer-end,name,description)
define timer-end
	@if [ -f /tmp/make-timer-$(1) ]; then \
		START=$$(cat /tmp/make-timer-$(1)); \
		END=$$(date +%s); \
		ELAPSED=$$((END - START)); \
		MINS=$$((ELAPSED / 60)); \
		SECS=$$((ELAPSED % 60)); \
		if [ $$MINS -gt 0 ]; then \
			printf "$(GREEN)✓ $(2) $(YELLOW)($$MINS min $$SECS sec)$(RESET)\n"; \
		else \
			printf "$(GREEN)✓ $(2) $(YELLOW)($$SECS sec)$(RESET)\n"; \
		fi; \
		rm -f /tmp/make-timer-$(1); \
	fi
endef

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

.PHONY: all ui ui-deps go clean test dev install help lint lint-go lint-c format format-go format-c fix verify build-linux-docker deploy smoke-test-remote logs logs-100 remote-status update update-go update-npm version-check tools tools-go tools-frontend security security-backend security-frontend security-secrets license-check license-check-go license-check-npm license-report

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
# Remote Deployment
# ============================================================================

# Deployment configuration (override with environment variables)
DEPLOY_HOST ?= 192.168.64.7
DEPLOY_USER ?= krisarmstrong
DEPLOY_PATH ?= /home/$(DEPLOY_USER)/stem
DEPLOY_PORT ?= 8443

# Build Linux binary via Docker (for cross-compilation from macOS)
build-linux-docker:
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

# Full deployment pipeline
deploy: ## Deploy to remote Ubuntu server
	@printf "\n$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  DEPLOYING TO $(DEPLOY_HOST)$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n\n"
	$(call timer-start,deploy)
	@printf "$(BOLD)[1/5]$(RESET) Building Linux binary...\n"
ifeq ($(UNAME),Darwin)
	@$(MAKE) build-linux-docker
else
	@$(MAKE) go BINARY_NAME=bin/stem-linux-amd64
endif
	@printf "$(BOLD)[2/5]$(RESET) Syncing to remote server...\n"
	rsync -avz --progress bin/stem-linux-amd64 $(DEPLOY_USER)@$(DEPLOY_HOST):$(DEPLOY_PATH)/stem-new
	@printf "$(BOLD)[3/5]$(RESET) Installing on remote server...\n"
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) "\
		cd $(DEPLOY_PATH) && \
		sudo systemctl stop stem 2>/dev/null || true && \
		mv stem-new stem && \
		sudo setcap cap_net_raw=+ep ./stem 2>/dev/null || true && \
		sudo systemctl start stem"
	@printf "$(BOLD)[4/5]$(RESET) Waiting for service startup...\n"
	@sleep 3
	@printf "$(BOLD)[5/5]$(RESET) Running smoke tests...\n"
	@$(MAKE) smoke-test-remote
	$(call timer-end,deploy,Deployment)
	@printf "\n$(GREEN)✓ DEPLOYMENT SUCCESSFUL$(RESET)\n"
	@printf "  URL: https://$(DEPLOY_HOST):$(DEPLOY_PORT)\n\n"

# Remote smoke tests - verify deployment
smoke-test-remote: ## Run smoke tests against deployed server
	@printf "$(BOLD)$(CYAN)Running remote smoke tests...$(RESET)\n"
	@ERRORS=0; \
	\
	printf "  [1/4] Checking process...\n"; \
	if ssh $(DEPLOY_USER)@$(DEPLOY_HOST) "pgrep -f 'stem' > /dev/null"; then \
		printf "$(GREEN)    ✓ Process running$(RESET)\n"; \
	else \
		printf "$(RED)    ✗ Process not running$(RESET)\n"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	\
	printf "  [2/4] Checking API health...\n"; \
	if curl -sf -k "https://$(DEPLOY_HOST):$(DEPLOY_PORT)/api/health" > /dev/null 2>&1; then \
		printf "$(GREEN)    ✓ API responding$(RESET)\n"; \
	else \
		printf "$(RED)    ✗ API not responding$(RESET)\n"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	\
	printf "  [3/4] Checking version endpoint...\n"; \
	VERSION_RESP=$$(curl -sf -k "https://$(DEPLOY_HOST):$(DEPLOY_PORT)/api/version" 2>/dev/null); \
	if [ -n "$$VERSION_RESP" ]; then \
		printf "$(GREEN)    ✓ Version: $$VERSION_RESP$(RESET)\n"; \
	else \
		printf "$(YELLOW)    ⚠ Version endpoint not available$(RESET)\n"; \
	fi; \
	\
	printf "  [4/4] Checking WebUI...\n"; \
	if curl -sf -k "https://$(DEPLOY_HOST):$(DEPLOY_PORT)/" 2>/dev/null | grep -q "<!DOCTYPE"; then \
		printf "$(GREEN)    ✓ WebUI loading$(RESET)\n"; \
	else \
		printf "$(RED)    ✗ WebUI not loading$(RESET)\n"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	\
	if [ $$ERRORS -gt 0 ]; then \
		printf "\n$(RED)✗ Smoke tests failed ($$ERRORS errors)$(RESET)\n"; \
		exit 1; \
	fi; \
	printf "\n$(GREEN)✓ All smoke tests passed$(RESET)\n"

# Remote log access
logs: ## Stream live logs from remote server
	@printf "$(BOLD)$(CYAN)Streaming logs from $(DEPLOY_HOST)...$(RESET)\n"
	@printf "$(YELLOW)Press Ctrl+C to stop$(RESET)\n\n"
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) "sudo journalctl -u stem -f"

logs-100: ## Show last 100 log lines from remote server
	@printf "$(BOLD)$(CYAN)Last 100 log lines from $(DEPLOY_HOST):$(RESET)\n\n"
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) "sudo journalctl -u stem -n 100 --no-pager"

remote-status: ## Check service status on remote server
	@printf "$(BOLD)$(CYAN)Service status on $(DEPLOY_HOST):$(RESET)\n\n"
	ssh $(DEPLOY_USER)@$(DEPLOY_HOST) "\
		echo '=== Systemd Status ===' && \
		sudo systemctl status stem --no-pager && \
		echo '' && \
		echo '=== Process Info ===' && \
		ps aux | grep -E 'PID|stem' | grep -v grep && \
		echo '' && \
		echo '=== Listening Ports ===' && \
		sudo ss -tlnp | grep -E 'State|stem' || true"

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
	@printf "$(BOLD)$(CYAN)Deployment:$(RESET)\n"
	@printf "  deploy           Deploy to remote Ubuntu server\n"
	@printf "  smoke-test-remote Run smoke tests against deployed server\n"
	@printf "  logs             Stream live logs from remote server\n"
	@printf "  logs-100         Show last 100 log lines from remote\n"
	@printf "  remote-status    Check service status on remote\n"
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
	@printf "$(BOLD)$(CYAN)C Dataplane:$(RESET) (Linux only)\n"
	@printf "  dataplane        Build C dataplane library\n"
	@printf "\n"
	@printf "$(YELLOW)Environment Variables:$(RESET)\n"
	@printf "  DEPLOY_HOST      Target server IP (default: 192.168.64.7)\n"
	@printf "  DEPLOY_USER      SSH user (default: krisarmstrong)\n"
	@printf "  DEPLOY_PATH      Remote path (default: /home/USER/stem)\n"
	@printf "  DEPLOY_PORT      HTTPS port (default: 8443)\n"
