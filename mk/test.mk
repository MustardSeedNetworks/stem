# =============================================================================
# Test Targets
# =============================================================================
#
# All testing targets:
#   - Go unit tests
#   - C unit tests (Linux only)
#   - Coverage reports
#   - Smoke tests
#
# =============================================================================

.PHONY: test test-coverage test-coverage-html test-all c-test smoke-test

# =============================================================================
# Go Tests
# =============================================================================

test: ## Run Go tests
	@echo "Running Go tests..."
	$(GO) test -v -race ./internal/... ./cmd/...

test-coverage: ## Run tests with coverage
	@echo "Running Go tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/...
	$(GO) tool cover -func=coverage.out

test-coverage-html: test-coverage ## Generate HTML coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# =============================================================================
# C Tests (Linux only)
# =============================================================================

c-test: ## Build and run C unit tests (Linux only)
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

smoke-test: ## Run smoke tests (requires root, Linux only)
ifeq ($(UNAME),Linux)
	@echo "Running smoke tests..."
	sudo tests/smoke/run_smoke_tests.sh
else
	@echo "Smoke tests require Linux"
endif

# =============================================================================
# Combined Test Targets
# =============================================================================

test-all: test c-test ## Run all tests (Go + C)
	@echo "All tests complete"
