# =============================================================================
# Security Scanning Targets
# =============================================================================
#
# Security and compliance targets:
#   - Go vulnerability scanning (govulncheck, gosec, staticcheck)
#   - npm audit
#   - Secret scanning (gitleaks)
#   - Container scanning (Trivy)
#   - License compliance
#
# =============================================================================

.PHONY: security security-backend security-backend-quiet security-frontend security-frontend-quiet \
        security-secrets security-secrets-quiet security-trivy \
        license-check license-check-go license-check-npm license-report

# =============================================================================
# Security Scanning
# =============================================================================

security: ## Run all security scans
	@printf "$(BOLD)$(CYAN)┌─ Security Scanning ──────────────────────────────────────────────────────────┐$(RESET)\n"
	@printf "$(CYAN)│$(RESET) $(BOLD)[1/3]$(RESET) Go Vulnerabilities (govulncheck)                                       $(CYAN)│$(RESET)\n"
	$(call timer-start,security-backend)
	@$(MAKE) --no-print-directory security-backend-quiet
	$(call timer-end,security-backend,Go vulnerability scan)
	@printf "$(CYAN)│$(RESET) $(BOLD)[2/3]$(RESET) npm Vulnerabilities (npm audit)                                        $(CYAN)│$(RESET)\n"
	$(call timer-start,security-frontend)
	@$(MAKE) --no-print-directory security-frontend-quiet
	$(call timer-end,security-frontend,npm audit)
	@printf "$(CYAN)│$(RESET) $(BOLD)[3/3]$(RESET) Secret Scanning (gitleaks)                                             $(CYAN)│$(RESET)\n"
	$(call timer-start,security-secrets)
	@$(MAKE) --no-print-directory security-secrets-quiet
	$(call timer-end,security-secrets,Secret scan)
	@printf "$(CYAN)└──────────────────────────────────────────────────────────────────────────────┘$(RESET)\n"

security-backend: ## Run Go security scans
	@printf "$(BOLD)🔒 Running Go security scans...$(RESET)\n"
	$(call timer-start,security-backend)
	@FAILED=0; \
	printf "  [1/3] Running govulncheck...\n"; \
	if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./... || FAILED=1; \
	else \
		printf "$(YELLOW)    ⚠ govulncheck not installed (run: make tools-go)$(RESET)\n"; \
	fi; \
	printf "  [2/3] Running gosec...\n"; \
	if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./... || FAILED=1; \
	else \
		printf "$(YELLOW)    ⚠ gosec not installed (run: make tools-go)$(RESET)\n"; \
	fi; \
	printf "  [3/3] Running staticcheck...\n"; \
	if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./... || FAILED=1; \
	else \
		printf "$(YELLOW)    ⚠ staticcheck not installed (run: make tools-go)$(RESET)\n"; \
	fi; \
	if [ $$FAILED -ne 0 ]; then exit 1; fi
	$(call timer-end,security-backend,Go security scan)

security-backend-quiet:
	@if ! command -v govulncheck > /dev/null 2>&1; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@printf "   Scanning Go dependencies...\n"
	@OUTPUT=$$(govulncheck ./... 2>&1); STATUS=$$?; \
	echo "$$OUTPUT" | grep -E "(Vulnerability|No vulnerabilities)" | head -5; \
	if [ $$STATUS -ne 0 ]; then \
		echo "$$OUTPUT"; \
		exit $$STATUS; \
	fi

security-frontend: ## Run npm security audit
	@printf "$(BOLD)🔒 Running npm security audit...$(RESET)\n"
	$(call timer-start,security-frontend)
	cd ui && npm audit --audit-level=high
	$(call timer-end,security-frontend,npm security audit)

security-frontend-quiet:
	@printf "   Auditing npm packages...\n"
	@cd ui && OUTPUT=$$(npm audit --audit-level=high 2>&1); STATUS=$$?; \
	echo "$$OUTPUT" | grep -E "(found|vulnerabilities)" | head -3; \
	if [ $$STATUS -ne 0 ]; then \
		echo "$$OUTPUT"; \
		exit $$STATUS; \
	fi

