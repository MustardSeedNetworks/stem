#!/usr/bin/env python3
"""Guards for two merge-queue races.

release-please's enqueue step used to trust `gh pr checks --required`. That
command exits 0 when every required check that has *reported* passed, so it
passes while `CI Complete` has not been created. GitHub then rejects the
enqueue (`Required status check "CI Complete" is expected`) and the
jump=false fallback exits 1.

upload-sarif on a merge_group run targets the temporary
`gh-readonly-queue/*` ref. The queue deletes that ref when the group merges
or is ejected, and the upload fails with "ref not found" after the scan
itself succeeded.
"""

from __future__ import annotations

import os
import stat
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
WORKFLOWS = ROOT / ".github" / "workflows"
RELEASE_PLEASE = WORKFLOWS / "release-please.yml"

FAKE_GH = r"""#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "${GH_LOG}"

jq_filter=""
args=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --jq)
      jq_filter="$2"
      shift 2
      ;;
    *)
      args+=("$1")
      shift
      ;;
  esac
done

joined="${args[*]}"

emit() {
  local fixture="$1"
  if [[ -n "${jq_filter}" ]]; then
    jq -r "${jq_filter}" <<<"${fixture}"
  else
    printf '%s\n' "${fixture}"
  fi
}

nth() {
  local file="$1"
  local csv="$2"
  local n=0
  if [[ -f "${file}" ]]; then
    n="$(cat "${file}")"
  fi
  n=$((n + 1))
  printf '%s\n' "${n}" > "${file}"
  local IFS=','
  local -a items
  read -r -a items <<< "${csv}"
  local idx=$((n - 1))
  if (( idx >= ${#items[@]} )); then
    idx=$((${#items[@]} - 1))
  fi
  printf '%s\n' "${items[$idx]}"
}

if [[ "${joined}" == pr\ checks* ]]; then
  exit "${FAKE_CHECKS_EXIT:-0}"
fi

if [[ "${joined}" == *releases/latest* ]]; then
  emit '{"tag_name":"v0.30.0"}'
  exit 0
fi

if [[ "${joined}" == *"/compare/"* ]]; then
  emit '{"ahead_by":3,"commits":[{"commit":{"committer":{"date":"2026-10-01T00:00:00Z"}}}]}'
  exit 0
fi

if [[ "${joined}" == pr\ list* ]]; then
  emit '[{"number":1822,"labels":[],"isCrossRepository":false}]'
  exit 0
fi

if [[ "${joined}" == *commits/release-please--branches--main* ]]; then
  emit '{"sha":"abc123","parents":[{"sha":"abc123"}]}'
  exit 0
fi

if [[ "${joined}" == *commits/main* ]]; then
  emit '{"sha":"abc123","parents":[{"sha":"base"}]}'
  exit 0
fi

if [[ "${joined}" == pr\ view* ]]; then
  emit '{"id":"PR_kwDO","headRefOid":"deadbeef"}'
  exit 0
fi

if [[ "${joined}" == *check-runs* ]]; then
  conc="$(nth "${GH_STATE}/ci_n" "${FAKE_CI_CONCLUSIONS:-pending}")"
  if [[ "${conc}" == missing ]]; then
    emit '{"check_runs":[]}'
  elif [[ "${conc}" == pending ]]; then
    emit '{"check_runs":[{"name":"CI Complete","id":1,"conclusion":null,"status":"in_progress"}]}'
  else
    emit "{\"check_runs\":[{\"name\":\"CI Complete\",\"id\":1,\"conclusion\":\"${conc}\",\"status\":\"completed\"}]}"
  fi
  exit 0
fi

if [[ "${joined}" == *enqueuePullRequest* ]]; then
  res="$(nth "${GH_STATE}/enq_n" "${FAKE_ENQUEUE:-ok}")"
  if [[ "${res}" == ok ]]; then
    emit '{"data":{"enqueuePullRequest":{"mergeQueueEntry":{"position":1}}}}'
    exit 0
  fi
  if [[ "${res}" == expected ]]; then
    msg="Pull request Required status check \"CI Complete\" is expected."
    printf '%s\n' "{\"data\":{\"enqueuePullRequest\":null},\"errors\":[{\"type\":\"UNPROCESSABLE\",\"message\":\"${msg}\"}]}"
    printf 'gh: %s\n' "${msg}" >&2
    exit 1
  fi
  printf '%s\n' '{"message":"Resource not accessible by integration"}'
  printf '%s\n' 'gh: Resource not accessible by integration' >&2
  exit 1
fi

if [[ "${joined}" == *graphql* ]]; then
  emit "{\"data\":{\"repository\":{\"mergeQueue\":{\"entries\":{\"totalCount\":${FAKE_QUEUE_DEPTH:-0}}}}}}"
  exit 0
fi

printf 'unhandled gh invocation: %s\n' "${joined}" >&2
exit 97
"""


