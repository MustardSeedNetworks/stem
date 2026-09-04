# Build Commands - The Stem

## Quick Reference

| Command | Purpose |
| --------- | --------- |
| `make build` | Build all (Go + C dataplane) |
| `make test` | Run all Go tests |
| `make lint` | Run golangci-lint + clang-tidy |
| `make clean` | Clean build artifacts |

## UI Commands

```bash
cd ui
npm install        # Install dependencies
npm run dev        # Development server (hot reload)
npm run build      # Production build
npm run lint       # Biome linting
npm run lint:fix   # Auto-fix lint issues
```

## Go Commands

```bash
go build ./...     # Build all packages
go test ./...      # Run all tests
go test -v ./internal/auth/...  # Test specific package
```

## Pre-commit

The project has pre-commit hooks that run:

- Secret detection (gitleaks)
- Go linting
- Sensitive file checks
