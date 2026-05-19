# Stem Auth Audit — 2026-05-19

Task: #77 — audit + modernize auth across all 3 repos.

Read-only audit of `internal/auth/` and `internal/api/` covering the
full login flow (hashing, session storage, CSRF, rate limit, lockout,
error messages, password policy, setup token, audit logging).

No severity ratings are assigned here. Humans should triage and
prioritize from the findings below.

---

## Summary

| Category                  | Result |
|---------------------------|--------|
| Password hashing          | WARN — bcrypt DefaultCost (10) (Argon2id is target) |
| Session token storage     | PASS — httpOnly cookies, refresh, JTI blacklist |
| CSRF protection           | WARN — middleware skips check if no token registered for session |
| Rate limiting             | PASS — 5/min per IP for auth, 100/min for API |
| Account lockout           | PASS — UserStore-backed (5 attempts / 15 min) |
| Error messages            | PASS — uniform "Invalid username or password" |
| Password policy           | WARN — length-only (>= 12), no complexity, no breach check |
| Setup wizard token        | PASS — single-use, 15 min, constant-time |
| Audit logging             | PASS — structured `SecurityEvent` with IP, UA, outcome |
| Client IP (auth bucket)   | FAIL — `getClientIP` trusts `X-Forwarded-For` unconditionally |

Totals: 6 PASS, 3 WARN, 1 FAIL.

---

## Checklist

### Password hashing algorithm
- **Result**: WARN
- **Where**: `internal/auth/password.go:17-23, 89-95`
- **Detail**: `bcryptCost()` returns `bcrypt.DefaultCost` (10) outside
  of test mode. RFC 9106 / NIST 800-63B recommend Argon2id for new code,
  and OWASP currently recommends bcrypt cost 12+ if bcrypt is retained.
- **Remediation note**: Either raise cost to 12 (seed parity) or
  migrate to Argon2id. Migration, not a fix.

### Session token storage
- **Result**: PASS
- **Where**:
  - JWT signing: `internal/auth/auth.go:166-196` (HS256, 256-bit
    secret via `crypto/rand`).
  - Cookie store: `internal/auth/cookie.go:49-78`
    (`HttpOnly: true`, `SameSite: SameSiteStrictMode`, `Secure`
    parameter is passed in by server based on TLS config).
  - Refresh: `internal/auth/auth.go:236-254`
    (15-min access, 7-day refresh).
  - Revocation: per-token JTI in claims + `TokenBlacklist`
    (`internal/auth/auth.go:198-234`).

### CSRF protection
- **Result**: WARN
- **Where**: `internal/auth/csrf.go:217-230`
- **Detail**: The middleware skips CSRF validation if no CSRF token
  has been generated for the session (`if !m.HasToken(sessionID) { ... }`).
  That makes CSRF protection opt-in by the client. A malicious
  cross-site request that arrives before the legitimate client has
  fetched `/api/v1/auth/csrf` will pass.
  In practice the UI always fetches the CSRF token immediately after
  login, but the middleware itself does not enforce that.
- **Remediation note**: Either issue the CSRF cookie/token at login
  time (so every authenticated session has one), or change the
  middleware to fail closed when no token is registered. Both are
  more than a 10-line change because callers in tests rely on the
  current soft-enforce behavior. Filed as followup.

### Rate limiting
- **Result**: PASS
- **Where**: `internal/api/ratelimit.go:101-115`
- **Detail**: Token-bucket via `golang.org/x/time/rate`. Auth bucket is
  5/min with burst 5; API bucket is 100/min. Global fallback limiter
  when max 10000 visitors hit; adaptive TTL cleanup at 80%/90%.

### Account lockout
- **Result**: PASS
- **Where**: `internal/auth/userstore.go:178-215`
- **Detail**: `MemoryUserStore.RecordLoginFailure` locks the account
  for 15 minutes after 5 failures; `IsLocked` honors the lock window.

### Error messages
- **Result**: PASS
- **Where**: `internal/auth/auth.go:127-129`,
  `internal/api/errors.go:240-248`
