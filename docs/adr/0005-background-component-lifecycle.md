# ADR 0005: Background-component lifecycle holder (composition root for long-lived goroutines)

**Status:** Accepted (2026-06-08)

## Context

`internal/api/server.go` had grown to ~990 lines and owned every concern of the
web server, including the lifecycle of its long-lived background goroutines. The
reflector-stats SSE publisher (#296) was launched ad hoc inside `Run` with a
fire-and-forget `go func()` whose only stop signal was context cancellation —
`Shutdown` never waited for it to exit, so on shutdown the goroutine could still
be mid-broadcast against state being torn down. As more server-push producers
arrive (the test-progress publisher is a planned #296 follow-up) the pattern of
scattering `go func()` launches through `Run` would repeat the same gap.

seed solved the equivalent problem with a `BackgroundComponents` holder that
exposes one ordered `Start(ctx)`/`Stop()` seam (seed `internal/api/background.go`).
The harmonization directive is to match seed's *shape*, not copy its code. seed's
holder owns detached feature services (the reporting scheduler, the Wi-Fi
visibility loop); stem's single background producer instead reads live `Server`
state (the reflector executor and stats), so a verbatim port does not fit.

## Decision

Introduce `internal/api/background.go` with a `BackgroundComponents` holder that
owns the **Run-scoped** background goroutines and provides one ordered
`Start(ctx)`/`Stop()` lifecycle:

- The reflector-stats publisher becomes a blocking `Server.runReflectorStatsPublisher(ctx)`
  loop; `BackgroundComponents` owns the goroutine, a derived cancellable context,
  and a `sync.WaitGroup`.
- `Start(ctx)` launches the goroutine under a context derived from `Run`'s signal
  context. `Stop()` cancels it and **blocks until it has exited**, closing the
  shutdown race. `Stop()` is nil-safe (no-op when `Start` was never called) and
  idempotent.
- `Server.Run` constructs the holder and calls `Start`; `Server.Shutdown` calls
  `Stop` first — before tearing down the reflector executor the publisher reads.

Construction-scoped cleanup goroutines (the rate limiters, CSRF manager, auth
manager created in `NewServer`) are deliberately **not** moved into the holder.
They outlive `Run` and must be stopped by `Shutdown` even when `Run` was never
called — the test suite constructs servers and calls `Shutdown` directly. Folding
them into a holder gated by `Start` would leak them whenever `Start` was skipped.

## Consequences

- A single seam owns long-lived goroutine lifecycle; new server-push producers
  register in `BackgroundComponents` instead of adding `go func()` to `Run`.
- Shutdown now deterministically waits for the publisher to finish, removing a
  latent use-during-teardown race (verified by `background_test.go` under `-race`).
- The holder keeps a back-reference to the `Server` rather than owning standalone
  services — an intentional divergence from seed driven by stem's coupling of the
  publisher to live server state.
- A full `internal/app` package extraction of `NewServer` was scoped out: stem's
  `Server` is tightly coupled to `internal/api` internals, so a wrapper package
  would add indirection without moving real wiring. Revisit if feature services
  with independent lifecycles are added.

## Related issues and PRs

- #296 (SSE publishers), and the composition-root item in the stem/niac
  remediation plan.
