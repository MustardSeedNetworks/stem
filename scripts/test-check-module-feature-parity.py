#!/usr/bin/env python3
"""Self-test for check-module-feature-parity.py against a throwaway tree."""

from __future__ import annotations

import importlib.util
import io
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("module_parity", HERE / "check-module-feature-parity.py")
gate = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(gate)

POLICY_GO = '''package license
const (
	FeatureReflector = "reflector"
	FeatureRFC2544   = "rfc2544"
	FeatureTSN       = "tsn"
)
'''

FEATURES_GO = '''package api
func featuresByTestType() map[string]string {
	return map[string]string{
		"reflect":            "",
		"rfc2544_throughput": license.FeatureRFC2544,
		"tsn_timing":         license.FeatureTSN,
	}
}
'''

CATALOG_TS = """export const MODULE_FEATURES = {
  '/tests/benchmark': { features: ['rfc2544'], i18nKey: 'benchmark' },
  '/tests/certify': { features: ['tsn'], i18nKey: 'certify' },
} as const satisfies Record<string, ModuleFeatures>;
"""


class Tree:
    def __init__(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        for rel, text in (
            (gate.POLICY_GO, POLICY_GO),
            (gate.FEATURES_GO, FEATURES_GO),
            (gate.CATALOG_TS, CATALOG_TS),
        ):
            path = self.root / rel
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(text, encoding="utf-8")

    def write(self, rel: str, text: str) -> None:
        (self.root / rel).write_text(text, encoding="utf-8")

    def run(self) -> tuple[int, str]:
        out = io.StringIO()
        return gate.run(self.root, out=out), out.getvalue()


class ModuleFeatureParityTest(unittest.TestCase):
    def test_agreeing_tree_passes(self) -> None:
        t = Tree()
        code, out = t.run()
        self.assertEqual(code, 0, out)
        self.assertIn("2 features agree", out)

    def test_the_free_reflector_is_not_priced(self) -> None:
        """"reflect" maps to the empty string, so reflector is a Free grant and
        a UI gate on it would pitch something the server runs anyway."""
        t = Tree()
        self.assertEqual(gate.priced_features(t.root), {"rfc2544", "tsn"})

    def test_ui_gate_on_an_unpriced_feature_fails(self) -> None:
        t = Tree()
        t.write(gate.CATALOG_TS, CATALOG_TS.replace("['tsn']", "['tsn', 'reflector']"))
        code, out = t.run()
        self.assertEqual(code, 1)
        self.assertIn("reflector", out)
        self.assertIn("no test type", out)

    def test_a_feature_id_outside_the_house_shape_still_fails(self) -> None:
        """A hyphen or a capital used to make the id invisible to the catalog
        regex, so the gate reported agreement it had never checked."""
        for bogus in ("wifi-survey", "RFC2544"):
            with self.subTest(bogus=bogus):
                t = Tree()
                t.write(gate.CATALOG_TS, CATALOG_TS.replace("['tsn']", f"['tsn', '{bogus}']"))
                code, out = t.run()
                self.assertEqual(code, 1, out)
                self.assertIn(bogus, out)
                self.assertIn("no test type", out)

    def test_a_priced_feature_no_module_names_fails(self) -> None:
        t = Tree()
        t.write(gate.CATALOG_TS, CATALOG_TS.replace("  '/tests/certify': { features: ['tsn'], i18nKey: 'certify' },\n", ""))
        code, out = t.run()
        self.assertEqual(code, 1)
        self.assertIn("tsn", out)
        self.assertIn("402s with no warning", out)

    def test_a_const_with_no_value_is_named_rather_than_ignored(self) -> None:
        t = Tree()
        t.write(gate.POLICY_GO, POLICY_GO.replace('\tFeatureTSN       = "tsn"\n', ""))
        with self.assertRaises(SystemExit) as raised:
            gate.priced_features(t.root)
        self.assertIn("FeatureTSN", str(raised.exception))


if __name__ == "__main__":
    sys.exit(unittest.main(verbosity=1))
