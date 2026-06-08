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
#   Release artifacts:  built by GitHub Actions on tag/release
#
# REQUIREMENTS
# ------------
#   - Go 1.25+ (with CGO for certain features)
#   - Node.js 25.2.1+ and npm
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
VERSION_PKG := github.com/MustardSeedNetworks/stem/internal/version

# Embedded UI assets — Vite outputs here directly; Go //go:embed reads from here.
EMBED_DIR := internal/api/ui

# UI build hash for local build verification (md5 of all embedded assets).
# Mirrors the canonical computation in niac/go and seed.
UI_BUILD_HASH := $(shell if [ -d "$(EMBED_DIR)" ] && [ -n "$$(ls -A $(EMBED_DIR) 2>/dev/null)" ]; then \
	find $(EMBED_DIR) -type f -exec md5 -q {} \; 2>/dev/null | sort | md5 -q 2>/dev/null || \
	find $(EMBED_DIR) -type f -exec md5sum {} \; 2>/dev/null | sort | md5sum 2>/dev/null | cut -d' ' -f1; \
else echo ""; fi)

# Canonical ldflags contract shared with seed and niac:
# internal/version.{Version,Commit,BuildTime,UIBuildHash} (PascalCase).
LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).BuildTime=$(BUILD_TIME) \
	-X $(VERSION_PKG).UIBuildHash=$(UI_BUILD_HASH)
GOFLAGS := -trimpath -buildvcs=false -ldflags "$(LDFLAGS)"

# Local build output. GitHub Actions owns platform-suffixed release artifacts.
BINARY := stem
BINARY_NAME := bin/$(BINARY)

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
#   Release: GitHub Actions builds release artifacts on the required platforms
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
include mk/dev.mk

# =============================================================================
# Default Target
# =============================================================================

all: verify ## Full local verification

# =============================================================================
# Cleanup
# =============================================================================

.PHONY: clean clean-all

clean: ## Clean build artifacts
	rm -f $(BINARY) $(BINARY)-*
	rm -f bin/$(BINARY) bin/$(BINARY)-*
	rm -f coverage.out coverage.html
	find internal/api/ui -mindepth 1 ! -name .gitkeep -exec rm -rf {} +
	rm -rf bin/test_*
	find src -name '*.o' -delete

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
	@printf "$(GREEN)Local verification complete. GitHub Actions owns release artifacts.$(RESET)\n\n"

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
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) 2>/dev/null | sort -u | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build                    Build current-host binary"
	@echo "  make verify                   Full local verification"
	@echo "  make dataplane                Build current-host C dataplane"