security-secrets: ## Scan for secrets in codebase
	@printf "$(BOLD)🔒 Scanning for secrets (gitleaks)...$(RESET)\n"
	$(call timer-start,security-secrets)
	@if command -v gitleaks >/dev/null 2>&1; then \
		if [ -f .gitleaks.toml ]; then \
			gitleaks detect --source . --config .gitleaks.toml --verbose; \
		else \
			gitleaks detect --source . --verbose; \
		fi; \
	else \
		printf "$(YELLOW)⚠ gitleaks not installed (run: make tools-go)$(RESET)\n"; \
	fi
	$(call timer-end,security-secrets,Secret scan)

security-secrets-quiet:
	@GITLEAKS=$$(command -v gitleaks 2>/dev/null || echo "$$(go env GOPATH)/bin/gitleaks"); \
	if [ ! -x "$$GITLEAKS" ]; then \
		go install github.com/zricethezav/gitleaks/v8@latest; \
		GITLEAKS="$$(go env GOPATH)/bin/gitleaks"; \
	fi; \
	printf "   Scanning for secrets...\n"; \
	if [ -f .gitleaks.toml ]; then \
		OUTPUT=$$($$GITLEAKS detect --source . --config .gitleaks.toml 2>&1); \
	else \
		OUTPUT=$$($$GITLEAKS detect --source . 2>&1); \
	fi; \
	STATUS=$$?; \
	echo "$$OUTPUT" | grep -E "(leaks found|no leaks)"; \
	if [ $$STATUS -ne 0 ]; then \
		echo "$$OUTPUT"; \
		exit $$STATUS; \
	fi

security-trivy: ## Run Trivy vulnerability scan
	@printf "$(BOLD)🔒 Running Trivy filesystem scan...$(RESET)\n"
	@if command -v trivy > /dev/null 2>&1; then \
		trivy fs --severity HIGH,CRITICAL .; \
	else \
		printf "$(YELLOW)SKIP: trivy not installed (brew install trivy)$(RESET)\n"; \
	fi

# =============================================================================
# License Compliance
# =============================================================================

license-check: license-check-go license-check-npm ## Check dependency licenses
	@printf "\n$(GREEN)✓ License check complete$(RESET)\n"

license-check-go: ## Check Go module licenses
	@printf "$(BOLD)🔍 Checking Go dependency licenses...$(RESET)\n"
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		printf "$(YELLOW)Installing go-licenses...$(RESET)\n"; \
		go install github.com/google/go-licenses@latest; \
	fi
	@go-licenses check ./... \
		--disallowed_types=forbidden,restricted \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

license-check-npm: ## Check npm package licenses
	@printf "$(BOLD)🔍 Checking npm dependency licenses...$(RESET)\n"
	@cd ui && npx license-checker --summary --onlyAllow \
		"MIT;Apache-2.0;BSD-2-Clause;BSD-3-Clause;ISC;CC0-1.0;Unlicense;0BSD" \
		2>/dev/null || printf "$(YELLOW)⚠ Some license issues found$(RESET)\n"

license-report: ## Generate full license report
	@printf "$(BOLD)Generating license report...$(RESET)\n"
	@mkdir -p reports
	@printf "Go Licenses:\n" > reports/licenses.txt
	@printf "============\n" >> reports/licenses.txt
	@go-licenses csv ./... 2>/dev/null >> reports/licenses.txt || true
	@printf "\n\nnpm Licenses:\n" >> reports/licenses.txt
	@printf "=============\n" >> reports/licenses.txt
	@cd ui && npx license-checker --csv 2>/dev/null >> ../reports/licenses.txt || true
	@printf "$(GREEN)✓ License report: reports/licenses.txt$(RESET)\n"