def enqueue_script() -> str:
    text = RELEASE_PLEASE.read_text(encoding="utf-8")
    marker = "- name: Enqueue the release PR at the front of the merge queue"
    start = text.index(marker)
    run_at = text.index("        run: |\n", start)
    body_start = run_at + len("        run: |\n")
    end = text.find("\n      - name:", start + 1)
    if end == -1:
        end = len(text)
    lines: list[str] = []
    for line in text[body_start:end].splitlines():
        if line.startswith("          "):
            lines.append(line[10:])
        elif line.strip() == "":
            lines.append("")
        else:
            raise AssertionError(f"unexpected enqueue-script indent: {line!r}")
    return "\n".join(lines) + "\n"


class EnqueueWaitsForCIComplete(unittest.TestCase):
    def run_enqueue(self, **env: str) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as tmp:
            bin_dir = Path(tmp) / "bin"
            state = Path(tmp) / "state"
            bin_dir.mkdir()
            state.mkdir()
            gh = bin_dir / "gh"
            gh.write_text(FAKE_GH, encoding="utf-8")
            gh.chmod(gh.stat().st_mode | stat.S_IEXEC)
            sleep = bin_dir / "sleep"
            sleep.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
            sleep.chmod(sleep.stat().st_mode | stat.S_IEXEC)
            log = Path(tmp) / "gh.log"
            full_env = os.environ.copy()
            full_env.update(
                {
                    "PATH": f"{bin_dir}:{full_env.get('PATH', '')}",
                    "GH_LOG": str(log),
                    "GH_STATE": str(state),
                    "GITHUB_REPOSITORY": "MustardSeedNetworks/stem",
                    "GITHUB_REPOSITORY_OWNER": "MustardSeedNetworks",
                    "GITHUB_EVENT_NAME": "workflow_run",
                    "ENQUEUE": "mutation { enqueuePullRequest(input: {pullRequestId: $id, jump: $jump}) { mergeQueueEntry { position } } }",
                    "QUEUE_DEPTH": "query { repository { mergeQueue { entries { totalCount } } } }",
                    "FAKE_CHECKS_EXIT": "0",
                    "FAKE_CI_CONCLUSIONS": "success",
                    "FAKE_ENQUEUE": "ok",
                    "FAKE_QUEUE_DEPTH": "0",
                }
            )
            full_env.update(env)
            proc = subprocess.run(
                ["bash", "-c", enqueue_script()],
                env=full_env,
                capture_output=True,
                text=True,
                timeout=20,
                check=False,
            )
            proc.gh_log = log.read_text(encoding="utf-8") if log.exists() else ""  # type: ignore[attr-defined]
            return proc

    def test_missing_ci_complete_does_not_fail_the_run(self) -> None:
        proc = self.run_enqueue(FAKE_CI_CONCLUSIONS="missing", FAKE_ENQUEUE="expected")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertNotIn("jump=true", proc.gh_log)
        self.assertNotIn("jump=false", proc.gh_log)

    def test_pending_ci_complete_is_waited_for(self) -> None:
        proc = self.run_enqueue(
            FAKE_CI_CONCLUSIONS="pending,pending,success",
            FAKE_ENQUEUE="ok",
        )
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertGreaterEqual(proc.gh_log.count("check-runs"), 3)
        self.assertEqual(proc.gh_log.count("jump=true"), 1)
        self.assertLess(
            proc.gh_log.rindex("check-runs"),
            proc.gh_log.index("jump=true"),
        )

    def test_failed_ci_complete_does_not_enqueue(self) -> None:
        proc = self.run_enqueue(FAKE_CI_CONCLUSIONS="failure", FAKE_ENQUEUE="ok")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout)
        self.assertNotIn("jump=true", proc.gh_log)

    def test_unreported_check_is_retried_then_does_not_fail(self) -> None:
        proc = self.run_enqueue(FAKE_CI_CONCLUSIONS="success", FAKE_ENQUEUE="expected")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertGreaterEqual(proc.gh_log.count("jump=true"), 2)
        self.assertNotIn("jump=false", proc.gh_log)

    def test_unreported_check_retries_until_enqueue_succeeds(self) -> None:
        proc = self.run_enqueue(
            FAKE_CI_CONCLUSIONS="success",
            FAKE_ENQUEUE="expected,expected,ok",
        )
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertEqual(proc.gh_log.count("jump=true"), 3)
        self.assertNotIn("jump=false", proc.gh_log)
        self.assertIn("queued at position 1", proc.stdout)

    def test_required_checks_that_fail_do_not_enqueue(self) -> None:
        proc = self.run_enqueue(FAKE_CHECKS_EXIT="1", FAKE_ENQUEUE="ok")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout)
        self.assertNotIn("enqueuePullRequest", proc.gh_log)

    def test_jump_refusal_on_an_empty_queue_falls_back(self) -> None:
        proc = self.run_enqueue(FAKE_ENQUEUE="denied,ok", FAKE_QUEUE_DEPTH="0")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertIn("jump=true", proc.gh_log)
        self.assertIn("jump=false", proc.gh_log)

    def test_jump_refusal_on_a_busy_queue_waits(self) -> None:
        proc = self.run_enqueue(FAKE_ENQUEUE="denied", FAKE_QUEUE_DEPTH="3")
        self.assertEqual(proc.returncode, 0, proc.stderr + proc.stdout + proc.gh_log)
        self.assertIn("jump=true", proc.gh_log)
        self.assertNotIn("jump=false", proc.gh_log)

    def test_fallback_enqueue_failure_still_fails(self) -> None:
        proc = self.run_enqueue(FAKE_ENQUEUE="denied,denied", FAKE_QUEUE_DEPTH="0")
        self.assertNotEqual(proc.returncode, 0, proc.stdout + proc.gh_log)
        self.assertNotIn("is expected", proc.gh_log)


