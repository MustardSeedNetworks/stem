# =============================================================================
# Stem Makefile
# =============================================================================
#
# Build, test, and package automation for Stem network performance testing tool.
#
# QUICK START
# -----------
#   make build          Build binary (UI + Go)
#   make test           Run all tests
#   make verify         Full CI pipeline (lint, test, security, build)
#   make dev            Development mode instructions
#   make help           Show all available targets
#
# COMMON WORKFLOWS
# ----------------
#   Development:        make dev (then run ui-dev and go-dev in separate terminals)
#   Before commit:      make verify
#   Release:            make release VERSION=v1.0.0
#   Cross-compile:      make build-linux-docker
#   Package:            make packages (deb + rpm)
#
# REQUIREMENTS
# ------------
#   - Go 1.25+ (with CGO for certain features)
#   - Node.js 25.2.1+ and npm
#   - Docker (optional, for cross-compilation)
#   - Linux (for C dataplane builds)
#
# =============================================================================

# =============================================================================
# Shared Infrastructure (version, platform, colors)
# =============================================================================

include mk/vars.mk

# =============================================================================
# Display Helpers
# =============================================================================

# Print a section header
define section
	@printf "\n$(BOLD)$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  $(1)$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n\n"
endef

# Print a step in a multi-step process
define step
	@printf "$(BOLD)[$(1)/$(2)]$(RESET) $(3)\n"
endef

# Print a success message
define success
	@printf "$(GREEN)✓ $(1)$(RESET)\n"
endef

# Print a warning message
define warn
	@printf "$(YELLOW)⚠ $(1)$(RESET)\n"
endef

# Print an error message
define error
	@printf "$(RED)✗ $(1)$(RESET)\n"
endef

# =============================================================================
# Timer Functions
# =============================================================================

# Start a named timer
define timer-start
	@date +%s > /tmp/make-timer-$(1)
endef

# End a timer and display elapsed time
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

# =============================================================================
# Project Configuration
# =============================================================================

# Go settings
GO := go
VERSION_PKG := github.com/krisarmstrong/stem/internal/version
LDFLAGS := -s -w \
	-X $(VERSION_PKG).semver=$(VERSION) \
	-X $(VERSION_PKG).commit=$(COMMIT) \
	-X $(VERSION_PKG).buildTime=$(BUILD_TIME)
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

# Docker settings
DOCKER_IMAGE?=stem
DOCKER_TAG?=$(VERSION)

# =============================================================================
# C/Dataplane Configuration
# =============================================================================
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
# =============================================================================

# C compiler settings - C23 standard
CC := gcc
CFLAGS := -D_GNU_SOURCE -D_DEFAULT_SOURCE -std=c23 -Wall -Wextra -Wpedantic -O3 -march=native -pthread -Iinclude
C_LDFLAGS := -pthread -lm

