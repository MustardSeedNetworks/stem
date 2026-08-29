# Development Guide

## Prerequisites

- Go 1.25.5+
- Node.js 25.2.1+
- npm 11.7.0+
- GCC/Clang 7.3.0+ (for C components)

## Setup

```bash
# Clone and setup
git clone <repo-url>
cd stem

# Backend
make build

# Frontend
cd ui
npm install
```

## Development

```bash
# Backend
make dev          # Run backend with hot reload
make test         # Run tests
make lint         # Run linters (Go + C)

# Frontend (cd ui/)
npm run dev       # Start dev server (port 5173)
npm run test      # Run unit tests
npm run test:e2e  # Run E2E tests
npm run lint      # Run Biome linter
```

## Project Structure

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architecture documentation.

## Common Tasks

| Task | Command |
| ------ | ------- |
| Build all | `make build` |
| Run tests | `make test` |
| Lint Go | `make lint-go` |
| Lint C | `make lint-c` |
| Clean | `make clean` |
| Dev server | `cd ui && npm run dev` |
| Reflect-path benchmark | `make c-bench` |
| Reflect-path perf gate | `make c-bench-compare` |

## Reflect-path performance gate

stem measures network performance, so a throughput regression in the reflect
path is a correctness regression. `tests/c/` covers behaviour, `tests/load/`
drives the HTTP API and the sanitizer targets cover memory safety — none of
them would notice a change that halved packets-per-second.

`make c-bench` builds `bench/bench_reflect.c` and prints packets-per-second for
each reflect case. `make c-bench-compare` is the gate, and runs in CI as the
**C Performance (reflect path)** job.

### How the comparison works

The gate is a **same-runner A/B**: it builds and measures both the baseline
revision (the merge base with `origin/main`) and the working tree, on the same
machine, in the same job, then compares rates per case.

Absolute packets-per-second is a property of the machine, not of the code — it
moves with CPU model, frequency scaling and whatever else shares the host — so
a figure recorded in the repository would false-positive as soon as the runner
changed. Measuring both revisions side by side cancels that out.

The default allowance is a **15% regression per case**, set against a measured
run-to-run spread of well under 1% for best-of-7 sampling.

### Re-baselining

There is nothing to re-baseline. Because the comparison is always against the
current merge base, a genuine improvement becomes the new baseline the moment
it merges — no committed constant, and so no step to forget.

If a change _deliberately_ trades reflect throughput for something else, say so
in the PR body and raise the allowance for that run:

```bash
BENCH_MAX_REGRESSION_PCT=25 make c-bench-compare
```

Do not remove the job or delete a case to make the gate pass.

### Failure modes

The gate fails closed. A benchmark that does not compile, does not run, or
emits no `BENCH` records fails the job — a perf gate that silently no-ops is
the `--passWithNoTests` failure mode with a longer feedback loop. A case
present in the baseline and missing from the current run is also a failure,
so a case cannot be quietly dropped.
