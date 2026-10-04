#!/usr/bin/env python3
"""validate-touched.py — validate what the branch touched, not the whole tree.

The inner loop. `make test` runs every Go package under -race and the whole
Vitest suite; a branch that edits one leaf package needs that package,
everything in the module that imports it (directly, through another package,
or from its tests), and the gates that read the files it changed. `make test`
still runs once before the PR; this is what runs between edits.

The touched set is every path that differs between the merge base with
origin/main and the working tree, committed or not, plus untracked files.
Renames count as a delete and an add, so the package a file left is validated
too.

The C dataplane is linked into the cgo packages as build/libreflector.a, and
Go's build cache hashes neither that archive nor the headers under include/
(D-STEM-25). A change to src/ or include/ therefore rebuilds the library, runs
the C tests, and puts the cgo packages in scope with `go test -a`.

Every command is printed before it runs, and a failing step does not stop the
others: the exit status is non-zero when any step failed.

Run locally: make validate-touched   (or scripts/validate-touched.py --dry-run)
"""

from __future__ import annotations

import argparse
import fnmatch
import os
import subprocess
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path, PurePosixPath

ROOT = Path(__file__).resolve().parent.parent

# Fixed, not a flag: whatever a caller passes reaches the commands this runs.
BASE = "origin/main"

# A change to any of these can alter how every package builds or lints.
WHOLE_MODULE = ("go.mod", "go.sum", ".golangci.yml")

# The C sources and headers linked into the cgo packages.
C_LIBRARY = ("src/*", "include/*")

# C sources and their tests; `make c-test` covers both.
C_TESTED = (*C_LIBRARY, "tests/c/*")

# What the cgo packages link against; `make dataplane` writes it.
DATAPLANE_LIBRARY = "build/libreflector.a"

# A change to any of these can alter how every UI test resolves or runs.
WHOLE_UI = (
    "ui/package.json",
    "ui/package-lock.json",
    "ui/vite.config.ts",
    "ui/vitest.config.ts",
    "ui/tsconfig*.json",
    "ui/src/test/*",
)

UI_SOURCE_SUFFIXES = {".ts", ".tsx", ".css", ".json"}

# The same entry point ui/package.json's `test` script uses.
VITEST = ["node", "--disable-warning=DEP0205", "./node_modules/vitest/vitest.mjs"]

# The locale catalogues live with the Go i18n package; the UI imports them as
# @locales.
UI_INPUTS = ("ui/src/*", "internal/i18n/locales/*")


def matches_any(path: str, patterns: tuple[str, ...]) -> bool:
    return any(fnmatch.fnmatchcase(path, p) for p in patterns)


@dataclass(frozen=True)
class Gate:
    """A scripts/check-* gate and the paths it reads.

    Patterns are fnmatch patterns over repo-relative paths, where `*` also
    crosses `/`. The gate's own files (the script, its self-test, its baseline
    or allow-list: `scripts/<name>-*`) are inputs of every gate implicitly.
    """

    script: str
    inputs: tuple[str, ...]
    # A generator the gate runs with `go run`: the gate also reads every
    # in-module package that generator imports, and links the C library.
    generator: str | None = None

    @property
    def name(self) -> str:
        return PurePosixPath(self.script).stem.removeprefix("check-")

    def own_files(self) -> tuple[str, ...]:
        return (
            f"scripts/check-{self.name}.*",
            f"scripts/test-check-{self.name}.*",
            f"scripts/{self.name}-*",
        )

    def selected(self, touched: list[str], affected: set[str]) -> bool:
        if self.generator and self.generator in affected:
            return True
        return any(matches_any(f, self.inputs + self.own_files()) for f in touched)

    def commands(self) -> list[list[str]]:
        scripts = ROOT / "scripts"
        if self.script.endswith(".ts"):
            self_test = scripts / f"check-{self.name}.test.ts"
            run = ["node", self.script]
            test = ["node", "--test", f"scripts/{self_test.name}"]
        else:
            suffix = PurePosixPath(self.script).suffix
            self_test = scripts / f"test-check-{self.name}{suffix}"
            interpreter = ["python3"] if suffix == ".py" else []
            run = [*interpreter, f"./{self.script}"]
            test = [*interpreter, f"./scripts/{self_test.name}"]
        return [test, run] if self_test.exists() else [run]


