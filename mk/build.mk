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
        build c-build dataplane c-build-docker \
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

ui: ui-deps ## Build React WebUI
	@echo "Building React WebUI..."
	cd ui && npm run build
	@echo "Copying dist to internal/web for embedding..."
	mkdir -p internal/web/dist
	cp -r ui/dist/* internal/web/dist/

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

# =============================================================================
# C Dataplane Build (Linux only)
# =============================================================================

# Build C dataplane library (Linux only)
dataplane: ## Build C dataplane + reflector library (Linux only)
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

# Alias for dataplane target
c-build: dataplane ## Alias for dataplane

# Build C code using Docker (cross-platform)
c-build-docker: ## Build C code using Docker (cross-platform)
	@echo "Building C dataplane in Docker..."
	docker run --rm -v $(PWD):/src -w /src gcc:latest make c-build

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
				-X $(VERSION_PKG).Version=$(VERSION) \
				-X $(VERSION_PKG).Commit=$(COMMIT) \
				-X $(VERSION_PKG).BuildTime=$(BUILD_TIME)" \
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