- **Detail**: Auth always returns `ErrInvalidCredentials` whether the
  username matched or not; the API maps this to a single
  "Invalid username or password" response. No user-enumeration leak.

### Password policy
- **Result**: WARN
- **Where**: `internal/auth/password.go:81-86`
- **Detail**: `ValidatePasswordStrength` only checks `len >= 12`. No
  character-class requirement, no breach-corpus lookup.
- **Remediation note**: A length-only policy is defensible per NIST
  800-63B if paired with a breach-corpus check. Add HIBP k-anonymity
  on setup (filed as followup).

### Setup wizard token
- **Result**: PASS
- **Where**: `internal/auth/setup_token.go:40-117`
- **Detail**: 32-byte base64url, single-use (marked `Used` after
  validation), 15-min expiry, constant-time compare.

### Audit logging
- **Result**: PASS
- **Where**: `internal/logging/audit.go:180-310`
- **Detail**: All login attempts, logouts, token refreshes, token
  invalid/expired/revoked events emit structured `SecurityEvent`
  records carrying IP, UA, username, request_id, reason.
  `AuditLoginFailure` also feeds a per-IP suspicious-activity tracker
  (`audit.go:212-232`).

### Client IP used for auth rate-limiting / audit logs
- **Result**: FAIL
- **Where**: `internal/api/ratelimit.go:253-284`,
  `internal/logging/audit.go:196` (uses `GetClientIP(r)` from logging
  package, but the API path uses the local `getClientIP` which trusts
  headers).
- **Detail**: `getClientIP` reads `X-Forwarded-For` and `X-Real-IP`
  *unconditionally*. An attacker can send
  `X-Forwarded-For: 1.2.3.4` to make every login attempt look like it
  came from a different IP, defeating the per-IP rate limiter and
  polluting audit logs.
- **Remediation note**: Restrict trust of forwarded headers to a
  configured trusted-proxy CIDR list (seed already does this via
  `TrustedProxies` in `internal/api/ratelimit.go:318-323`). Fix
  proposed below.

---

## Small fixes shipped in this PR

### `getClientIP` no longer trusts forwarded headers blindly
- **File**: `internal/api/ratelimit.go`
- **Change**: Restrict `X-Forwarded-For` and `X-Real-IP` parsing to
  requests whose `RemoteAddr` is a loopback address. Public-internet
  deployments will see the actual TCP peer address used for both
  rate-limiting and audit logging. Loopback is preserved so the
  intended dev-mode behavior (reverse proxy on localhost) still works.

This is a behavior-preserving fix for the common dev scenario
(localhost reverse proxy) and closes the spoofing path for the common
prod scenario (direct internet binding).

---

## Followup tickets (deferred work)

1. **CSRF middleware fail-closed** — issue a CSRF token at login so
   every authenticated session has one, then require it on every
   non-GET request unconditionally. Touches tests; not a 10-liner.
   Proposed task: `fix(auth): enforce CSRF for every authenticated session, not opt-in`.
2. **Argon2id migration** — same as seed.
   Proposed task: `refactor(auth): migrate password hashing to Argon2id (RFC 9106)`.
3. **TOTP second factor**. Proposed task:
   `feat(auth): TOTP 2FA enrollment + login challenge`.
4. **WebAuthn / Passkeys**. Proposed task:
   `feat(auth): WebAuthn passkey enrollment + login`.
5. **HIBP breach-corpus check** at setup + password change. Proposed
   task: `feat(auth): zxcvbn password meter + HIBP breach check`.
6. **Trusted-proxy CIDR config** so deployments behind real load
   balancers (not just localhost) can opt in to trusting forwarded
   headers. Proposed task:
   `feat(api): configurable trusted-proxy CIDRs for X-Forwarded-For`.
7. **Magic-link recovery via email** — replace the file-trigger
   recovery flow with email-based magic links.
   Proposed task: `feat(auth): magic-link account recovery via email`.
