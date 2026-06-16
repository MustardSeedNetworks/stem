# ADR 0011: Decompose `internal/api` into isolated sub-packages

**Status:** Accepted (2026-06-16)

## Context

`internal/api` is the Stem HTTP transport layer (handlers, routing, the
capability registry). It has accreted several **self-contained infrastructure
concerns** into the same flat package alongside the transport code:

- the per-client rate limiter (`ratelimit.go`, ~334 LOC);
- the SSE hub + streaming engine (`sse.go` / `handlers_sse.go` /
  `sse_publishers.go`, ~345 LOC);
- the TLS listener setup + client-fingerprint logic (`tls.go` /
  `tls_fingerprint.go`, ~462 LOC).

Because everything is one Go package, any file can reach any other: there is no
signal or boundary stopping, say, the SSE engine from coupling to an auth
handler. The dependency-direction depguard (ADR-0003) protects the domain core
from importing the API layer, but it cannot express boundaries *within*
`internal/api`. The flat package is also the root cause of the snake_case
filename violations (`handlers_*.go`, `sse_*.go`, `tls_*.go`, …): the prefixes
are a package emulated inside one directory.

This mirrors the situation seed (ADR-0001/0016/0020 strangle) and niac
(**ADR-0006**, Accepted 2026-06-08) already addressed. Stem aligns to the same
fleet convention. Note: stem's per-session CSRF manager already lives in its own
package (`internal/auth/csrf.go`), so it is **not** part of this decomposition.

## Decision

Extract the cohesive infrastructure concerns into sibling **leaf packages** under
`internal/api/<concern>/`, one at a time, lowest-coupling first
(`ratelimit` → `sse` → `tls`). Then strangle the remaining capability handlers
(`mfa`, `reflector`, `settings`, `license`, …) out of the flat layer so that
`internal/api` becomes pure transport + composition, with role-named files
(`handler.go`, `types.go`, …) and **no snake_case filenames**.

Each leaf package:

- owns its type(s), unexported helpers, and tests;
- imports **only** the standard library and inward domain packages — never the
  `internal/api` transport layer;
- takes its transport-specific dependencies (client-IP extraction, error
  rendering, event sources) as **injected function values**, so the dependency
  arrow points inward;
- is composed by the `Server` (and `internal/daemon` where relevant), which
  holds the concrete manager and wires it into the declarative route registry
  (`register()`/`registerAll()` in `route.go`). The registry, `/__capabilities`,
  the middleware composition order, and all security invariants
  (CSRF, rate-limit, role/scope gates) are **unchanged** — only *where the
  building blocks live* changes.

Each extraction adds a depguard `api-<concern>-isolated` rule (modelled on the
existing inward-only rule and niac's ADR-0006 rules) that denies the sub-package
importing `github.com/MustardSeedNetworks/stem/internal/api`, so the leaf
boundary is **CI-enforced**, not convention.

Once the flat layer's snake_case filenames are eliminated, port seed's
`check-filename-policy.sh` gate into stem CI so the convention cannot regress.

## Consequences

- Boundaries within the API layer become explicit and CI-enforced; a future
  change cannot silently couple a leaf to the transport layer.
- The snake_case filename violations disappear as a *consequence* of
  decomposition (role-named files inside real packages), not via a cosmetic
  rename — matching the fleet "guards encode best practice" stance.
- Behaviour is preserved at every step: extractions are verbatim moves with
  inward-injected dependencies; the route registry and security middleware are
  untouched. Each slice is independently gated (build, vet, full golangci,
  tests, route-policy + output-escaping gates) before merge.
- Slices are **serial**, not parallel: they share `server.go` wiring and
  `.golangci.yml`, so concurrent extraction PRs collide. Land one, rebase the
  next.
- Stem reaches the same enforced-decomposition state as seed and niac, closing
  the fleet-parity gap for `internal/api`.
