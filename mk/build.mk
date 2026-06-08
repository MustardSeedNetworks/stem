# =============================================================================
# Build Targets
# =============================================================================
#
# Local build targets for The Stem:
#   - Frontend build (React/Vite)
#   - Backend build for the current host
#   - C dataplane for the current host
#
# GitHub Actions owns cross-platform artifacts, provenance, and checksums.
#
# =============================================================================

.PHONY: ui ui-deps go quick schema \
        build c-build dataplane \
        ui-dev go-dev dev

schema: ## Regenerate docs/schemas/api/*.json from internal/api Go DTOs
	@echo "Generating JSON Schemas for API DTOs..."
	@go run ./cmd/stem-schema -o docs/schemas/api
	@echo "Wrote $$(ls -1 docs/schemas/api/*.json 2>/dev/null | wc -l | tr -d ' ') schema(s) to docs/schemas/api/"

# =============================================================================
# Main Build Targets
# =============================================================================

build: ui go ## Build current host binary with embedded UI
	@echo "Build complete: $(BINARY_NAME)"

# =============================================================================
# Frontend Build
# =============================================================================

ui-deps: ## Install UI dependencies
	@if [ ! -d ui/node_modules ] || [ ui/package-lock.json -nt ui/node_modules/.package-lock.json ]; then \
		echo "Installing UI dependencies..."; \
		cd ui && npm ci; \
	else \
		echo "UI dependencies up to date"; \
	fi

ui: ui-deps ## Build React WebUI (output: internal/api/ui/)
	@echo "Building React WebUI..."
	cd ui && npm run build

# =============================================================================
# Backend Build
# =============================================================================

go: ## Build Go binary
	@echo "Building $(BINARY)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/stem/
	@echo "Built: $(BINARY_NAME)"

quick: ## Quick current-host Go rebuild (assumes UI is already built)
	@echo "Quick build (Go only)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/stem/

# =============================================================================
# C Dataplane Build
# =============================================================================

# Source groupings for build profiles
C_COMMON_SRCS := $(wildcard src/dataplane/common/*.c)
C_PACKET_SRCS := $(wildcard src/dataplane/linux_packet/*.c)
C_XDP_SRCS    := $(wildcard src/dataplane/linux_xdp/*.c)
C_DPDK_SRCS   := $(wildcard src/dataplane/linux_dpdk/*.c)

# Build C dataplane library (default: AF_PACKET on Linux, common libs on macOS)
dataplane: ## Build C dataplane (Linux: AF_PACKET, macOS: common libs only)
ifeq ($(UNAME),Linux)
	@echo "Building C dataplane + reflector library (AF_PACKET)..."
	@for src in $(C_ALL_SRCS); do \
		$(CC) $(CFLAGS) -c $$src -o $${src%.c}.o; \
	done
	mkdir -p build
	ar rcs build/libreflector.a $(C_ALL_OBJS)
	cp build/libreflector.a librfc2544.a
	@echo "Built: build/libreflector.a"
else ifeq ($(UNAME),Darwin)
	@echo "Building C common libraries (macOS, no network backends)..."
	mkdir -p build
	$(eval C_MACOS_SRCS := $(filter-out src/dataplane/common/nic_detect.c src/dataplane/common/packet.c src/dataplane/common/core.c src/dataplane/common/main.c,$(C_COMMON_SRCS)))
	@for src in $(C_MACOS_SRCS); do \
		$(CC) $(CFLAGS) -DSTUB_PLATFORM -c $$src -o $${src%.c}.o; \
	done
	ar rcs build/libstem-common.a $(C_MACOS_SRCS:.c=.o)
	@rm -f src/dataplane/common/*.o
	@echo "Built: build/libstem-common.a (common code only)"
	@echo "  Note: Network backends require Linux and are built by GitHub Actions for release artifacts."
else
	@echo "Dataplane requires Linux or macOS"
endif

# Alias for dataplane target
c-build: dataplane ## Alias for dataplane

# =============================================================================
# Development Targets
# =============================================================================

ui-dev: ## Run UI dev server
	cd ui && npm run dev

go-dev: ## Run Go backend
	$(GO) run ./cmd/stem/ web -p 8080

dev: ## Development mode (show instructions)
	@echo "Starting development servers..."
	@echo "UI: http://localhost:3000"
	@echo "API: http://localhost:8080"
	@echo ""
	@echo "Run in separate terminals:"
	@echo "  make ui-dev    # React dev server"
	@echo "  make go-dev    # Go backend"
