#!/usr/bin/env python3
"""check-module-feature-parity.py — UI module gate ⇄ the server's price list.

`internal/api/features.go` maps every registered test type to the catalog
feature it needs, and its own comment calls that map the single place a
capability is priced. The UI now gates a module on the same features
(ui/src/constants/moduleFeatures.ts, read by <ModuleGate>), and nothing held
the two together: a module could pitch a feature the server does not charge
for, or a module could be sold by the server and shown for free by the UI.

Two checks:

  A. Every feature the UI names is charged for by at least one test type in
     featuresByTestType(). A UI gate on an unpriced string is a pitch for
     something the server would have run anyway.
  B. Every feature the server charges for is named by some UI module. One the
     UI does not name is a module that reaches Start and 402s with no warning
     — which is the defect this gate exists to stop coming back.

It does not check that a given module's features are the ones ITS test types
need; the page-to-test-type mapping lives in the module forms.

Run locally: scripts/check-module-feature-parity.py
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

FEATURES_GO = "internal/api/features.go"
POLICY_GO = "internal/license/policy.go"
CATALOG_TS = "ui/src/constants/moduleFeatures.ts"

# `"rfc2544_throughput": license.FeatureRFC2544,` — the const, not a literal.
PRICED = re.compile(r'"[^"]+":\s*license\.(Feature\w+)\s*,')
# `FeatureRFC2544    = "rfc2544"` in the policy const block.
CONST = re.compile(r'\b(Feature\w+)\s*=\s*"([^"]*)"')
UI_BODY = re.compile(r"export const MODULE_FEATURES = \{(.*?)\n\} as const", re.S)
UI_ENTRY = re.compile(r"features:\s*\[([^\]]*)\]")
# Any quoted token, not `[a-z0-9_]+`: a feature id the house shape does not
# cover (a hyphen, a capital, a typo) must fail the set comparison below, not
# slip through the regex and leave the gate reporting agreement it never checked.
UI_FEATURE = re.compile(r"'([^']*)'")


def priced_features(root: Path) -> set[str]:
    """Feature ids the server charges for, resolved through policy.go's consts."""
    consts = dict(CONST.findall((root / POLICY_GO).read_text(encoding="utf-8")))
    names = set(PRICED.findall((root / FEATURES_GO).read_text(encoding="utf-8")))
    unknown = sorted(name for name in names if name not in consts)
    if unknown:
        raise SystemExit(f"{FEATURES_GO}: no value in {POLICY_GO} for {', '.join(unknown)}")
    return {consts[name] for name in names}


def ui_features(root: Path) -> set[str]:
    body = UI_BODY.search((root / CATALOG_TS).read_text(encoding="utf-8"))
    if not body:
        raise SystemExit(f"{CATALOG_TS}: MODULE_FEATURES not found — is the regex stale?")
    features: set[str] = set()
    for entry in UI_ENTRY.findall(body.group(1)):
        features.update(UI_FEATURE.findall(entry))
    return features


def run(root: Path, out=sys.stdout) -> int:
    priced = priced_features(root)
    named = ui_features(root)

    failures: list[str] = []
    for feature in sorted(named - priced):
        failures.append(f"{feature}: gated in the UI but no test type in {FEATURES_GO} is sold under it")
    for feature in sorted(priced - named):
        failures.append(f"{feature}: sold by {FEATURES_GO} but no UI module names it, so it 402s with no warning")

    if failures:
        print("::error::the UI's module gates disagree with the server's price list:", file=out)
        for line in failures:
            print(f"  {line}", file=out)
        return 1

    print(f"Module-feature parity: {len(named)} features agree with {FEATURES_GO}.", file=out)
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parent.parent)
    return run(parser.parse_args().root)


if __name__ == "__main__":
    sys.exit(main())
