# Code index

## Authentication state

The administrator account (password hash, TOTP secret, passkeys, passkey user
handle) is kept by foundation's `pkg/credstore` in `<datadir>/auth`, sealed with
AES-256-GCM under a key of its own. `internal/auth.Manager` holds a working copy
for the login path; `NewManagerFromStore` loads it, and every change
(`ReplaceCredential`, `VerifyAndEnableTOTP`, `DisableTOTP`, `SavePasskey`) is
committed to the store before the copy changes. `internal/api/admin_state.go`
opens the store, refuses to start on one it cannot read (never falling back to
unclaimed), seeds it once from `STEM_AUTH_USERNAME`/`STEM_AUTH_PASSWORD`, and
does not read an old `credentials.json`. Recovery replaces the password and
clears the TOTP secret and passkeys in one write.
