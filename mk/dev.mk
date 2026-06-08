# =============================================================================
# Dev-experience helpers — capability-aware run for local iteration
# =============================================================================
# Production service installs use units that already declare CAP_NET_RAW +
# CAP_NET_ADMIN + CAP_NET_BIND_SERVICE and run as a
# non-root user. For local dev (`make build && ./bin/stem web`) those
# capabilities have to be set on the binary OR you have to sudo. These
# targets handle that automatically per platform.
#
#   Linux: `sudo setcap` once on the binary, then run as user (no sudo)
#   macOS: `sudo` the binary (macOS userspace has no fcaps for raw sockets)
#
# Stem also needs STEM_AUTH_USERNAME / STEM_AUTH_PASSWORD env vars; for dev
# they default to admin/admin if unset. NEVER ship that in production —
# this default only applies under `make dev-*`.
# =============================================================================

.PHONY: dev-run dev-start dev-stop dev-status dev-restart

DEV_BINARY ?= ./bin/stem
DEV_ARGS   ?= web -p 8081
DEV_PID    ?= /tmp/stem.pid
DEV_LOG    ?= /tmp/stem.log
DEV_AUTH_USER ?= admin
DEV_AUTH_PASS ?= admin

DEV_UNAME := $(shell uname -s)

dev-run: build ## Build + run in foreground with required capabilities
ifeq ($(DEV_UNAME),Linux)
	@if ! getcap $(DEV_BINARY) 2>/dev/null | grep -q cap_net_raw; then \
		printf "$(BOLD)Setting capabilities on $(DEV_BINARY) (one-time per build)...$(RESET)\n"; \
		sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+ep $(DEV_BINARY) || { \
			printf "$(YELLOW)setcap unavailable — falling back to sudo$(RESET)\n"; \
			exec sudo STEM_AUTH_USERNAME=$(DEV_AUTH_USER) STEM_AUTH_PASSWORD=$(DEV_AUTH_PASS) $(DEV_BINARY) $(DEV_ARGS); \
		}; \
	fi
	@printf "$(GREEN)→ $(DEV_BINARY) $(DEV_ARGS)$(RESET)\n"
	@STEM_AUTH_USERNAME=$(DEV_AUTH_USER) STEM_AUTH_PASSWORD=$(DEV_AUTH_PASS) $(DEV_BINARY) $(DEV_ARGS)
else
	@printf "$(BOLD)Running via sudo (macOS has no userspace fcaps for raw sockets)$(RESET)\n"
	@sudo STEM_AUTH_USERNAME=$(DEV_AUTH_USER) STEM_AUTH_PASSWORD=$(DEV_AUTH_PASS) $(DEV_BINARY) $(DEV_ARGS)
endif

dev-start: build ## Background variant; PID -> $(DEV_PID), logs -> $(DEV_LOG)
ifeq ($(DEV_UNAME),Linux)
	@if ! getcap $(DEV_BINARY) 2>/dev/null | grep -q cap_net_raw; then \
		sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+ep $(DEV_BINARY); \
	fi
	@STEM_AUTH_USERNAME=$(DEV_AUTH_USER) STEM_AUTH_PASSWORD=$(DEV_AUTH_PASS) \
		nohup $(DEV_BINARY) $(DEV_ARGS) > $(DEV_LOG) 2>&1 & echo $$! > $(DEV_PID)
else
	@nohup sudo STEM_AUTH_USERNAME=$(DEV_AUTH_USER) STEM_AUTH_PASSWORD=$(DEV_AUTH_PASS) \
		$(DEV_BINARY) $(DEV_ARGS) > $(DEV_LOG) 2>&1 & echo $$! > $(DEV_PID)
endif
	@sleep 2
	@printf "$(GREEN)✓ started (PID $$(cat $(DEV_PID))). log: $(DEV_LOG)$(RESET)\n"
	@printf "  Web UI: https://localhost:8081 (user: $(DEV_AUTH_USER), pass: $(DEV_AUTH_PASS))\n"

dev-stop: ## Stop a backgrounded dev-start
	@if [ -f $(DEV_PID) ] && ps -p $$(cat $(DEV_PID)) > /dev/null 2>&1; then \
		kill $$(cat $(DEV_PID)) && rm -f $(DEV_PID); \
		printf "$(GREEN)✓ stopped$(RESET)\n"; \
	else \
		printf "$(YELLOW)not running$(RESET)\n"; \
		rm -f $(DEV_PID); \
	fi

dev-status: ## Show whether the backgrounded process is running
	@if [ -f $(DEV_PID) ] && ps -p $$(cat $(DEV_PID)) > /dev/null 2>&1; then \
		printf "$(GREEN)✓ running (PID $$(cat $(DEV_PID))), log: $(DEV_LOG)$(RESET)\n"; \
	else \
		printf "$(YELLOW)not running$(RESET)\n"; \
	fi

dev-restart: dev-stop dev-start ## Stop + start
