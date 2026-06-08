# ADR 0002: Single source of truth for module metadata + executor

**Status:** Accepted (2026-06-07)

## Context

A test module had its **metadata** (name, test types, display info) in the
services `Registry`, but the factory that constructs its **executor** lived in a
separate hardcoded map in the API layer (`internal/api/executors.go`). The two
were keyed by module name independently, so adding a module meant editing both
places — and forgetting the factory produced a silent runtime
`executor not implemented for module: X`.

The module `Module` interface is intentionally metadata-only (execution is a
separate concern), and the executors are dataplane-backed (cgo on Linux), so the
two cannot simply be merged onto one interface without an import cycle (the
sub-modules would have to import their parent, which imports them).

## Decision

The registry owns **both** the module and its executor factory, registered
together in one place (`internal/services/init.go::buildDefaultRegistry`):

- `modtypes` (the existing cycle-free shared package) defines the canonical
  `Executor` and `ExecutorFactory` types.
- `Registry` gains a `factories` map, `RegisterExecutable(module, factory)`, and
  `Factory(name)`.
- The API layer's `testExecutor` / `executorFactory` become **type aliases** of
  the `modtypes` types, and `executeTest` resolves the factory via
  `services.Factory` instead of a local map.
- The reflector registers **metadata-only** (`Register`) because it has a
  distinct lifecycle and is dispatched directly, not via a factory.

## Consequences

- Adding a module is a **single edit** in `buildDefaultRegistry`; metadata and
  execution can no longer drift apart.
- The API layer no longer imports each module sub-package; it depends on the
  registry. `registry_executor_test.go` pins the invariant.
- The metadata/execution separation is preserved (the factory is a registry
  attribute, not a method on the `Module` interface).

## Alternatives considered

- **`Execute` on the `Module` interface**: forces the reflector (no executor) to
  implement it and risks an import cycle via the shared executor type. Rejected.
- **A sync-check test** (assert the two maps match): guards drift but keeps two
  sources of truth. Rejected in favor of actually collapsing them.

## Related issues and PRs

- #404 (this collapse)
