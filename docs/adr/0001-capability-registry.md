# ADR 0001: Capability registry for route policy

**Status:** Accepted (2026-06-07)

## Context

API routes were registered imperatively in `setupRoutes`, each one hand-wrapped
with its middleware (`handleAuthRateLimited` / `handleRateLimited` / raw
`s.mux.Handle`). The per-route policy — authentication and rate limiting — was
re-applied at every call site, so the nesting could be applied inconsistently or
forgotten. This is exactly how `POST /api/v1/reflector/config` shipped
**unauthenticated**: it was registered with only a rate limiter (fixed in #398).
"Forgot the wrapper" is a documented, recurring regression class.

## Decision

Routes are declared as **data** and a single `register()` / `registerAll()`
composes their middleware in **one canonical order** for every route:
`rateLimit → auth → methodGate → handler` (CSRF is global). A route is a
`route{path, handler, methods, auth, limiter}` value; `register()` records each
route in a manifest and installs it on the mux. See `internal/api/route.go`.

A route therefore **cannot be installed without its policy**. Two supporting
mechanisms make this enforceable:

- `GET /__capabilities` serves the route-policy manifest (auth/rate-limit per
  route) for deployment/audit introspection.
- `scripts/check-route-policy.sh` (a CI gate) fails if any `/api/` route is
  registered directly via `s.mux.Handle*`/`s.handle` instead of through
  `register()`.

## Consequences

- Adding or changing a route's policy is declarative and uniform; the policy is
  visible in one table.
- The reflector-hole regression class is structurally prevented (enforcement by
  construction), not merely caught in review.
- The capability layer mirrors seed's ADR-0002, harmonizing the fleet while each
  repo keeps its own implementation.

## Alternatives considered

- **A grep-only CI gate** (no registry): catches a direct registration but does
  not prevent it — still relies on every author wrapping correctly. Rejected as a
  band-aid; kept only as a complement to the registry.
- **Per-route middleware at call sites** (status quo): the source of the bug.

## Related issues and PRs

- #398 (reflector authentication — the motivating regression)
- #401 (this registry: `route.go`, `/__capabilities`, `check-route-policy.sh`)