GATES = (
    Gate("scripts/check-aria-label-i18n.sh", ("ui/src/*",)),
    Gate("scripts/check-banned-vocabulary.py", ("*",)),
    Gate("scripts/check-component-variants.py", ("ui/src/*",)),
    Gate("scripts/check-file-size.sh", ("*.go", "ui/src/*")),
    Gate("scripts/check-filename-policy.sh", ("*.go",)),
    Gate(
        "scripts/check-help-i18n.ts",
        ("internal/i18n/locales/*", "ui/src/*"),
    ),
    Gate("scripts/check-json-casing.sh", ("internal/api/*", "internal/reflector/*")),
    Gate(
        "scripts/check-module-feature-parity.py",
        (
            "internal/api/features.go",
            "internal/license/policy.go",
            "ui/src/constants/moduleFeatures.ts",
        ),
    ),
    Gate(
        "scripts/check-openapi-drift.sh",
        ("docs/openapi*.yaml",),
        generator="cmd/stem-openapi",
    ),
    Gate(
        "scripts/check-output-escaping.sh",
        ("internal/api/*", "internal/help/*", "ui/src/*"),
    ),
    Gate("scripts/check-package-reachability.sh", ("*.go", "go.mod")),
    Gate(
        "scripts/check-release-workflow-contract.sh",
        (".github/workflows/release.yml", ".github/actions/*"),
    ),
    Gate("scripts/check-request-fields.sh", ("internal/api/*", "scripts/requestfields/*")),
    Gate("scripts/check-route-policy.sh", ("internal/api/*", "ui/src/*", "go.mod")),
    Gate(
        "scripts/check-schema-drift.sh",
        ("docs/schemas/api/*",),
        generator="cmd/stem-schema",
    ),
    Gate(
        "scripts/check-service-restart-policy.sh",
        ("deploy/nfpm/*", "deploy/systemd/*"),
    ),
    Gate("scripts/check-token-discipline.sh", ("ui/src/*",)),
    Gate("scripts/check-tsconfig-flags.py", ("ui/tsconfig*.json",)),
    Gate(
        "scripts/check-types-drift.sh",
        (
            "docs/schemas/api/*",
            "ui/scripts/gen-types.mjs",
            "ui/src/types/generated/*",
            "ui/package-lock.json",
        ),
    ),
)

# Gates with no file inputs to select on, and why each is not in GATES.
NOT_INPUT_DRIVEN = {
    # A precondition: the make target runs it before anything else.
    "scripts/check-stale-tests.sh",
    # `make c-test` runs it, and the C step runs `make c-test`.
    "scripts/check-c-test-harness.sh",
    # Reads ui/coverage/coverage-summary.json, which only a whole-suite
    # coverage run (`npm run test:coverage`) writes.
    "scripts/check-ui-coverage.sh",
    # Inspects the installed Playwright WebKit build, not repo files; only the
    # e2e-webkit CI job runs it. Temporary, goes with that job (#1528).
    "scripts/check-webkit-libsoup.sh",
    # Compares the changelog against the commits since the last tag, so its
    # input is git history; the Release Notes CI job runs it on every PR.
    "scripts/check-release-notes.py",
}


@dataclass(frozen=True)
class GoPackage:
    import_path: str
    rel_dir: str
    imports: frozenset[str]
    test_imports: frozenset[str]
    cgo: bool = False


def run_capture(cmd: list[str]) -> str:
    return subprocess.run(
        cmd, cwd=ROOT, check=True, capture_output=True, text=True
    ).stdout


def touched_files() -> list[str]:
    merge_base = run_capture(["git", "merge-base", BASE, "HEAD"]).strip()
    diff = run_capture(["git", "diff", "--name-only", "--no-renames", merge_base])
    untracked = run_capture(["git", "ls-files", "--others", "--exclude-standard"])
    return sorted(set(diff.split()) | set(untracked.split()))


def go_packages() -> dict[str, GoPackage]:
    """Every package in the module under the host's default build tags, by import path."""
    sep = "\x1f"
    fmt = sep.join(
        [
            "{{.ImportPath}}",
            "{{.Dir}}",
            '{{join .Imports ","}}',
            '{{join .TestImports ","}}',
            '{{join .XTestImports ","}}',
            '{{join .CgoFiles ","}}',
        ]
    )
    pkgs = {}
    for line in run_capture(["go", "list", "-f", fmt, "./..."]).splitlines():
        path, directory, imports, test_imports, xtest_imports, cgo = line.split(sep)
        pkgs[path] = GoPackage(
            import_path=path,
            rel_dir=Path(directory).relative_to(ROOT).as_posix(),
            imports=frozenset(filter(None, imports.split(","))),
            test_imports=frozenset(
                filter(None, (test_imports + "," + xtest_imports).split(","))
            ),
            cgo=bool(cgo),
        )
    return pkgs


def owning_package(path: str, by_dir: dict[str, str]) -> str | None:
    """The package a changed file belongs to, if any.

    A .go file belongs to the package in its own directory. Any other file —
    an embedded asset, anything under testdata, including .go fixtures there —
    belongs to the nearest enclosing package, because its tests read it.
    """
    directory = PurePosixPath(path).parent
    if path.endswith(".go") and "testdata" not in directory.parts:
        return by_dir.get(directory.as_posix())
    while directory.as_posix() not in by_dir:
        if directory == PurePosixPath("."):
            return None
        directory = directory.parent
    return by_dir[directory.as_posix()]


