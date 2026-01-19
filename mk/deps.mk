# =============================================================================
# Dependency Management & Tools
# =============================================================================
#
# Dependency and tooling management:
#   - Go modules and npm package management
#   - Development tools installation
#   - Version checking and updates
#
# =============================================================================

.PHONY: update update-go update-npm version-check tools tools-go tools-frontend

# =============================================================================
# Dependency Updates
# =============================================================================

update: update-go update-npm ## Update all dependencies
	@printf "\n$(GREEN)✓ All dependencies updated$(RESET)\n"
	@printf "$(YELLOW)Remember to test before committing!$(RESET)\n"

update-go: ## Update Go modules
	@printf "$(BOLD)$(CYAN)Updating Go dependencies...$(RESET)\n"
	$(call timer-start,update-go)
	go get -u ./...
	go mod tidy
	$(call timer-end,update-go,Go dependencies update)

update-npm: ## Update npm packages
	@printf "$(BOLD)$(CYAN)Updating npm dependencies...$(RESET)\n"
	$(call timer-start,update-npm)
	cd ui && npm update
	cd ui && npm audit fix || true
	$(call timer-end,update-npm,npm dependencies update)

# =============================================================================
# Version Checking
# =============================================================================

version-check: ## Show version info and outdated packages
	@printf "$(BOLD)$(CYAN)Version Information$(RESET)\n"
	@printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	@printf "$(BOLD)Runtime:$(RESET)\n"
	@printf "  Go:              $$(go version | awk '{print $$3}')\n"
	@printf "  Node.js:         $$(node --version)\n"
	@printf "  npm:             $$(npm --version)\n"
	@printf "\n$(BOLD)Go Tools:$(RESET)\n"
	@printf "  golangci-lint:   $$(golangci-lint --version 2>/dev/null | head -1 || echo 'not installed')\n"
	@printf "  govulncheck:     $$(govulncheck -version 2>/dev/null || echo 'not installed')\n"
	@printf "  staticcheck:     $$(staticcheck -version 2>/dev/null || echo 'not installed')\n"
	@printf "\n$(BOLD)Dependencies:$(RESET)\n"
	@printf "  Go modules:      $$(go list -m all 2>/dev/null | wc -l | tr -d ' ') packages\n"
	@printf "  npm packages:    $$(cd ui && npm ls --depth=0 2>/dev/null | wc -l | tr -d ' ') packages\n"
	@printf "\n$(BOLD)Outdated:$(RESET)\n"
	@GO_OUTDATED=$$(go list -u -m all 2>/dev/null | grep '\[' | wc -l | tr -d ' '); \
	printf "  Go outdated:     $$GO_OUTDATED packages\n"
	@cd ui && npm outdated 2>/dev/null | tail -n +2 | wc -l | xargs -I {} printf "  npm outdated:    {} packages\n"

# =============================================================================
# Developer Tools
# =============================================================================

tools: tools-go tools-frontend ## Install all development tools
	@printf "\n$(GREEN)✓ All development tools installed$(RESET)\n"

tools-go: ## Install Go development tools
	@printf "$(BOLD)$(CYAN)Installing Go development tools...$(RESET)\n"
	$(call timer-start,tools-go)
	@printf "  Installing golangci-lint v2...\n"
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@printf "  Installing govulncheck...\n"
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@printf "  Installing gosec...\n"
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@printf "  Installing gofumpt...\n"
	go install mvdan.cc/gofumpt@latest
	@printf "  Installing goimports...\n"
	go install golang.org/x/tools/cmd/goimports@latest
	@printf "  Installing staticcheck...\n"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@printf "  Installing gitleaks...\n"
	go install github.com/zricethezav/gitleaks/v8@latest
	@printf "  Installing gotestsum...\n"
	go install gotest.tools/gotestsum@latest
	$(call timer-end,tools-go,Tool installation)
	@printf "\n$(GREEN)✓ All Go tools installed to $$(go env GOPATH)/bin$(RESET)\n"
	@printf "\nInstalled tools:\n"
	@printf "  • golangci-lint  - Comprehensive Go linter\n"
	@printf "  • govulncheck    - Go vulnerability checker\n"
	@printf "  • gosec          - Go security scanner\n"
	@printf "  • gofumpt        - Stricter gofmt\n"
	@printf "  • goimports      - Import management\n"
	@printf "  • staticcheck    - Static analysis\n"
	@printf "  • gitleaks       - Secret detection\n"
	@printf "  • gotestsum      - Better test output\n"

tools-frontend: ## Install frontend development tools
	@printf "$(BOLD)$(CYAN)Installing frontend tools...$(RESET)\n"
	cd ui && npm install
	@printf "$(GREEN)✓ Frontend tools installed$(RESET)\n"
