# ADR-0011: internal/api sub-package decomposition

**Status:** Accepted  
**Date:** 2026-06-16

## Context

`internal/api` is a single ~6 000-line package that contains the HTTP server,
all route handlers, middleware, the SSE broadcaster, and the rate limiter in
one flat namespace. Cross-cutting concerns (rate limiting, SSE) have no formal
boundary: any file can import any symbol, and the dependency direction from
leaf concerns back into the HTTP transport is invisible to static analysis.

The first decomposition slice (PR #451) extracted the rate-limiter into
`internal/api/ratelimit`. This ADR records the second slice and the general
decomposition plan.

## Decision

Extract isolated, self-contained concerns from `internal/api` into **leaf
sub-packages** of the form `internal/api/<concern>`. A leaf:

- imports ONLY stdlib and other inward packages (never the api transport layer)
- exports a constructor (`New(...)` or equivalent) returning an exported type
- is depguard-gated: a rule named `api-<concern>-isolated` prevents accidental
  upward imports in CI

The api transport layer wires leaves at construction time; no leaf knows about
`internal/api`.

### Completed slices

| Slice | Package | PR |
|-------|---------|-----|
| Rate limiter | `internal/api/ratelimit` | #451 |
| SSE broadcaster | `internal/api/sse` | this ADR |

### SSE slice (this ADR)

`internal/api/sse` holds the `Broadcaster` type (fan-out engine), `Frame`
(wire type + `Encode`), and the subscriber bookkeeping. It has zero HTTP
imports. The HTTP handler (`handleSSEEvents`), the publisher loop
(`runReflectorStatsPublisher`), and the endpoint wiring stay in `internal/api`.

The `sse.HeartbeatInterval` constant is exported so the transport layer can
reference it without duplicating the value; the api-layer file (`sse.go`)
re-derives its own constant from the same numeric literal to avoid an import
dependency in the other direction.

### Future slices (candidates)

| Concern | Notes |
|---------|-------|
| TLS utilities | `ensureSelfSignedCert`, `createTLSConfig`, ACME helpers |
| CORS logic | RFC 1918 origin validation |

## Consequences

- The leaf boundary is statically enforced by depguard (`api-sse-isolated`,
  `api-ratelimit-isolated` rules in `.golangci.yml`).
- `go vet` + `golangci-lint` catch upward imports at CI time.
- `internal/api` package size decreases incrementally with each slice.
- No behaviour change: endpoints, event types, and publish sites are identical.
