#!/usr/bin/env python3
"""Self-test for validate-touched.py.

The selection is the whole point of the script: a reverse dependency it fails
to find is a test that silently stops running in the inner loop, and a gate
missing from its table is one that never runs there at all. Each half is
checked on a synthetic module, and the gate table against scripts/ itself.
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "vt", Path(__file__).with_name("validate-touched.py")
)
vt = importlib.util.module_from_spec(spec)
sys.modules["vt"] = vt
spec.loader.exec_module(vt)

M = "example.com/m"


def pkg(
    path: str,
    imports: tuple[str, ...] = (),
    test_imports: tuple[str, ...] = (),
    cgo: bool = False,
):
    return vt.GoPackage(
        import_path=f"{M}/{path}",
        rel_dir=path,
        imports=frozenset(f"{M}/{i}" for i in imports) | {"fmt"},
        test_imports=frozenset(f"{M}/{i}" for i in test_imports),
        cgo=cgo,
    )


# leaf <- mid <- top; harness's tests import top; other is unrelated;
# dataplane links the C library and internal/api imports it; the schema
# generator imports internal/api.
PKGS = {
    p.import_path: p
    for p in (
        pkg("internal/leaf"),
        pkg("internal/mid", ("internal/leaf",)),
        pkg("cmd/top", ("internal/mid",)),
        pkg("internal/harness", test_imports=("cmd/top",)),
        pkg("internal/other"),
        pkg("internal/dataplane", cgo=True),
        pkg("internal/api", ("internal/dataplane",)),
        pkg("cmd/stem-schema", ("internal/api",)),
    )
}


def plan_for(touched: list[str], library_present: bool = True) -> vt.Plan:
    return vt.build_plan(touched, PKGS, "golangci-lint", "0.0.0", library_present)


def labels(plan: vt.Plan) -> list[str]:
    return [s.label for s in plan.steps]


def go_test_args(plan: vt.Plan) -> list[str]:
    steps = [s.cmd for s in plan.steps if s.label == "go test"]
    return steps[0][3:] if steps else []


def check_reverse_dependencies() -> list[str]:
    failures = []
    got = set(go_test_args(plan_for(["internal/leaf/leaf.go"])))
    want = {
        f"{M}/{p}"
        for p in ("internal/leaf", "internal/mid", "cmd/top", "internal/harness")
    }
    if got != want:
        failures.append(f"leaf change tested {sorted(got)}, want {sorted(want)}")
    lint = [
        s.cmd
        for s in plan_for(["internal/leaf/leaf.go"]).steps
        if s.label.startswith("golangci-lint")
    ]
    if lint != [["golangci-lint", "run", "--allow-parallel-runners", "./internal/leaf"]]:
        failures.append(f"leaf change linted {lint}, want only the touched package")
    if go_test_args(plan_for(["internal/other/x_test.go"])) != [f"{M}/internal/other"]:
        failures.append("a package nothing imports must test alone")
    windows = [
        s
        for s in plan_for(["internal/api/server.go"]).steps
        if s.label == "golangci-lint (GOOS=windows)"
    ]
    if not windows or windows[0].env != {"GOOS": "windows"}:
        failures.append("an internal/api change must also lint as Windows compiles it")
    return failures


def check_file_ownership() -> list[str]:
    by_dir = {p.rel_dir: p.import_path for p in PKGS.values()}
    cases = {
        "internal/mid/testdata/golden/a.json": f"{M}/internal/mid",
        "internal/mid/testdata/fixture/main.go": f"{M}/internal/mid",
        "internal/mid/assets/embedded.yaml": f"{M}/internal/mid",
        "internal/leaf/leaf_test.go": f"{M}/internal/leaf",
        "internal/nopkg/only_windows.go": None,
        "docs/README.md": None,
    }
    return [
        f"{path} owned by {vt.owning_package(path, by_dir)}, want {want}"
        for path, want in cases.items()
        if vt.owning_package(path, by_dir) != want
    ]


def check_scope_triggers() -> list[str]:
    failures = []
    docs = plan_for(["docs/ARCHITECTURE.md"])
    if any(s in {"go test", "golangci-lint", "vitest"} for s in labels(docs)):
        failures.append("a docs-only diff must run no Go or UI tests")
    if "markdownlint" not in labels(docs):
        failures.append("a docs diff must lint the markdown it changed")
    if go_test_args(plan_for(["go.mod"])) != ["./..."]:
        failures.append("go.mod must put every package in scope")
    ui = [s for s in plan_for(["ui/src/App.tsx"]).steps if s.label == "vitest"]
    if not ui or ui[0].cmd[-2:] != ["--run", "src/App.tsx"] or ui[0].cwd != "ui":
        failures.append(f"a UI source change must run vitest related on it, got {ui}")
    if "typecheck" not in labels(plan_for(["ui/src/App.tsx"])):
        failures.append("a UI source change must type-check: Vitest does not")
    whole = [s for s in plan_for(["ui/vitest.config.ts"]).steps if s.label == "vitest"]
    if not whole or whole[0].cmd != ["npm", "test"]:
        failures.append("a Vitest config change must run the whole suite")
    return failures


def check_c_library() -> list[str]:
    failures = []
    c = plan_for(["src/reflector/packet.c"])
    if labels(c)[:1] != ["dataplane"]:
        failures.append(f"a C source change must rebuild the library first, got {labels(c)}")
    if "c-test" not in labels(c):
        failures.append("a C source change must run the C tests")
    want = {f"{M}/internal/dataplane", f"{M}/internal/api", f"{M}/cmd/stem-schema"}
    got = go_test_args(c)
    if got[:1] != ["-a"] or set(got[1:]) != want:
        failures.append(
            f"a C source change tested {got}, want -a over the cgo packages "
            "and their importers (the build cache hashes neither archive nor headers)"
        )
    tests_only = labels(plan_for(["tests/c/test_pacing.c"]))
    if "c-test" not in tests_only or {"dataplane", "go test"} & set(tests_only):
        failures.append(f"a C test change must run the C tests and no Go, got {tests_only}")
    missing = plan_for(["internal/leaf/leaf.go"], library_present=False)
    if labels(missing)[:1] != ["dataplane"]:
        failures.append("Go tests on a host without the library must build it first")
    if "-a" in go_test_args(missing):
        failures.append("an unchanged library must not force a full rebuild")
    if "dataplane" in labels(plan_for(["docs/ARCHITECTURE.md"], library_present=False)):
        failures.append("a docs-only diff must not build the library")
    return failures


def check_gate_table() -> list[str]:
    """Every scripts/check-* is either selectable or named as not input-driven."""
    failures = []
    scripts = {
        f"scripts/{p.name}"
        for p in (vt.ROOT / "scripts").glob("check-*")
        if not p.name.endswith(".test.ts")
    }
    known = {g.script for g in vt.GATES} | vt.NOT_INPUT_DRIVEN
    for missing in sorted(scripts - known):
        failures.append(f"{missing} is in neither GATES nor NOT_INPUT_DRIVEN")
    for stale in sorted(known - scripts):
        failures.append(f"{stale} is listed but does not exist")
    gates = labels(plan_for(["scripts/json-casing-baseline.txt"]))
    if "gate json-casing" not in gates:
        failures.append("a gate's own baseline must select the gate")
    if "gate schema-drift" not in labels(plan_for(["internal/dataplane/dp.go"])):
        failures.append("a package the schema generator imports must select its gate")
    if "gate schema-drift" in labels(plan_for(["internal/leaf/leaf.go"])):
        failures.append("a package the schema generator does not import must not select it")
    help_cmds = [
        s.cmd for s in plan_for(["scripts/check-help-i18n.ts"]).steps
        if s.label == "gate help-i18n"
    ]
    if help_cmds != [
        ["node", "--test", "scripts/check-help-i18n.test.ts"],
        ["node", "scripts/check-help-i18n.ts"],
    ]:
        failures.append(f"a TypeScript gate must run its node:test self-test, got {help_cmds}")
    return failures


def main() -> int:
    failures = (
        check_reverse_dependencies()
        + check_file_ownership()
        + check_scope_triggers()
        + check_c_library()
        + check_gate_table()
    )
    for failure in failures:
        print(f"FAIL: {failure}")
    if failures:
        return 1
    print("validate-touched self-test: ok")
    return 0


if __name__ == "__main__":
    sys.exit(main())
