# =============================================================================
# Security Scanning Targets
# =============================================================================
#
# Security and compliance targets:
#   - Go vulnerability scanning (govulncheck, gosec, staticcheck)
#   - npm audit
#   - Secret scanning (gitleaks)
#   - License compliance
#
# =============================================================================

.PHONY: security security-backend security-frontend security-secrets \
        license-check license-check-go license-check-npm license-report

# =============================================================================
# Security Scanning
# =============================================================================

security: security-backend security-frontend security-secrets ## Run all security scans
	@printf "\n$(GREEN)✓ All security scans complete$(RESET)\n"

security-backend: ## Run Go security scans
	@printf "$(BOLD)$(CYAN)Running Go security scans...$(RESET)\n"
	$(call timer-start,security-backend)
	@printf "  [1/3] Running govulncheck...\n"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ govulncheck not installed (run: make tools-go)$(RESET)\n"; \
	fi
	@printf "  [2/3] Running gosec...\n"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ gosec not installed (run: make tools-go)$(RESET)\n"; \
	fi
	@printf "  [3/3] Running staticcheck...\n"
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./... || true; \
	else \
		printf "$(YELLOW)    ⚠ staticcheck not installed (run: make tools-go)$(RESET)\n"; \
	fi
	$(call timer-end,security-backend,Go security scan)

security-frontend: ## Run npm security audit
	@printf "$(BOLD)$(CYAN)Running npm security audit...$(RESET)\n"
	$(call timer-start,security-frontend)
	cd ui && npm audit --audit-level=high || true
	$(call timer-end,security-frontend,npm security audit)

security-secrets: ## Scan for secrets in codebase
	@printf "$(BOLD)$(CYAN)Scanning for secrets...$(RESET)\n"
	$(call timer-start,security-secrets)
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --source . --verbose || true; \
	else \
		printf "$(YELLOW)⚠ gitleaks not installed (run: make tools-go)$(RESET)\n"; \
	fi
	$(call timer-end,security-secrets,Secret scan)

# =============================================================================
# License Compliance
# =============================================================================

license-check: license-check-go license-check-npm ## Check dependency licenses
	@printf "\n$(GREEN)✓ License check complete$(RESET)\n"

license-check-go: ## Check Go module licenses
	@printf "$(BOLD)$(CYAN)Checking Go dependency licenses...$(RESET)\n"
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		printf "$(YELLOW)Installing go-licenses...$(RESET)\n"; \
		go install github.com/google/go-licenses@latest; \
	fi
	@go-licenses check ./... \
		--disallowed_types=forbidden,restricted \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

license-check-npm: ## Check npm package licenses
	@printf "$(BOLD)$(CYAN)Checking npm dependency licenses...$(RESET)\n"
	@cd ui && npx license-checker --summary --onlyAllow \
		"MIT;Apache-2.0;BSD-2-Clause;BSD-3-Clause;ISC;CC0-1.0;Unlicense;0BSD" \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

license-report: ## Generate full license report
	@printf "$(BOLD)$(CYAN)Generating license report...$(RESET)\n"
	@mkdir -p reports
	@printf "Go Licenses:\n" > reports/licenses.txt
	@printf "============\n" >> reports/licenses.txt
	@go-licenses csv ./... 2>/dev/null >> reports/licenses.txt || true
	@printf "\n\nnpm Licenses:\n" >> reports/licenses.txt
	@printf "=============\n" >> reports/licenses.txt
	@cd ui && npx license-checker --csv 2>/dev/null >> ../reports/licenses.txt || true
	@printf "$(GREEN)✓ License report: reports/licenses.txt$(RESET)\n"
