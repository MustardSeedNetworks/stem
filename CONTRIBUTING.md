# Contributing to The Stem

Thank you for your interest in contributing to The Stem! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions.

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 25+
- GCC 7.3.0+ (for C dataplane)
- Git

### Development Setup

```bash
# Clone the repository
git clone https://github.com/krisarmstrong/stem.git
cd stem

# Install Go dependencies
go mod download

# Install frontend dependencies
cd ui && npm install && cd ..

# Build everything
make build

# Run tests
make test

# Run linting
make lint
```

## Development Workflow

### Branch Naming

Use descriptive branch names with prefixes:

- `feat/` - New features (e.g., `feat/y1564-support`)
- `fix/` - Bug fixes (e.g., `fix/latency-calculation`)
- `docs/` - Documentation (e.g., `docs/api-reference`)
- `chore/` - Maintenance (e.g., `chore/update-deps`)
- `refactor/` - Code refactoring (e.g., `refactor/packet-engine`)

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/). All commits must follow this format:

```
type(scope): description

[optional body]

[optional footer]
```

#### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `ci` - CI/CD changes
- `build` - Build system changes

#### Examples

```
feat(benchmark): add RFC 2544 back-to-back test
fix(reflector): resolve packet drop on high load
docs: update installation instructions
chore(deps): upgrade Go to 1.25.5
```

### Pull Request Process

1. **Create an issue first** - Discuss the change before implementing
2. **Fork and branch** - Create a feature branch from `main`
3. **Write tests** - Ensure adequate test coverage
4. **Update docs** - Update relevant documentation
5. **Run checks** - Ensure all tests and linting pass
6. **Submit PR** - Reference the related issue

## Code Standards

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write table-driven tests
- Document exported functions

### C

- Use C23 standard
- Follow `.clang-format` configuration
- Run `clang-tidy` before committing
- Memory safety is critical - no buffer overflows

### TypeScript/React

- TypeScript only - NO JavaScript
- Follow the existing code style
- Use Biome for linting and formatting
- Write unit tests with Vitest

## Testing

### Running Tests

```bash
# All tests
make test

# Go tests with coverage
go test -coverprofile=coverage.out ./...

# Go tests with race detection
go test -race ./...

# Frontend tests
cd ui && npm test
```

### Test Requirements

- Unit tests for business logic
- Integration tests for API endpoints
- Standards compliance tests (RFC 2544, Y.1564)
- Aim for >80% coverage on new code

## Quality Gates

Commits + PRs are gated by automation, all configured in-repo:

- **Commit message format** — enforced by `commitlint.config.js`
  (conventional commits + per-project scope list). Husky runs it on
  every commit via `.husky/commit-msg`.
- **Pre-commit checks** — `.pre-commit-config.yaml` runs on
  `git commit`: secret detection (gitleaks), formatting (biome,
  clang-format), shellcheck on `.sh` scripts, large-file checks,
  schema validation.
- **CI required checks** — `CI Complete`, `License Compliance Check`,
  and CodeQL (`Analyze (go)` + `Analyze (javascript-typescript)`)
  must pass before merge.
- **Coverage floor** — `ui/vitest.config.ts` enforces a per-project
  anti-regression threshold; the Go test job in CI enforces its own.

If pre-commit blocks a commit, fix the issue locally — **do not** use
`--no-verify` (forbidden by CLAUDE.md).

## Reporting Security Vulnerabilities

**Do not open a public issue for a security vulnerability.** See
[SECURITY.md](SECURITY.md) for the private disclosure channels
(GitHub Security Advisories or email).

## Questions?

- Check existing issues and documentation
- Open a discussion for general questions
- File an issue for bugs or features

Thank you for contributing!