def reverse_dependencies(changed: set[str], pkgs: dict[str, GoPackage]) -> set[str]:
    """Changed packages, every in-module package that imports one of them
    directly or transitively, and every package whose tests import any of those."""
    importers: dict[str, set[str]] = {p: set() for p in pkgs}
    for pkg in pkgs.values():
        for dep in pkg.imports & pkgs.keys():
            importers[dep].add(pkg.import_path)
    affected = set(changed)
    frontier = list(changed)
    while frontier:
        for importer in importers[frontier.pop()]:
            if importer not in affected:
                affected.add(importer)
                frontier.append(importer)
    affected |= {p.import_path for p in pkgs.values() if p.test_imports & affected}
    return affected


@dataclass(frozen=True)
class Step:
    label: str
    cmd: list[str]
    cwd: str = "."
    env: dict[str, str] = field(default_factory=dict)


@dataclass
class Plan:
    steps: list[Step]
    notes: list[str]


@dataclass(frozen=True)
class GoScope:
    """The Go packages a change reaches, as import paths."""

    touched: set[str]
    affected: set[str]
    everything: bool
    library_changed: bool


def go_scope(
    touched: list[str], pkgs: dict[str, GoPackage], plan: Plan
) -> GoScope:
    library_changed = any(matches_any(f, C_LIBRARY) for f in touched)
    whole = [f for f in touched if f in WHOLE_MODULE]
    if whole:
        plan.notes.append(
            f"Go: {', '.join(whole)} changed, so every package is in scope"
        )
        return GoScope(set(pkgs), set(pkgs), True, library_changed)
    by_dir = {p.rel_dir: p.import_path for p in pkgs.values()}
    changed = set()
    for path in touched:
        owner = owning_package(path, by_dir)
        if owner:
            changed.add(owner)
        elif path.endswith(".go"):
            plan.notes.append(
                f"Go: {path} is in no package under this host's build tags"
            )
    linked = {p for p, pkg in pkgs.items() if pkg.cgo} if library_changed else set()
    if linked:
        plan.notes.append(
            f"Go: the C library changed, so its {len(linked)} cgo package(s) are in scope"
        )
    affected = reverse_dependencies(changed | linked, pkgs)
    if affected:
        plan.notes.append(
            f"Go: {len(changed)} package(s) touched, "
            f"{len(affected - changed)} reverse dependent(s)"
        )
    else:
        plan.notes.append("Go: no package touched, no Go lint or tests")
    return GoScope(changed, affected, False, library_changed)


def plan_c(touched: list[str], plan: Plan) -> None:
    if any(matches_any(f, C_TESTED) for f in touched):
        plan.steps.append(Step("c-test", ["make", "c-test"]))


def plan_go(
    scope: GoScope, pkgs: dict[str, GoPackage], plan: Plan, golangci: str
) -> None:
    if scope.everything:
        lint_dirs = ["./..."]
        windows = ["./internal/api/..."]
        test_args = ["./..."]
    else:
        lint_dirs = sorted(f"./{pkgs[p].rel_dir}" for p in scope.touched)
        windows = [d for d in lint_dirs if d.startswith("./internal/api")]
        test_args = sorted(scope.affected)
    if lint_dirs:
        lint = [golangci, "run", "--allow-parallel-runners"]
        plan.steps.append(Step("golangci-lint", [*lint, *lint_dirs]))
        if windows:
            # As `make lint-go`: the transport layer is also linted as Windows
            # compiles it, which the host's default lint never type-checks.
            plan.steps.append(
                Step(
                    "golangci-lint (GOOS=windows)",
                    [*lint, *windows],
                    env={"GOOS": "windows"},
                )
            )
    if test_args:
        rebuild = ["-a"] if scope.library_changed else []
        plan.steps.append(Step("go test", ["go", "test", "-race", *rebuild, *test_args]))


def plan_ui(touched: list[str], plan: Plan) -> None:
    whole = [f for f in touched if matches_any(f, WHOLE_UI)]
    sources = [
        f
        for f in touched
        if matches_any(f, UI_INPUTS) and PurePosixPath(f).suffix in UI_SOURCE_SUFFIXES
    ]
    deleted = [f for f in sources if not (ROOT / f).exists()]
    present = [os.path.relpath(f, "ui") for f in sources if f not in deleted]
    if whole or deleted:
        reason = whole or [f"{f} (deleted)" for f in deleted]
        plan.notes.append(
            f"UI: {', '.join(reason)} changed, so the whole Vitest suite runs"
        )
        plan.steps.append(Step("vitest", ["npm", "test"], cwd="ui"))
    elif present:
        plan.steps.append(
            Step("vitest", [*VITEST, "related", "--run", *present], cwd="ui")
        )
    else:
        plan.notes.append("UI: no UI source touched, no Vitest")
        return
    # Vitest strips types without checking them.
    plan.steps.append(Step("typecheck", ["npm", "run", "typecheck"], cwd="ui"))
    biome = [f for f in present if f.startswith("src/")]
    if biome:
        plan.steps.append(
            Step("biome", ["npx", "@biomejs/biome", "check", *biome], cwd="ui")
        )


