# =============================================================================
# Deployment validation
# =============================================================================
# Packages are built by goreleaser in CI, not here. What the build contract
# needs locally is the check that runs AFTER an install: that the service
# answering on the host is the artifact the release published.
#
#   make deploy-validate HOST=10.44.40.23 RELEASE=v0.26.0
#
# HOST is the address stem serves HTTPS on, supplied by variable because the
# last set of hardcoded host names went stale when the lab re-addressed.
# RELEASE, not VERSION: mk/vars.mk owns VERSION as this tree's `git describe`.
# =============================================================================

.PHONY: deploy-validate

deploy-validate: ## Validate /__version of an installed release (HOST=, RELEASE=, PORT=)
ifndef HOST
	$(error HOST is required, e.g. make deploy-validate HOST=10.44.40.23 RELEASE=v0.26.0)
endif
ifndef RELEASE
	$(error RELEASE is required, e.g. make deploy-validate HOST=10.44.40.23 RELEASE=v0.26.0)
endif
	@scripts/deploy-validate.sh --host $(HOST) --release $(RELEASE) \
		$(if $(PORT),--port $(PORT))
