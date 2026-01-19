# =============================================================================
# Linting & Formatting Targets
# =============================================================================
#
# Code quality and formatting:
#   - Go linting (golangci-lint v2)
#   - C linting (clang-tidy, Linux only)
#   - Formatting (gofmt, clang-format)
#   - Auto-fix capabilities
#
# =============================================================================

.PHONY: lint lint-go lint-c format format-go format-c fix

# =============================================================================
# Linting
# =============================================================================

lint: lint-go ## Run all linters
	@echo "✓ All linters passed"

lint-go: ## Run Go linter (golangci-lint)
	@echo "Running Go linter (golangci-lint)..."
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi; \
	$$GOLANGCI_LINT run --allow-parallel-runners ./...
	@echo "✓ Go lint passed"

lint-c: ## Run C linter (clang-tidy, Linux only)
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

# =============================================================================
# Formatting
# =============================================================================

format: format-go ## Format all code
	@echo "✓ All code formatted"

format-go: ## Format Go code
	@echo "Formatting Go code..."
	@gofmt -w -s .
	@echo "✓ Go code formatted"

format-c: ## Format C code (Linux only)
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

# =============================================================================
# Auto-Fix
# =============================================================================

fix: ## Auto-fix linting issues
	@echo "Auto-fixing Go code..."
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi; \
	$$GOLANGCI_LINT run --fix ./...
	@gofmt -w -s .
	@echo "✓ Auto-fix complete"
