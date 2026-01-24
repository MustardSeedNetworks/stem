# =============================================================================
# Build Targets
# =============================================================================
#
# All build-related targets for The Stem:
#   - Frontend build (React/Vite)
#   - Backend build (Go binary)
#   - C dataplane (Linux only)
#   - Cross-compilation (Linux via Docker)
#   - Docker image builds
#
# =============================================================================

.PHONY: all ui ui-deps go quick clean \
        build build-darwin build-linux \
        c-build dataplane c-build-docker \
        build-minimal build-xdp build-dpdk \
        build-iperf3 build-iperf3-docker \
        build-linux-docker \
        ui-dev go-dev dev

# =============================================================================
# Main Build Targets
# =============================================================================

# Default: build everything
all: ui go ## Build everything (UI + Go binary)
	@echo ""
	@echo "Build complete!"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"

# Build binary (creates symlink for convenience)
build: go ## Build Go binary with symlink
	mkdir -p bin
	ln -sf $(notdir $(BINARY_NAME)) bin/stem 2>/dev/null || cp $(BINARY_NAME) bin/stem

# =============================================================================
# Frontend Build
# =============================================================================

ui-deps: ## Install UI dependencies
	@echo "Installing UI dependencies..."
	cd ui && npm install

ui: ui-deps ## Build React WebUI (output: internal/api/dist/)
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

quick: ## Quick build (Go only, assumes UI is already built)
	@echo "Quick build (Go only)..."
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/stem/

build-darwin: ui ## Build for macOS (native architecture)
	@echo "Building for macOS ($(shell uname -m))..."
	mkdir -p bin
	@CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o bin/stem-darwin ./cmd/stem/
	@echo "Built: bin/stem-darwin"

build-linux: ui ## Build for Linux AMD64 (CGO_ENABLED=0, no DPDK)
	@echo "Building for Linux AMD64 (pure Go)..."
	mkdir -p bin
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o bin/stem-linux ./cmd/stem/
	@echo "Built: bin/stem-linux"

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
	@echo "  Note: Network backends require Linux. Use 'make c-build-docker' for full build."
else
	@echo "Dataplane requires Linux or macOS"
endif

# Alias for dataplane target
c-build: dataplane ## Alias for dataplane

# =============================================================================
# C Dataplane Build Profiles (Linux only)
# =============================================================================

build-minimal: go ## AF_PACKET only (most compatible, no external deps)
ifeq ($(UNAME),Linux)
	@echo "Building C dataplane (AF_PACKET, minimal)..."
	mkdir -p bin
	$(CC) $(CFLAGS) -DAF_PACKET_MODE -o bin/stem-dataplane \
		$(C_COMMON_SRCS) $(C_PACKET_SRCS) $(C_LDFLAGS)
	@echo "Built: bin/stem-dataplane (AF_PACKET)"
else
	@echo "Dataplane requires Linux"
endif

build-xdp: go ## AF_XDP backend (good performance, needs libbpf)
ifeq ($(UNAME),Linux)
	@echo "Building C dataplane (AF_XDP)..."
	mkdir -p bin
	$(CC) $(CFLAGS) -DAF_XDP_MODE -o bin/stem-dataplane \
		$(C_COMMON_SRCS) $(C_XDP_SRCS) $(C_LDFLAGS) -lbpf -lxdp
	@echo "Built: bin/stem-dataplane (AF_XDP)"
else
	@echo "AF_XDP dataplane requires Linux"
endif

build-dpdk: go ## DPDK backend (max performance, needs DPDK SDK)
ifeq ($(UNAME),Linux)
	@echo "Building C dataplane (DPDK)..."
	mkdir -p bin
	$(CC) $(CFLAGS) $(shell pkg-config --cflags libdpdk 2>/dev/null) -DDPDK_MODE \
		-o bin/stem-dataplane \
		$(C_COMMON_SRCS) $(C_DPDK_SRCS) \
		$(C_LDFLAGS) $(shell pkg-config --libs libdpdk 2>/dev/null)
	@echo "Built: bin/stem-dataplane (DPDK)"
else
	@echo "DPDK dataplane requires Linux"
endif

# Build C code using Docker (cross-platform)
c-build-docker: ## Build C code using Docker (cross-platform)
	@echo "Building C dataplane in Docker..."
	docker run --rm -v $(PWD):/src -w /src gcc:latest make c-build

# =============================================================================
# iperf3 Bundling
# =============================================================================

IPERF3_VERSION ?= 3.18

build-iperf3: ## Build iperf3 from source for bundling
	@echo "Building iperf3 $(IPERF3_VERSION)..."
	@chmod +x scripts/build-iperf3.sh
	./scripts/build-iperf3.sh $(IPERF3_VERSION)

build-iperf3-docker: ## Build iperf3 in Docker (cross-platform)
	@echo "Building iperf3 $(IPERF3_VERSION) via Docker..."
	docker run --rm -v $(PWD):/workspace -w /workspace \
		ubuntu:24.04 bash -c "apt-get update -qq && apt-get install -y -qq build-essential curl > /dev/null 2>&1 && ./scripts/build-iperf3.sh $(IPERF3_VERSION)"

# =============================================================================
# Cross-Platform Builds
# =============================================================================

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
				-X $(VERSION_PKG).semver=$(VERSION) \
				-X $(VERSION_PKG).commit=$(COMMIT) \
				-X $(VERSION_PKG).buildTime=$(BUILD_TIME)" \
			-o bin/stem-linux-amd64 \
			./cmd/stem/
	@printf "$(GREEN)✓ Linux binary built: bin/stem-linux-amd64$(RESET)\n"

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

# =============================================================================
# Cleanup
# =============================================================================

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf ui/dist/
	rm -rf ui/node_modules/
	rm -rf internal/web/dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"
