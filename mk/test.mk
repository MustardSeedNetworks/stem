# =============================================================================
# Test Targets
# =============================================================================
#
# All testing targets:
#   - Go unit tests
#   - Frontend tests (Vitest)
#   - C unit tests (Linux only)
#   - E2E tests (Playwright)
#   - Coverage reports
#   - Smoke tests
#
# =============================================================================

.PHONY: test test-all test-backend test-backend-quiet test-frontend test-frontend-quiet \
        test-coverage test-coverage-html c-test c-test-asan c-fuzz smoke-test \
        test-e2e test-e2e-ui test-e2e-install

# =============================================================================
# Main Test Targets
# =============================================================================

test: ## Run unit tests (backend + frontend)
	@printf "$(BOLD)$(CYAN)┌─ Unit Tests ─────────────────────────────────────────────────────────────────┐$(RESET)\n"
	@printf "$(CYAN)│$(RESET) $(BOLD)[1/2]$(RESET) Backend (Go)                                                          $(CYAN)│$(RESET)\n"
	$(call timer-start,test-backend)
	@$(MAKE) --no-print-directory test-backend-quiet
	$(call timer-end,test-backend,Backend tests)
	@printf "$(CYAN)│$(RESET) $(BOLD)[2/2]$(RESET) Frontend (Vitest)                                                      $(CYAN)│$(RESET)\n"
	$(call timer-start,test-frontend)
	@$(MAKE) --no-print-directory test-frontend-quiet
	$(call timer-end,test-frontend,Frontend tests)
	@printf "$(CYAN)└──────────────────────────────────────────────────────────────────────────────┘$(RESET)\n"

test-all: test c-test test-e2e ## Run ALL tests (Go + C + E2E)
	@echo "All tests complete"

# =============================================================================
# Backend Tests
# =============================================================================

test-backend: ## Run Go tests with progress
	@printf "\n$(BOLD)🧪 Running backend tests...$(RESET)\n"
	@PKGS=$$(go list ./... | grep -v '/ui$$'); \
	PKG_COUNT=$$(echo "$$PKGS" | wc -l | tr -d ' '); \
	printf "   📦 Testing $$PKG_COUNT packages...\n\n"; \
	if command -v gotestsum > /dev/null 2>&1; then \
		gotestsum --format pkgname-and-test-fails -- -race -parallel 8 -coverprofile=coverage.out $$PKGS; \
	else \
		$(GO) test -v -race -parallel 8 -coverprofile=coverage.out $$PKGS; \
	fi
	@if [ -f coverage.out ]; then \
		COV=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
		printf "\n   📊 Coverage: %s\n" "$$COV"; \
	fi
	@printf "\n$(GREEN)✓ Backend tests complete$(RESET)\n"

test-backend-quiet:
	@PKGS=$$(go list ./... | grep -v '/ui$$'); \
	PKG_COUNT=$$(echo "$$PKGS" | wc -l | tr -d ' '); \
	printf "   Testing $$PKG_COUNT packages...\n"; \
	OUTPUT=$$($(GO) test -race -parallel 8 -coverprofile=coverage.out $$PKGS 2>&1); \
	STATUS=$$?; \
	echo "$$OUTPUT" | grep -E "^(ok|FAIL|---)"; \
	if [ $$STATUS -ne 0 ]; then \
		echo "$$OUTPUT"; \
		exit $$STATUS; \
	fi
	@if [ -f coverage.out ]; then \
		COV=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
		printf "   📊 Coverage: %s\n" "$$COV"; \
	fi

# =============================================================================
# Frontend Tests
# =============================================================================

test-frontend: ## Run frontend tests with progress
	@printf "\n$(BOLD)🧪 Running frontend tests...$(RESET)\n"
	@STORY_COUNT=$$(find ui/src -name "*.test.ts" -o -name "*.test.tsx" 2>/dev/null | wc -l | tr -d ' '); \
	printf "   📦 Running $$STORY_COUNT test files...\n\n"
	@cd ui && npm test
	@printf "\n$(GREEN)✓ Frontend tests complete$(RESET)\n"

test-frontend-quiet:
	@STORY_COUNT=$$(find ui/src -name "*.test.ts" -o -name "*.test.tsx" 2>/dev/null | wc -l | tr -d ' '); \
	printf "   Running $$STORY_COUNT test files...\n"
	@cd ui && OUTPUT=$$(npm test 2>&1); STATUS=$$?; \
	echo "$$OUTPUT" | grep -E "(PASS|FAIL|Tests:)"; \
	if [ $$STATUS -ne 0 ]; then \
		echo "$$OUTPUT"; \
		exit $$STATUS; \
	fi

# =============================================================================
# Coverage Reports
# =============================================================================