def _is_sarif_upload(block: str) -> bool:
    """True when the step's `uses:` line uploads SARIF, not when a comment mentions it."""
    for line in block.splitlines():
        stripped = line.strip()
        if not stripped.startswith("uses:"):
            continue
        if "upload-sarif" in stripped or "codeql-action/analyze" in stripped:
            return True
    return False


def _step_blocks(text: str) -> list[str]:
    """Step bodies only. A later job's steps must not leak into an earlier step."""
    lines = text.splitlines(keepends=True)
    blocks: list[str] = []
    current: list[str] = []

    def close() -> None:
        if current:
            blocks.append("".join(current))
            current.clear()

    for line in lines:
        if line.startswith("      - name:") or line.startswith("      - uses:"):
            close()
            current = [line]
            continue
        if not current:
            continue
        if line.strip() == "" or line.startswith("        ") or line.startswith("      #"):
            current.append(line)
            continue
        close()
    close()
    return blocks


class SarifUploadSkipsMergeGroup(unittest.TestCase):
    def test_merge_group_workflows_do_not_fail_on_sarif_upload(self) -> None:
        offenders: list[str] = []
        for path in sorted(WORKFLOWS.glob("*.yml")):
            text = path.read_text(encoding="utf-8")
            if "\n  merge_group:" not in text:
                continue
            for block in _step_blocks(text):
                if not _is_sarif_upload(block):
                    continue
                guarded = (
                    "github.event_name != 'merge_group'" in block
                    or 'github.event_name != "merge_group"' in block
                    or "continue-on-error" in block
                )
                if not guarded:
                    name = block.splitlines()[0].strip()
                    offenders.append(f"{path.name}: {name}")
        self.assertEqual(offenders, [], "\n".join(offenders))

    def test_ci_sarif_uploads_are_actually_scanned(self) -> None:
        text = (WORKFLOWS / "ci.yml").read_text(encoding="utf-8")
        uploads = [block for block in _step_blocks(text) if _is_sarif_upload(block)]
        self.assertGreaterEqual(len(uploads), 2)
        for block in uploads:
            self.assertIn("github.event_name != 'merge_group'", block)


if __name__ == "__main__":
    unittest.main()