def plan_docs(touched: list[str], plan: Plan, markdownlint: str) -> None:
    docs = [f for f in touched if f.endswith(".md") and (ROOT / f).exists()]
    if docs:
        cmd = ["npx", "--yes", f"markdownlint-cli2@{markdownlint}", *docs]
        plan.steps.append(Step("markdownlint", cmd))


def plan_gates(
    touched: list[str], scope: GoScope, pkgs: dict[str, GoPackage], plan: Plan
) -> None:
    affected = {pkgs[p].rel_dir for p in scope.affected}
    for gate in GATES:
        if gate.selected(touched, affected):
            for cmd in gate.commands():
                plan.steps.append(Step(f"gate {gate.name}", cmd))


def links_library(step: Step) -> bool:
    """Whether a step builds Go code that may link the C library."""
    if step.label == "go test":
        return True
    return any(
        step.label == f"gate {g.name}" for g in GATES if g.generator is not None
    )


def build_plan(
    touched: list[str],
    pkgs: dict[str, GoPackage],
    golangci: str,
    markdownlint: str,
    library_present: bool,
) -> Plan:
    plan = Plan(steps=[], notes=[])
    scope = go_scope(touched, pkgs, plan)
    plan_c(touched, plan)
    plan_go(scope, pkgs, plan, golangci)
    plan_ui(touched, plan)
    plan_docs(touched, plan, markdownlint)
    plan_gates(touched, scope, pkgs, plan)
    stale = scope.library_changed or not library_present
    if stale and any(links_library(s) for s in plan.steps):
        plan.steps.insert(0, Step("dataplane", ["make", "dataplane"]))
    return plan


def lint_pin(name: str) -> str:
    """A tool version pinned in mk/lint.mk, which in turn matches ci.yml."""
    for line in (ROOT / "mk" / "lint.mk").read_text().splitlines():
        key, _, value = line.partition(":=")
        if key.strip() == name:
            return value.strip()
    sys.exit(f"validate-touched: mk/lint.mk pins no {name}")


def golangci_binary() -> str:
    """The GOPATH golangci-lint, refused unless it is the version CI pins."""
    want = lint_pin("GOLANGCI_LINT_VERSION")
    binary = str(
        Path(run_capture(["go", "env", "GOPATH"]).strip()) / "bin" / "golangci-lint"
    )
    try:
        version = run_capture([binary, "version"])
    except (OSError, subprocess.CalledProcessError):
        version = ""
    if f"version {want.removeprefix('v')} " not in version:
        sys.exit(
            f"validate-touched: {binary} is not golangci-lint {want}; "
            "`make lint-go` installs the pinned version"
        )
    return binary


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--dry-run", action="store_true", help="print the plan, run nothing"
    )
    args = parser.parse_args()

    touched = touched_files()
    print(f"validate-touched: {len(touched)} path(s) differ from {BASE}")
    if not touched:
        return 0
    golangci = "golangci-lint" if args.dry_run else golangci_binary()
    plan = build_plan(
        touched,
        go_packages(),
        golangci,
        lint_pin("MARKDOWNLINT_CLI2_VERSION"),
        (ROOT / DATAPLANE_LIBRARY).exists(),
    )
    for note in plan.notes:
        print(f"  {note}")

    failed = []
    for step in plan.steps:
        where = "" if step.cwd == "." else f"(cd {step.cwd}) "
        env = "".join(f"{k}={v} " for k, v in step.env.items())
        print(f"\n+ {where}{env}{' '.join(step.cmd)}", flush=True)
        if args.dry_run:
            continue
        started = time.monotonic()
        status = subprocess.run(
            step.cmd, cwd=ROOT / step.cwd, env={**os.environ, **step.env}, check=False
        ).returncode
        print(f"  [{step.label}: exit {status}, {time.monotonic() - started:.1f} s]")
        if status != 0:
            failed.append(step.label)
    if failed:
        print(f"\nvalidate-touched: FAILED: {', '.join(failed)}")
        return 1
    verdict = "planned" if args.dry_run else "passed"
    print(f"\nvalidate-touched: {len(plan.steps)} step(s) {verdict}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
