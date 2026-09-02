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
#   BENCH_RUNS                measurements per side, best taken (default 3)
#   CC                        compiler (default gcc)

set -euo pipefail

BASELINE_REF="${1:-}"
MAX_REGRESSION="${BENCH_MAX_REGRESSION_PCT:-15}"
BENCH_RUNS="${BENCH_RUNS:-3}"

# Cases that report but do not block.
#
# reflect_inplace_v4 cannot currently measure itself. Evidence, all on one idle
# machine with byte-identical code on both sides (#965):
#
#   - the same binary run ten times produced 93.8M-175.3M pps, a 1.87x spread
#   - the slow mode survives all 7 of the benchmark's own internal REPEATS, so
#     it is fixed at process start rather than varying per repeat
#   - it tracks process layout: padding the environment moved the result
#     between 110M and 175M with nothing else changed
#   - aligning the frame buffer to 64 bytes did not fix it, so the buffer's
#     own alignment is not the cause
#
# Interleaving fixed the six other cases -- they now agree within 0.5% -- but
# not this one. Blocking on a case that cannot measure would fail roughly one
# PR in three for no reason, and a gate people re-run until it goes green is a
# gate nobody reads. It stays measured and printed; it just does not fail the
# build until #965 explains it.
#
# This list should stay empty. Adding to it needs the same standard: evidence
# that the case cannot measure, not that a change made it slower.
ADVISORY_CASES="${ADVISORY_CASES:-reflect_inplace_v4}"
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

# Builds the benchmark in $1 to a binary labelled $2.
build_bench() {
  local tree="$1" label="$2"
  echo "==> building benchmark for $label"
  if ! ( cd "$tree" && "$CC" "${BENCH_CFLAGS[@]}" -o "$WORKDIR/bench_$label" \
           "${BENCH_SRCS[@]}" -pthread -lm ); then
    echo "FAIL: benchmark did not compile for $label" >&2
    exit 1
  fi
}

# Runs the $1 binary once, appending "name pps" records to its sample file.
sample_bench() {
  local label="$1" round="$2"
  local raw="$WORKDIR/raw_${label}_${round}"
  if ! "$WORKDIR/bench_$label" >"$raw" 2>/dev/null; then
    echo "FAIL: benchmark did not run cleanly for $label (round $round)" >&2
    exit 1
  fi
  # Only the BENCH records; anything else on stdout is ignored, not trusted.
  awk '$1 == "BENCH" && NF == 3 { print $2, $3 }' "$raw" >>"$WORKDIR/all_$label"
}

# Measures both sides, INTERLEAVED, and writes the best result per case.
#
# Two separate problems make a single measurement per side unusable, and they
# need different answers (#965):
#
#   Within a side: reflect_inplace_v4 is bimodal on some hosts. The same
#   binary, run ten times back to back on an idle machine, produced 93.8M-
#   175.3M pps -- a 1.87x spread with no code difference at all. Best-of-N
#   fixes that, because interference only ever makes a run slower, so the
#   fastest observation is the closest to what the code can do.
#
#   Across sides: the machine drifts during the job. Measuring all of the
#   baseline and then all of the current made that drift look like a
#   regression -- one observed round had all six other cases "regress" by
#   16-23% simultaneously, which no code change can do. Interleaving puts both
#   sides in the same time window so a slow patch hits them equally, and the
#   order alternates each round so neither side is permanently second.
measure_both() {
  : >"$WORKDIR/all_baseline"
  : >"$WORKDIR/all_current"

  echo "==> measuring, best of $BENCH_RUNS interleaved rounds"
  for round in $(seq 1 "$BENCH_RUNS"); do
    if [ $((round % 2)) -eq 1 ]; then
      sample_bench baseline "$round"
      sample_bench current "$round"
    else
      sample_bench current "$round"
      sample_bench baseline "$round"
    fi
  done

  for label in baseline current; do
    awk '{ if (!($1 in best) || $2 > best[$1]) best[$1] = $2 }
         END { for (name in best) print name, best[name] }' \
      "$WORKDIR/all_$label" | sort >"$WORKDIR/$label.txt"

    if [ ! -s "$WORKDIR/$label.txt" ]; then
      echo "FAIL: benchmark produced no BENCH records for $label" >&2
      echo "--- raw output ---" >&2
      cat "$WORKDIR/raw_${label}_1" >&2
      exit 1
    fi
  done
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
  build_bench "$REPO_ROOT" current
  : >"$WORKDIR/all_current"
  for round in $(seq 1 "$BENCH_RUNS"); do sample_bench current "$round"; done
  awk '{ if (!($1 in best) || $2 > best[$1]) best[$1] = $2 }
       END { for (name in best) print name, best[name] }' \
    "$WORKDIR/all_current" | sort >"$WORKDIR/current.txt"
  echo
  echo "current measurements:"
  awk '{ printf "  %-26s %15.0f pps\n", $1, $2 }' "$WORKDIR/current.txt"
  exit 0
fi

build_bench "$BASELINE_TREE" baseline
build_bench "$REPO_ROOT" current
measure_both

echo
MAX_REGRESSION="$MAX_REGRESSION" ADVISORY_CASES="$ADVISORY_CASES" awk '
  FNR == NR { base[$1] = $2; next }
  {
    cur[$1] = $2
  }
  END {
    limit = ENVIRON["MAX_REGRESSION"] + 0
    split(ENVIRON["ADVISORY_CASES"], adv, " ")
    for (i in adv) if (adv[i] != "") advisory[adv[i]] = 1
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
      regressed = (delta < -limit)
      if (regressed && (name in advisory)) {
        verdict = "ADVISORY"
        advised = 1
      } else {
        verdict = regressed ? "FAIL" : "ok"
        if (regressed) failed = 1
      }
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
    if (advised) {
      printf "\nADVISORY: a case listed in ADVISORY_CASES regressed. Not failing the\n"
      print  "build, because that case cannot currently measure itself reliably --"
      print  "see the comment above ADVISORY_CASES and #965. Every other case is"
      print  "still blocking."
    }
    print "\nOK: no blocking reflect-path case regressed beyond the allowed margin."
  }
' "$WORKDIR/baseline.txt" "$WORKDIR/current.txt"
