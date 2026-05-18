# =============================================================================
# Linting & Formatting Targets
# =============================================================================
#
# Code quality and formatting:
#   - Go linting (golangci-lint v2)
#   - C linting (clang-tidy, Linux only)
#   - Frontend linting (Biome)
#   - Markdown linting
#   - Formatting (gofmt, clang-format, Biome)
#   - Auto-fix capabilities
#
# =============================================================================

.PHONY: lint lint-go lint-c lint-frontend lint-frontend-quiet lint-md \
        fmt fmt-go fmt-c fmt-frontend fix fix-all

# =============================================================================
# Linting
# =============================================================================

lint: lint-go lint-frontend ## Run all linters
	@printf "$(GREEN)✓ All linters passed$(RESET)\n"

lint-go: ## Run Go linter (golangci-lint)
	@printf "$(BOLD)🔍 Running Go linter (golangci-lint)...$(RESET)\n"
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		printf "📦 Installing golangci-lint v2...\n"; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.1; \
	fi; \
	$$GOLANGCI_LINT run --allow-parallel-runners ./...
	@printf "$(GREEN)✓ Go lint passed$(RESET)\n"

lint-frontend: ## Run frontend linter (Biome)
	@printf "$(BOLD)🔍 Running frontend linter (Biome)...$(RESET)\n"
	@cd ui && npx @biomejs/biome check src/
	@printf "$(GREEN)✓ Frontend lint complete$(RESET)\n"

lint-frontend-quiet:
	@FILE_COUNT=$$(find ui/src -name "*.ts" -o -name "*.tsx" 2>/dev/null | wc -l | tr -d ' '); \
	printf "   Checking $$FILE_COUNT files...\n"
	@cd ui && npx @biomejs/biome check src/ 2>&1 | tail -5 || true

lint-c: ## Run C linter (clang-tidy, Linux only)
ifeq ($(UNAME),Linux)
	@printf "$(BOLD)🔍 Running C linter (clang-tidy)...$(RESET)\n"
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
	@printf "$(GREEN)✓ C lint complete$(RESET)\n"
else
	@echo "C linting requires Linux"
endif

lint-md: ## Lint markdown files with markdownlint
	@printf "$(BOLD)🔍 Linting markdown files...$(RESET)\n"
	@if command -v markdownlint-cli2 > /dev/null 2>&1; then \
		markdownlint-cli2 "**/*.md"; \
	elif npx markdownlint-cli2 --help > /dev/null 2>&1; then \
		npx markdownlint-cli2 "**/*.md"; \
	else \
		printf "$(YELLOW)SKIP: markdownlint-cli2 not installed (npm install -g markdownlint-cli2)$(RESET)\n"; \
	fi
	@printf "$(GREEN)✓ Markdown lint complete$(RESET)\n"

# =============================================================================
# Formatting
# =============================================================================

fmt: fmt-go fmt-frontend ## Format all code
	@printf "$(GREEN)✓ All code formatted$(RESET)\n"

fmt-go: ## Format Go code
	@printf "$(BOLD)🔧 Formatting Go code...$(RESET)\n"
	@gofmt -w -s .
	@printf "$(GREEN)✓ Go code formatted$(RESET)\n"

fmt-frontend: ## Format frontend code with Biome
	@printf "$(BOLD)🔧 Formatting frontend code (Biome)...$(RESET)\n"
	@cd ui && npx @biomejs/biome format --write src/
	@printf "$(GREEN)✓ Frontend code formatted$(RESET)\n"

fmt-c: ## Format C code (Linux only)
ifeq ($(UNAME),Linux)
	@printf "$(BOLD)🔧 Formatting C code...$(RESET)\n"
	@if ! command -v clang-format >/dev/null 2>&1; then \
		echo "clang-format not found; install it to format C code."; \
		exit 1; \
	fi
	find src include tests -type f \( -name '*.c' -o -name '*.h' \) | xargs clang-format -i
	@printf "$(GREEN)✓ C code formatted$(RESET)\n"
else
	@echo "C formatting requires Linux"
endif

# =============================================================================
# Auto-Fix
# =============================================================================

fix: ## Auto-fix Go and frontend linting issues
	@printf "$(BOLD)🔧 Auto-fixing code...$(RESET)\n"
	@GOLANGCI_LINT="$$(go env GOPATH)/bin/golangci-lint"; \
	if [ ! -f "$$GOLANGCI_LINT" ]; then \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.1; \
	fi; \
	$$GOLANGCI_LINT run --fix ./...
	@gofmt -w -s .
	@cd ui && npx @biomejs/biome check --write .
	@printf "$(GREEN)✓ Auto-fix complete$(RESET)\n"

fix-all: fix fmt-c ## Auto-fix all code (Go + frontend + C)
