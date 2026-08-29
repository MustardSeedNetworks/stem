#!/usr/bin/env bash
# check-c-test-harness.sh — prove that a failed C assertion fails the process.
#
# The C suite reports results by printing them. For CI, the only thing that
# matters is the exit status, and until 2026-08-29 the two were disconnected:
# TEST_SUMMARY() printed "Some tests failed!" but never returned, and
# tests/c/test_packet_parse.c fell off the end of main — which returns success.
# That test guards attacker-controlled packet bounds and runs in the blocking
# ASAN step, so a parser regression returning a wrong logical result (without
# touching memory it shouldn't) printed FAIL while GitHub Actions stayed green.
#
# Log parsing is not the fix; a harness whose failures cannot fail the build is.
# This script compiles both outcomes against the real header and asserts the
# status each must produce, then checks that no test main can drift back.

set -euo pipefail

cd "$(dirname "$0")/.."

CC="${CC:-cc}"
CFLAGS_CONTRACT="-std=c23 -Wall -Wextra -Werror -Itests/c"
workdir="$(mktemp -d)"
trap 'rm -rf "${workdir}"' EXIT

status=0
fail() {
	printf 'FAIL: %s\n' "$1" >&2
	status=1
}

# A failing assertion must produce a nonzero status, and a passing one must
# produce zero. Checking only the failing half would be satisfied by a harness
# that fails unconditionally.
build_and_run() {
	local name="$1" body="$2"
	cat >"${workdir}/${name}.c" <<EOF
#include "test_framework.h"
TEST(contract) { ${body} }
int main(void)
{
    TEST_SUITE("harness contract");
    RUN_TEST(contract);
    TEST_SUMMARY();

    return TEST_EXIT_STATUS();
}
EOF
	# shellcheck disable=SC2086 # CFLAGS_CONTRACT is a deliberate word list.
	"${CC}" ${CFLAGS_CONTRACT} -o "${workdir}/${name}" "${workdir}/${name}.c" -lm
	set +e
	"${workdir}/${name}" >/dev/null 2>&1
	local rc=$?
	set -e
	printf '%s' "${rc}"
}

rc="$(build_and_run failing 'ASSERT_TRUE(0);')"
if [ "${rc}" -eq 0 ]; then
	fail "a failed assertion exited 0 — C test failures cannot fail CI"
else
	printf 'ok: failed assertion exits %s\n' "${rc}"
fi

rc="$(build_and_run passing 'ASSERT_TRUE(1);')"
if [ "${rc}" -ne 0 ]; then
	fail "a passing assertion exited ${rc} — the harness fails unconditionally"
else
	printf 'ok: passing assertion exits 0\n'
fi

# Every suite built from the shared harness must return its status. A main that
# ends at TEST_SUMMARY() compiles and runs fine; it just cannot fail.
for src in tests/c/*.c; do
	grep -q 'TEST_SUMMARY()' "${src}" || continue
	if ! grep -q 'return TEST_EXIT_STATUS();' "${src}"; then
		fail "${src} calls TEST_SUMMARY() but does not 'return TEST_EXIT_STATUS();'"
	fi
	# `return test_failed;` is truncated to 8 bits: 256 failures would exit 0.
	if grep -q 'return test_failed;' "${src}"; then
		fail "${src} returns the failure count; use TEST_EXIT_STATUS() (status is masked to 8 bits)"
	fi
done

if [ "${status}" -eq 0 ]; then
	printf 'C test harness contract holds.\n'
fi
exit "${status}"
