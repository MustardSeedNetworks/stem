# GitHub Issues - Stem (2026-01-26)

## Security / Test Data
- Title: Load tests default to weak password value
  Labels: security, tests
  Files: tests/load/api.js:38, tests/load/full.js:43
  Body: Load tests fall back to "password" when STEM_PASS is not set. Require explicit env var or use a stronger placeholder to avoid accidental weak defaults.

## Lint / Hygiene
- Title: Handle errors returned by os.Setenv
  Labels: lint, backend
  File: cmd/stem/main.go:1183
  Body: `gosec` G104 flags unhandled errors from `os.Setenv` calls in `ensureStemCredentials`.

- Title: Fix nolint directive format
  Labels: lint, backend
  File: internal/version/version.go:5
  Body: `nolintlint` requires `//nolint:gochecknoglobals` with no leading space.

- Title: Add t.Parallel in subtests or disable rule
  Labels: lint, tests
  File: internal/license/license_test.go:692
  Body: `tparallel` warns subtests in `TestGenerateLicenseKeyErrors` don't call `t.Parallel()`.
