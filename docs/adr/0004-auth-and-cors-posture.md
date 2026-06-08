# ADR 0004: Single-admin auth model + CORS posture

**Status:** Accepted (2026-06-07)

## Context

The Stem is a single-appliance tool with one operator-provisioned account
(`STEM_AUTH_USERNAME` / `STEM_AUTH_PASSWORD`); the `UserStore` abstraction
(ported from seed) carries a `role` field but is wired single-user. A security
review flagged "any authenticated user is admin" as the top risk — but with
exactly one account that **is** the admin, role differentiation has no live
effect. Two genuine issues did exist: a dataplane endpoint shipped
unauthenticated, and CORS reflected any RFC1918 origin with credentials.

## Decision

- **Keep the single-admin model; do not build a role/permission system.** Adding
  viewer/operator/admin enforcement to a one-account appliance is scaffolding
  with no current benefit (over-engineering). Roles are revisited only if/when
  multi-user or scoped API tokens are introduced.
- **Every mutating API route requires authentication**, enforced through the
  capability registry (ADR-0001). The previously open
  `/api/v1/reflector/{config,stats}` endpoints are now authenticated (#398).
- **CSRF** is enforced globally via a per-session `CSRFManager`; **auth endpoints
  are rate-limited** (`authLimiter`, 5/min).
- **CORS is secure by default**: only localhost and same-origin are allowed
  (normal UI access is same-origin). Reflecting RFC1918 private-network origins
  with credentials — a cross-origin CSRF-bypass vector on a hostile LAN — is
  **opt-in** via `STEM_CORS_ALLOW_PRIVATE` (logs a warning when enabled).

## Consequences

- The audit's "everyone is admin" finding is documented as moot for the current
  single-admin design; the real holes (unauth dataplane route, RFC1918 CORS) are
  closed.
- Operators who need cross-origin LAN access make a deliberate, logged opt-in.
- A future multi-user requirement has a clear extension point: populate roles in
  the JWT `Claims` + add `requireRole`, gated through the registry.

## Alternatives considered

- **Full seed-style viewer/operator/admin roles now**: future-proofing with no
  live benefit for one account. Deferred.
- **niac-style scoped API tokens**: useful for automation, but a larger build
  than the appliance needs today. Deferred.
- **Remove RFC1918 CORS entirely**: would break legitimate LAN cross-origin
  integrations; the opt-in preserves the capability behind an explicit switch.

## Related issues and PRs

- #398 (reflector authentication), #399 (CORS opt-in)
