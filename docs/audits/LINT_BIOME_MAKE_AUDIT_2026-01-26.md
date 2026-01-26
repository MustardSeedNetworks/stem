# Lint, Biome, and Make Audit - Stem (2026-01-26)

## Go Lint (golangci-lint)
- Config: `.golangci.yml` identical to Seed/NiAC (maratori strict config).
- Strictness: **High**.
- Status: Strict and consistent across repos.

## Biome
- UI config is **strict**, but slightly more relaxed than Seed:
  - `noImplicitBoolean`, `noNegationElse`, `noNestedTernary`, `useAtIndex`, `useBlockStatements`, `useDefaultSwitchClause`, `useSimplifiedLogicExpression`, `noVoid`, `noForEach` set to `off`.
  - `useNamingConvention` allows `objectLiteralProperty` snake_case.
  - `App.tsx` and help content rules are relaxed via overrides.

## Make System
- Makefile uses `mk/*.mk`, includes C linting and formatting targets (Linux only).
- Go formatting uses `gofmt` (not `gofumpt`).
- Lint target includes Go + Biome. C lint requires `clang-tidy` with `compile_commands.json`.

## Findings
- If you want uniform strictness with Seed, consider:
  - Switching Go formatting to `gofumpt` (and align with Seed/NiAC).
  - Tightening Biome rules to match Seed UI config.