# C sources - both dataplane and reflector (excluding main.c)
C_DATAPLANE_SRCS := $(wildcard src/dataplane/common/*.c)
C_REFLECTOR_SRCS := $(filter-out src/reflector/main.c,$(wildcard src/reflector/*.c))
C_ALL_SRCS := $(C_DATAPLANE_SRCS) $(C_REFLECTOR_SRCS)
C_ALL_OBJS := $(C_ALL_SRCS:.c=.o)

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

# =============================================================================
# Include Domain-Specific Makefiles
# =============================================================================

include mk/build.mk
include mk/test.mk
include mk/lint.mk
include mk/security.mk
include mk/deps.mk
include mk/package.mk
include mk/container.mk

# =============================================================================
# Default Target
# =============================================================================

all: verify ## Full build and validation (recommended before release)

# =============================================================================
# Cleanup
# =============================================================================

.PHONY: clean clean-all

clean: ## Clean build artifacts
	rm -f $(BINARY) $(BINARY)-*
	rm -f bin/$(BINARY) bin/$(BINARY)-*
	rm -f coverage.out coverage.html
	rm -rf ui/dist
	rm -rf internal/api/embed/*
	rm -f bin/test_*
	rm -f src/**/*.o

clean-all: clean ## Clean everything including dependencies
	rm -rf ui/node_modules
	rm -rf dist/
	rm -rf bin/
	rm -rf reports/

# =============================================================================
# Verification Pipeline
# =============================================================================

.PHONY: verify pre-commit pre-commit-install

verify: ## Full verification (lint, test, security, build)
	@printf "\n$(BOLD)$(CYAN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BOLD)$(CYAN)║                        FULL VERIFICATION PIPELINE                           ║$(RESET)\n"
	@printf "$(BOLD)$(CYAN)║                        Version: $(VERSION)$(RESET)\n"
	@printf "$(BOLD)$(CYAN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)\n"
	$(call timer-start,verify-total)
	$(call step,1,5,Linting Code)
	$(call timer-start,lint)
	@$(MAKE) --no-print-directory lint
	$(call timer-end,lint,Linting)
	$(call step,2,5,Running Tests)
	$(call timer-start,test)
	@$(MAKE) --no-print-directory test
	$(call timer-end,test,Tests)
	$(call step,3,5,Security Scanning)
	$(call timer-start,security)
	@$(MAKE) --no-print-directory security
	$(call timer-end,security,Security)
	$(call step,4,5,Building Application)
	$(call timer-start,build)
	@$(MAKE) --no-print-directory build
	$(call timer-end,build,Build)
	$(call step,5,5,License Check)
	@$(MAKE) --no-print-directory license-check || true
	@printf "\n$(BOLD)$(GREEN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BOLD)$(GREEN)║                        ✓ VERIFICATION COMPLETE                               ║$(RESET)\n"
	@printf "$(BOLD)$(GREEN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)\n"
	$(call timer-end,verify-total,Total verification)
	@printf "\n  $(BOLD)Version:$(RESET)     $(VERSION)\n"
	@printf "  $(BOLD)Commit:$(RESET)      $(COMMIT)\n"
	@printf "  $(BOLD)Binary:$(RESET)      $(BINARY_NAME)\n\n"
	@printf "$(GREEN)Ready for release! Run 'make packages' to build packages.$(RESET)\n\n"

pre-commit: ## Run pre-commit hooks manually
	@if command -v pre-commit > /dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "pre-commit not installed. Install with: pip install pre-commit"; \
		exit 1; \
	fi

pre-commit-install: ## Install pre-commit hooks
	@if command -v pre-commit > /dev/null 2>&1; then \
		pre-commit install; \
		pre-commit install --hook-type pre-push; \
		echo "Pre-commit hooks installed successfully"; \
	else \
		echo "pre-commit not installed. Install with: pip install pre-commit"; \
		exit 1; \
	fi

# =============================================================================
# Release Automation
# =============================================================================

.PHONY: release release-check

release: ## Create release (VERSION=vX.Y.Z required)
	@if [ -z "$(VERSION)" ] || ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then \
		echo "ERROR: VERSION must be set in format vX.Y.Z"; \
		echo "Usage: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "ERROR: Working directory not clean. Commit or stash changes first."; \
		exit 1; \
	fi
	@printf "$(BOLD)$(CYAN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BOLD)$(CYAN)║                         RELEASE $(VERSION)                                   ║$(RESET)\n"
	@printf "$(BOLD)$(CYAN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)\n"
	@echo ""
	@echo "Step 1/5: Running verification..."
	@$(MAKE) --no-print-directory verify
	@echo ""
	@echo "Step 2/5: Building all platforms..."
	@$(MAKE) --no-print-directory build-all
	@echo ""
	@echo "Step 3/5: Creating git tag..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tagged: $(VERSION)"
	@echo ""
	@echo "Step 4/5: Pushing tag to origin..."
	@git push origin $(VERSION)
	@echo ""
	@echo "Step 5/5: Creating GitHub release..."
	@if command -v gh > /dev/null 2>&1; then \
		gh release create $(VERSION) \
			--title "$(VERSION)" \
			--generate-notes \
			bin/$(BINARY)-darwin-* bin/$(BINARY)-linux-* 2>/dev/null || \
			echo "Note: Upload binaries manually if gh release failed"; \
	else \
		echo "GitHub CLI not installed. Create release manually at:"; \
		echo "https://github.com/krisarmstrong/stem/releases/new?tag=$(VERSION)"; \
	fi
	@echo ""
	@printf "$(GREEN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(GREEN)║                     ✓ RELEASE $(VERSION) COMPLETE                            ║$(RESET)\n"
	@printf "$(GREEN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)\n"
	@echo ""
	@echo "Binaries:"
	@ls -lh bin/$(BINARY)-* 2>/dev/null || true

release-check: verify ## Full release validation (verify + install script check)
	@echo "Running release checks..."
	@bash -n deploy/systemd/install.sh 2>/dev/null && echo "PASS: install.sh syntax valid" || true
	@bash -n deploy/systemd/uninstall.sh 2>/dev/null && echo "PASS: uninstall.sh syntax valid" || true
	@if echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then \
		echo "PASS: Version format valid ($(VERSION))"; \
	else \
		echo "WARN: Version $(VERSION) doesn't match vX.Y.Z format"; \
	fi
	@if git diff --quiet && git diff --staged --quiet; then \
		echo "PASS: No uncommitted changes"; \
	else \
		echo "WARN: Uncommitted changes detected"; \
	fi
	@echo ""
	@echo "Release checks complete. Ready for 'git tag $(VERSION) && git push --tags'"

# =============================================================================
# Version Information
# =============================================================================

.PHONY: version

version: ## Show version info
	@printf "$(BOLD)Stem Version Information$(RESET)\n"
	@printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	@printf "  Version:     $(VERSION)\n"
	@printf "  Commit:      $(COMMIT)\n"
	@printf "  Build Time:  $(BUILD_TIME)\n"
	@printf "  Platform:    $(PLATFORM) ($(GOARCH))\n"
	@printf "  Go:          $$(go version | awk '{print $$3}')\n"
	@printf "  Node:        $$(node --version 2>/dev/null || echo 'not installed')\n"
	@if [ -f "./bin/$(BINARY)" ]; then \
		printf "\n$(BOLD)Binary:$(RESET)\n"; \
		ls -lh ./bin/$(BINARY); \
	fi

# =============================================================================
# Help
# =============================================================================

.PHONY: help

help: ## Show this help
	@echo "The Stem - Network Performance Testing Tool by Mustard Seed Networks"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) mk/*.mk 2>/dev/null | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build                    Build production binary"
	@echo "  make verify                   Full CI pipeline"
	@echo "  make release VERSION=v1.0.0   Create tagged release"
	@echo "  make packages                 Build deb and rpm packages"
	@echo "  make build-linux-docker       Build Linux binary via Docker"
