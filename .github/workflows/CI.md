# CI/CD Pipeline

The CI pipeline runs on every push and PR. **All checks must pass.**

Every job except `changes` is gated on the `changes` job's path filters, so a
docs-only PR does not pay for a Go build. `ci-complete` is the single required
status check — it depends on every other job, so adding a job to `ci.yml`
without adding it to `ci-complete`'s `needs:` list makes that job advisory.

## GitHub Actions Workflows

### ci.yml - Main CI Pipeline

| Job | Description | Checks |
| --- | --- | --- |
| `changes` | Path filtering | Decides which downstream jobs run |
| `build-ui` | Shared UI build | Builds `internal/api/ui/` once for `backend`/`race` to download |
| `backend` | Go checks | lint, vet, staticcheck, fmt, tests, coverage floor |
| `race` | Go race detector | `go test -race`, split from `backend` so it fails distinctly |
| `frontend` | React/TS checks | tsc typecheck, Biome, Vite build, Vitest, Storybook build |
| `c-lint` | C dataplane lint (C23) | clang-format, clang-tidy |
| `dataplane-safety` | C memory safety | ASAN + fuzz targets |
| `security` | Security scans | govulncheck (hard gate), gosec, npm audit, gitleaks, Trivy |
| `semgrep` | SAST | Semgrep rules |
| `quality` | Code quality gates | banned vocabulary, file size ratchet, output escaping, sensitive files |
| `workflow-lint` | Workflow static analysis | actionlint; zizmor (blocks on High) |
| `i18n` | Internationalization | Catalog completeness, no translated standard terms |
| `docs` | Documentation | Markdown lint (blocking, scoped to changed files) |
| `build` | Build verification | Multi-arch binaries with full ldflags |
| `darwin-compile-check` | macOS cross-compile | arm64 compile only |
| `e2e` | Browser tests | Playwright, chromium + webkit (browsers cached) |
| `ci-complete` | Aggregate gate | The required status check |

### Other Workflows

| Workflow | Purpose |
| --- | --- |
| `codeql.yml` | CodeQL security analysis (Go, JS/TS) |
| `dead-code.yml` | Weekly dead code detection |
| `docs-link-check.yml` | Weekly external link check (split out of `ci.yml`) |
| `label-sync.yml` | Sync label definitions |
| `labeler.yml` | Auto-label PRs and issues |
| `license-check.yml` | Verify dependency licenses |
| `pr-body-lint.yml` | Enforce the PR body template |
| `release-please.yml` | Automated version management and release PRs |
| `release.yml` | goreleaser release builds, signing, provenance |
| `scorecard.yml` | OpenSSF Scorecard |
| `title-lint.yml` | Lint PR and issue titles |
| `todo-tracker.yml` | Weekly TODO tracking |

## Workflow security

`workflow-lint` runs two scanners over `.github/workflows/` itself:

- **actionlint** — syntax, expression and shell errors inside `run:` blocks.
  It catches things a plain YAML parse does not, including duplicate `with:`
  keys, which `yaml.safe_load` accepts silently by keeping the last one.
  `SC2129` is ignored as a pure style preference; every correctness rule stays on.
- **zizmor** (pinned 1.29.0) — Actions security scanner, run against the whole
  `.github/workflows/` directory. **Blocks on High findings.** The repo sits
  at zero High. One finding elsewhere in the directory (`release-please.yml`)
  survived review and carries a `# zizmor: ignore[...]` comment with the
  reasoning inline; anything else that reaches High fails the build.
  Low/Informational are reported but not yet enforced.

Permissions follow least privilege: workflows declare `permissions: {}` (or
`contents: read`) at the top level and grant scopes per job. A new job that
needs a write scope declares it on the job, never workflow-wide. `release.yml`
deliberately runs without npm caching, because its output is published and
attested and a restored cache entry could land inside a signed artifact; it
opts out by passing `cache: ""` to the `setup-node` composite action.

## The Node.js pin lives in one file

`.nvmrc` is the single source of truth for the Node version. Every workflow that
needs Node uses `./.github/actions/setup-node`, and that composite reads
`.nvmrc` via `node-version-file` — it has **no `node-version` input**, so no
caller can override it and no second copy of the version can exist.

That input used to default to a literal, and it drifted: Renovate bumps the
manifests it can see and cannot see a default buried inside a composite, so CI
ran 26.7.0 against manifests demanding 26.8.1 and logged EBADENGINE on every
job for weeks. "Must stay in step" was the previous rule here, and a rule that
depends on someone remembering is not a mechanism.

The remaining pair that can disagree is `.nvmrc` and the `engines` field in
`package.json`. Making that a hard failure (`engine-strict=true`) is the obvious
next step and is deliberately **not** taken yet: Homebrew's newest `node` is
26.7.0, so 26.8.1 is not installable through the fleet's normal channel, and
turning the mismatch fatal would block local development in all four repos. See
the linked issue.

The npm version is still declared in the composite; `packageManager` in
`package.json` is what `engines` checks it against.

## CI Must Pass Before Merge

`main` is protected. Push a feature branch, open a PR, and let CI gate it:

```bash
gh pr create --fill
gh pr merge --auto --squash --delete-branch
```

Fix issues locally first:

```bash
make all       # Full local verification
make verify    # lint, test, security, build
make test-e2e  # Playwright E2E (requires the backend running)
```

## Running CI Checks Locally

### Backend

```bash
make lint-go           # golangci-lint v2.13.1
make test-backend      # Go tests
make test-coverage     # Coverage report
make security-backend  # gosec + govulncheck
```

### Frontend

```bash
make lint-frontend     # Biome
make test-frontend     # Vitest
make ui                # Vite build into internal/api/ui/
```

### C dataplane

```bash
make lint-c            # clang-tidy (Linux only)
make c-test            # C unit tests
make c-test-asan       # AddressSanitizer + UBSan
make c-fuzz            # libFuzzer + ASAN over the packet parser
```

### Security

```bash
make security          # All security scans
make security-secrets  # gitleaks
make security-trivy    # Trivy
```

### Workflows

```bash
actionlint -ignore 'SC2129'
zizmor --min-severity high .github/workflows/
```