test-coverage: ## Run tests with coverage
	@echo "Running Go tests with coverage..."
	$(GO) test -v -race -parallel 8 -coverprofile=coverage.out -covermode=atomic ./internal/...
	$(GO) tool cover -func=coverage.out

test-coverage-html: test-coverage ## Generate HTML coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# =============================================================================
# C Tests (Linux only)
# =============================================================================

c-test: ## Build and run C unit tests
ifeq ($(UNAME),Linux)
	@echo "Building C tests..."
	mkdir -p bin
	$(CC) $(CFLAGS) -o bin/test_pacing tests/c/test_pacing.c $(C_PACING_SRCS) $(C_LDFLAGS)
	$(CC) $(CFLAGS) -o bin/test_protocols tests/c/test_protocols.c $(C_PROTO_SRCS) $(C_LDFLAGS)
	$(CC) $(CFLAGS) -o bin/test_packet_parse tests/c/test_packet_parse.c src/dataplane/common/packet.c $(C_LDFLAGS)
	@echo "Running C tests..."
	./bin/test_pacing
	./bin/test_protocols
	./bin/test_packet_parse
else ifeq ($(UNAME),Darwin)
	@echo "Building C tests (common code only, macOS)..."
	mkdir -p bin
	$(CC) $(CFLAGS) -DSTUB_PLATFORM -o bin/test_pacing tests/c/test_pacing.c $(C_PACING_SRCS) $(C_LDFLAGS)
	$(CC) $(CFLAGS) -DSTUB_PLATFORM -o bin/test_packet_parse tests/c/test_packet_parse.c src/dataplane/common/packet.c $(C_LDFLAGS)
	@echo "Running C tests..."
	./bin/test_pacing
	./bin/test_packet_parse
	@echo "Note: Protocol tests require Linux networking APIs"
else
	@echo "C tests require Linux or macOS"
endif

# Sanitizer flags for the safety targets: drop -O3/-march=native, add ASAN +
# UBSan and debug frames. The dataplane parser is the one place attacker bytes
# meet C, so it gets an explicit memory-safety gate.
C_SAN_CFLAGS := -D_GNU_SOURCE -D_DEFAULT_SOURCE -std=c23 -Wall -Wextra -Wpedantic \
                -g -O1 -fno-omit-frame-pointer -fsanitize=address,undefined -Iinclude
FUZZ_CC ?= clang
FUZZ_SECONDS ?= 60

c-test-asan: ## Build + run the dataplane parser tests under AddressSanitizer/UBSan
	@echo "Building packet-parser tests under ASAN/UBSan..."
	mkdir -p bin
	$(CC) $(C_SAN_CFLAGS) -o bin/test_packet_parse_asan \
		tests/c/test_packet_parse.c src/dataplane/common/packet.c -pthread -lm
	@echo "Running packet-parser tests (ASAN)..."
	ASAN_OPTIONS=detect_leaks=0 UBSAN_OPTIONS=halt_on_error=1 ./bin/test_packet_parse_asan

c-fuzz: ## Fuzz the dataplane packet parser under libFuzzer+ASAN (FUZZ_SECONDS=60)
	@command -v $(FUZZ_CC) >/dev/null 2>&1 || { echo "$(FUZZ_CC) (clang) required for libFuzzer"; exit 1; }
	@echo "Building libFuzzer harness..."
	mkdir -p bin
	$(FUZZ_CC) -D_GNU_SOURCE -std=c23 -g -O1 -fno-omit-frame-pointer \
		-fsanitize=fuzzer,address,undefined -Iinclude \
		-o bin/fuzz_packet tests/c/fuzz_packet.c src/dataplane/common/packet.c
	@echo "Fuzzing packet parser for $(FUZZ_SECONDS)s..."
	ASAN_OPTIONS=detect_leaks=0 ./bin/fuzz_packet -max_total_time=$(FUZZ_SECONDS) -print_final_stats=1

smoke-test: ## Run smoke tests (requires root, Linux only)
ifeq ($(UNAME),Linux)
	@echo "Running smoke tests..."
	sudo tests/smoke/run_smoke_tests.sh
else
	@echo "Smoke tests require Linux"
endif

# =============================================================================
# E2E Tests
# =============================================================================

test-e2e: ## Run frontend E2E tests (requires backend running)
	@echo ""
	@echo "🎭 Running E2E tests (Playwright)..."
	@E2E_COUNT=$$(find ui/e2e -name "*.spec.ts" 2>/dev/null | wc -l | tr -d ' '); \
	echo "   📦 Running $$E2E_COUNT spec files..."
	@echo ""
	@cd ui && npm run test:e2e
	@echo ""
	@echo "✅ E2E tests complete"

test-e2e-ui: ## Run E2E tests with Playwright UI
	@echo "🎭 Starting Playwright UI mode..."
	cd ui && npx playwright test --ui

test-e2e-install: ## Install Playwright browsers
	cd ui && npx playwright install --with-deps chromium
