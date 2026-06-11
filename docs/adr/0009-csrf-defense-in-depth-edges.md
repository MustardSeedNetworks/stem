# ADR 0009: Accepted CSRF defense-in-depth edges (refresh, trial)

**Status:** Accepted (2026-06-11)

## Context

A cross-repo security-invariant audit (seed/stem/niac) confirmed Stem's core
posture from ADR-0004 holds: every mutating `/api` route goes through the
capability registry (ADR-0001) and the global per-session `CSRFManager`
(`server.go`, wrapping `s.mux`), and the `check-route-policy.sh` CI gate bans
direct mux registration. The audit surfaced two *defense-in-depth* edges where
the global CSRF middleware does not actively validate a token. Both are blocked
against browser-driven CSRF by `SameSite=Strict` on the auth cookies; neither is
an invariant violation, but both deserve a recorded decision so a future audit
does not re-flag them as gaps.

1. **`POST /api/v1/auth/refresh`** — `CSRFMiddleware` derives the session id via
   `GetSessionIDFromRequest`, which reads the **access**-token cookie. The
   canonical refresh scenario is an expired access token: once it expires the
   browser stops sending the cookie, so `sessionID == ""` and the middleware
   takes its skip-on-empty-session path (which preserves the UI's "401 if
   unauthenticated, 403 if CSRF-missing" contract). The result is that refresh —
   a state change on a session-scoped credential — runs without an active CSRF
   check in exactly its normal case.

2. **`POST /api/v1/license/trial`** — starts a trial clock. It is intentionally
   pre-session (self-serve trial activation before an account exists), so
   `sessionID == ""` and CSRF is skipped, and it uses the general `apiLimiter`
   (100/min) rather than `authLimiter` (5/min).

## Decision

- **Both are accepted as-is; no behavior change.** `SameSite=Strict` on the
  refresh and auth cookies blocks the browser-driven CSRF vector for refresh,
  and trial activation is self-serve by design with low abuse impact (it only
  starts a trial clock — no credential, no persistent escalation).
- The mitigations are load-bearing and recorded here and in code comments at
  both sites so the decision is discoverable from the source.

## Consequences

- Refresh's CSRF protection rests on `SameSite=Strict` rather than an active
  token check. The defensive token-validation path remains for the case where a
  (non-expired) access-token cookie is present.
- Trial activation stays frictionless for the self-serve flow.
- A future audit that re-discovers either edge should land here, not re-open it.

## Alternatives considered

- **Key the refresh CSRF check off the refresh-token session.** The clean
  defense-in-depth fix, but it changes session-identity derivation and risks the
  documented 401-vs-403 UX contract the UI relies on. Not justified while
  `SameSite=Strict` already closes the browser vector. Revisit if cookie
  `SameSite` posture is ever relaxed.
- **Require auth / a setup token on trial activation.** Breaks the self-serve
  trial flow for no real security gain (no credential is exposed). Rejected.
- **Tighten trial to `authLimiter` (5/min).** Marginal; deferred — 100/min on a
  start-a-trial-clock endpoint is not a meaningful abuse vector.

## Related

- ADR-0001 (capability registry), ADR-0004 (auth + CORS posture)
- Cross-repo security-invariant audit, 2026-06-11
