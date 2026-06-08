# ADR 0003: Dependency direction enforced by depguard

**Status:** Accepted (2026-06-07)

## Context

The module/service layer (`internal/services`) is the inward core: it owns test
modules and their executors and must not depend on the web/API, auth, or license
layers — those concerns are composed at the API layer. This direction was clean
**by convention** but unenforced, so a future upward import (e.g. a module
reaching into `internal/api` or `internal/license`) would compile and ship
silently, coupling the layers.

## Decision

Add a `depguard` rule (`service-layer-inward-only`) to `.golangci.yml` that bars
`internal/services/**` from importing `internal/api`, `internal/auth`, or
`internal/license`. golangci-lint already runs the strict golden config; this
makes the dependency direction a hard CI gate.

## Consequences

- An upward import now fails CI with an actionable message instead of silently
  coupling the layers — enforcement by construction for architecture, the same
  pattern seed uses.
- License-feature gating and authentication stay an API-layer responsibility;
  modules remain transport/auth-agnostic and independently testable.
- The rule is verified to fire (a probe import is rejected) and pass on the
  current tree.

## Alternatives considered

- **Convention + code review**: the status quo that left the direction
  unenforced. Rejected.

## Related issues and PRs

- #400 (this rule)
