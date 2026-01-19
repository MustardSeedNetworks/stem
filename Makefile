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
# Shared Infrastructure (version, colors, timers)
# =============================================================================

include Makefile.common

# =============================================================================
# Project Configuration
# =============================================================================

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
CFLAGS := -std=c23 -Wall -Wextra -Wpedantic -O3 -march=native -pthread -Iinclude
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

# =============================================================================
# Verification Pipeline
# =============================================================================

.PHONY: verify help version

verify: ## Full verification (lint, test, security, build)
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

# =============================================================================
# Version Information
# =============================================================================

version: ## Show version info
	@printf "$(BOLD)Stem Version Information$(RESET)\n"
	@printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	@printf "  Version:     $(VERSION)\n"
	@printf "  Commit:      $(COMMIT)\n"
	@printf "  Build Time:  $(BUILD_TIME)\n"
	@printf "  Platform:    $(PLATFORM_PRETTY) ($(PLATFORM)/$(GOARCH))\n"
	@printf "  Go:          $$(go version | awk '{print $$3}')\n"
	@printf "  Node:        $$(node --version 2>/dev/null || echo 'not installed')\n"

# =============================================================================
# Help
# =============================================================================

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
	@printf "$(BOLD)$(CYAN)Packaging:$(RESET)\n"
	@printf "  install-service  Install as systemd service (Linux, root)\n"
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
