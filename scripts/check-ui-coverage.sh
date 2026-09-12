#!/bin/sh
set -eu

# Enforce the UI coverage ratchet (#824).
#
# The UI suite runs in CI but its coverage was never checked: vitest.config.ts
# carried thresholds set to #824's targets, `npm run test` does not collect
# coverage, and nothing else looked. So 481 passing tests coexisted with
# branch coverage of 51%, and nothing would have said if it fell further.
#
# This compares the measured summary against scripts/ui-coverage-baseline.txt
# and fails when any metric drops below its floor. Raising the floor is the
# job of whoever improves coverage.

repo_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
summary=${UI_COVERAGE_SUMMARY:-$repo_dir/ui/coverage/coverage-summary.json}
baseline=$repo_dir/scripts/ui-coverage-baseline.txt

if [ ! -f "$summary" ]; then
    echo "::error::no coverage summary at $summary"
    echo "Run 'npm run test:coverage' in ui/ first."
    exit 1
fi

python3 - "$summary" "$baseline" <<'PY'
import json
import sys

summary_path, baseline_path = sys.argv[1], sys.argv[2]

with open(summary_path, encoding="utf-8") as handle:
    totals = json.load(handle)["total"]

floors = {}
targets = {}
with open(baseline_path, encoding="utf-8") as handle:
    for line in handle:
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        metric, floor, target = line.split()
        floors[metric] = float(floor)
        targets[metric] = float(target)

failures = []
raised = []
for metric, floor in sorted(floors.items()):
    measured = totals[metric]["pct"]
    target = targets[metric]
    status = "ok"
    if measured + 1e-9 < floor:
        status = "BELOW FLOOR"
        failures.append(f"{metric}: {measured:.2f}% is below the {floor:.1f}% floor")
    elif measured >= floor + 1.0:
        status = "floor can be raised"
        raised.append(f"{metric}: {measured:.2f}% (floor {floor:.1f}%)")
    print(f"  {metric:<11} {measured:6.2f}%  floor {floor:5.1f}%  target {target:5.1f}%  {status}")

if failures:
    print("::error::UI coverage regressed")
    for failure in failures:
        print(f"  {failure}")
    print("Add tests for what the change touched; the floor is never lowered.")
    sys.exit(1)

if raised:
    print("Coverage is a point or more above the floor. Raise it in "
          "scripts/ui-coverage-baseline.txt as part of this change:")
    for entry in raised:
        print(f"  {entry}")

print("OK: UI coverage is at or above every floor.")
PY
