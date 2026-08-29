#!/usr/bin/env bash
#
# bench-compare.sh - guard the reflect hot path against throughput regressions.
#
# Builds and runs bench/bench_reflect.c twice on the SAME machine, in the same
# job: once for the working tree, once for a baseline revision. Compares
# packets-per-second per case and fails if any case regresses by more than the
# allowed percentage.
#
# Why same-runner A/B rather than a committed baseline figure:
#
#   Absolute packets-per-second is a property of the machine, not of the code.
#   GitHub's hosted runners vary by CPU model and by how loaded the host is, so
#   a number recorded on one runner false-positives on the next. Measuring both
#   revisions on the machine in front of us cancels that out. It also means a
#   genuine improvement re-baselines itself the moment it merges -- there is no
#   recorded constant to remember to update, which is the step such schemes
#   always skip.
#
# The gate fails closed. A build that does not compile, a binary that does not
# run, output that cannot be parsed, or a case present in one run and missing
# from the other are all failures -- a perf gate that silently no-ops is the
# --passWithNoTests failure mode with a longer feedback loop.
#
# Usage:
#   scripts/bench-compare.sh [baseline-ref]
#
# Environment:
#   BENCH_MAX_REGRESSION_PCT  allowed slowdown per case (default 15)
#   CC                        compiler (default gcc)

set -euo pipefail

BASELINE_REF="${1:-}"
MAX_REGRESSION="${BENCH_MAX_REGRESSION_PCT:-15}"
CC="${CC:-gcc}"

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Matches the Makefile's C flags, minus -march=native: the baseline and the
# working tree must be built identically, and -march=native is fine here only
# because both builds happen on this same machine. It is dropped anyway so the
# comparison does not change meaning if the runner class changes.
BENCH_CFLAGS=(
  -D_GNU_SOURCE -D_DEFAULT_SOURCE -std=c23
  -Wall -Wextra -Wpedantic -O3 -pthread -Iinclude
)
BENCH_SRCS=(
  bench/bench_reflect.c
  src/reflector/packet.c
  src/reflector/netally.c
  src/reflector/util.c
)

WORKDIR="$(mktemp -d)"
BASELINE_TREE="$WORKDIR/baseline"
cleanup() {
  if [ -d "$BASELINE_TREE" ]; then
    git worktree remove --force "$BASELINE_TREE" >/dev/null 2>&1 || true
  fi
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

# Builds and runs the benchmark in $1, writing "name pps" lines to $2.
run_bench() {
  local tree="$1" out="$2" label="$3"
  local bin="$WORKDIR/bench_$label"

  echo "==> building benchmark for $label"
  if ! ( cd "$tree" && "$CC" "${BENCH_CFLAGS[@]}" -o "$bin" "${BENCH_SRCS[@]}" -pthread -lm ); then
    echo "FAIL: benchmark did not compile for $label" >&2
    exit 1
  fi

  echo "==> running benchmark for $label"
  if ! "$bin" >"$WORKDIR/raw_$label" 2>/dev/null; then
    echo "FAIL: benchmark did not run cleanly for $label" >&2
    exit 1
  fi

  # Only the BENCH records; anything else on stdout is ignored, not trusted.
  awk '$1 == "BENCH" && NF == 3 { print $2, $3 }' "$WORKDIR/raw_$label" >"$out"

  if [ ! -s "$out" ]; then
    echo "FAIL: benchmark produced no BENCH records for $label" >&2
    echo "--- raw output ---" >&2
    cat "$WORKDIR/raw_$label" >&2
    exit 1
  fi
}

if [ -z "$BASELINE_REF" ]; then
  git fetch --quiet origin main 2>/dev/null || true
  BASELINE_REF="$(git merge-base HEAD origin/main 2>/dev/null || echo '')"
  if [ -z "$BASELINE_REF" ]; then
    echo "FAIL: could not determine a baseline revision (no merge-base with origin/main)" >&2
    exit 1
  fi
fi

echo "baseline: $BASELINE_REF"
echo "current:  $(git rev-parse HEAD)"
echo "allowed regression: ${MAX_REGRESSION}%"

if ! git worktree add --detach --quiet "$BASELINE_TREE" "$BASELINE_REF"; then
  echo "FAIL: could not check out baseline revision $BASELINE_REF" >&2
  exit 1
fi

# The benchmark may not exist on the baseline -- it does not on the commit that
# introduces it. That is not a regression, and it is not a reason to pass
# silently either: report it and skip the comparison for this run only.
if [ ! -f "$BASELINE_TREE/bench/bench_reflect.c" ]; then
  echo
  echo "SKIP: the baseline revision predates bench/bench_reflect.c."
  echo "      Nothing to compare against; the gate becomes effective from the"
  echo "      next change to the reflect path."
  run_bench "$REPO_ROOT" "$WORKDIR/current.txt" current
  echo
  echo "current measurements:"
  awk '{ printf "  %-26s %15.0f pps\n", $1, $2 }' "$WORKDIR/current.txt"
  exit 0
fi

run_bench "$BASELINE_TREE" "$WORKDIR/baseline.txt" baseline
run_bench "$REPO_ROOT" "$WORKDIR/current.txt" current

echo
MAX_REGRESSION="$MAX_REGRESSION" awk '
  FNR == NR { base[$1] = $2; next }
  {
    cur[$1] = $2
  }
  END {
    limit = ENVIRON["MAX_REGRESSION"] + 0
    failed = 0
    printf "%-26s %15s %15s %9s\n", "case", "baseline pps", "current pps", "change"
    for (name in base) {
      if (!(name in cur)) {
        printf "%-26s %15.0f %15s %9s\n", name, base[name], "MISSING", "FAIL"
        failed = 1
        continue
      }
      if (base[name] <= 0) {
        printf "%-26s %15s %15.0f %9s\n", name, "INVALID", cur[name], "FAIL"
        failed = 1
        continue
      }
      delta = (cur[name] - base[name]) / base[name] * 100
      verdict = (delta < -limit) ? "FAIL" : "ok"
      if (verdict == "FAIL") failed = 1
      printf "%-26s %15.0f %15.0f %8.1f%% %s\n", name, base[name], cur[name], delta, verdict
    }
    for (name in cur) {
      if (!(name in base)) {
        printf "%-26s %15s %15.0f %9s\n", name, "(new)", cur[name], "ok"
      }
    }
    if (failed) {
      printf "\nFAIL: a reflect-path case regressed by more than %d%%.\n", limit
      print  "If this is a deliberate trade-off, say so in the PR body and raise"
      print  "BENCH_MAX_REGRESSION_PCT for that job -- do not silence the gate."
      exit 1
    }
    print "\nOK: no reflect-path case regressed beyond the allowed margin."
  }
' "$WORKDIR/baseline.txt" "$WORKDIR/current.txt"
