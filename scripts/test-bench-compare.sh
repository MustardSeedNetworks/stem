#!/usr/bin/env bash
# test-bench-compare.sh — proves bench-compare.sh only measures when the code
# under measurement actually changed.
#
# `merge_group` forces every path filter in ci.yml to `true`, so the reflect
# benchmark runs on every queued PR whatever it touched. Run 34687252273
# measured two trees that differ by ten lines of package-lock.json and reported
# reflect_mode_mac_ip_v4 at -25.0% and reflect_mode_mac_v4 at -27.3%, which
# ejected a bot PR from the merge queue. Identical C cannot regress: the only
# thing such a run can measure is the runner (#1198).
#
# Both directions matter, so both are exercised here against the real script:
# identical binaries must short-circuit without measuring, and a changed C
# source must still be measured.
set -euo pipefail

cd "$(dirname "$0")/.."
readonly REPO="$PWD"
readonly SCRIPT="$REPO/scripts/bench-compare.sh"

CC="${CC:-gcc}"
export CC

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# A throwaway repository holding the real benchmark sources, so a change to
# either the script or the benchmark is genuinely exercised.
repo="$work/repo"
mkdir -p "$repo/scripts"
git -C "$repo" init --quiet
git archive --format=tar HEAD bench src/reflector include | tar -x -C "$repo"
cp "$SCRIPT" "$repo/scripts/bench-compare.sh"
git -C "$repo" config user.email stem@example.invalid
git -C "$repo" config user.name "stem tests"
git -C "$repo" add -A
git -C "$repo" commit --quiet -m "baseline"
BASELINE=$(git -C "$repo" rev-parse HEAD)
readonly BASELINE

run_gate() {
  # One round per side: this test is about which path the script takes, not
  # about the estimator, and 30 rounds would put a minute on every CI run.
  ( cd "$repo" && BENCH_RUNS=1 BENCH_MAX_REGRESSION_PCT=100 \
      ./scripts/bench-compare.sh "$BASELINE" ) >"$1" 2>&1
}

# --- 1. a change that cannot touch the reflect path --------------------------
echo "package-lock-only change:"
echo '{ "lockfileVersion": 3 }' >"$repo/package-lock.json"
git -C "$repo" add -A
git -C "$repo" commit --quiet -m "lockfile only"

if ! run_gate "$work/out_identical"; then
  echo "FAIL: the gate failed on a change that does not touch the reflect path" >&2
  sed 's/^/  /' "$work/out_identical" >&2
  exit 1
fi
if ! grep -q "IDENTICAL" "$work/out_identical"; then
  echo "FAIL: identical benchmark binaries were not recognized as identical" >&2
  sed 's/^/  /' "$work/out_identical" >&2
  exit 1
fi
if grep -q "measuring" "$work/out_identical"; then
  echo "FAIL: the gate measured two identical binaries — that samples the" >&2
  echo "      runner, not the code, and is what false-failed #1187" >&2
  sed 's/^/  /' "$work/out_identical" >&2
  exit 1
fi
sed 's/^/  /' "$work/out_identical"

# --- 2. a real change to the measured code -----------------------------------
# The short-circuit must not swallow the case the gate exists for: a different
# reflect path is still measured and reported.
echo
echo "reflect-path change:"
printf '\nint stem_bench_gate_probe(void) { return 1; }\n' >>"$repo/src/reflector/util.c"
git -C "$repo" add -A
git -C "$repo" commit --quiet -m "touch the reflect path"

if ! run_gate "$work/out_changed"; then
  echo "FAIL: the gate did not complete on a changed reflect path" >&2
  sed 's/^/  /' "$work/out_changed" >&2
  exit 1
fi
if grep -q "IDENTICAL" "$work/out_changed"; then
  echo "FAIL: a changed reflect path was short-circuited as identical" >&2
  sed 's/^/  /' "$work/out_changed" >&2
  exit 1
fi
if ! grep -q "measuring" "$work/out_changed"; then
  echo "FAIL: a changed reflect path was not measured" >&2
  sed 's/^/  /' "$work/out_changed" >&2
  exit 1
fi
sed 's/^/  /' "$work/out_changed"

echo
echo "PASS: bench-compare.sh measures a changed reflect path and only that."
