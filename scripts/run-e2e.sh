#!/bin/sh
set -eu

# Run the Playwright suite against a real stem daemon.
#
# The suite needs the daemon, not the Vite dev server. Two things break when it
# is pointed at `npm run dev` on http://localhost:3000:
#
#   1. The session cookies (stem_access, stem_refresh) are Secure. WebKit will
#      not send a Secure cookie to an insecure origin, so every authenticated
#      request 401s, the shell unmounts back to the login overlay, and specs
#      die on "element was detached from the DOM" (#959). Chromium hides this:
#      it treats http://localhost as a trustworthy origin and sends the cookie.
#   2. The dev server serves a development React build, while CI and every
#      shipped binary serve the production bundle embedded in internal/api/ui.
#      A green local run said nothing about the artifact users get.
#
# Both disappear once local runs use the same target CI uses: the daemon, over
# HTTPS, serving the embedded UI. Mirrors seed's scripts/run-e2e.sh.
#
# Arguments are forwarded to Playwright:
#
#   ./scripts/run-e2e.sh --project=webkit
#   ./scripts/run-e2e.sh e2e/help-drawer-smoke.spec.ts
#
# E2E_SKIP_BUILD=1 reuses the current bin/stem and internal/api/ui.

repo_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
run_dir=$(mktemp -d "${TMPDIR:-/tmp}/stem-e2e.XXXXXX")
server_log=${E2E_SERVER_LOG:-$run_dir/server.log}
server_pid=
exit_status=0

cleanup() {
  exit_status=$?
  if [ -n "$server_pid" ] && kill -0 "$server_pid" 2>/dev/null; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  if [ "$exit_status" -ne 0 ] && [ -f "$server_log" ]; then
    printf '\nStem E2E server log:\n' >&2
    tail -200 "$server_log" >&2
  fi
  rm -rf "$run_dir"
  trap - EXIT HUP INT TERM
  exit "$exit_status"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

cd "$repo_dir"

if [ "${E2E_SKIP_BUILD:-0}" != 1 ]; then
  make --no-print-directory build
fi

if [ ! -x ./bin/stem ]; then
  printf '%s\n' 'bin/stem is missing; run without E2E_SKIP_BUILD or build it first' >&2
  exit 1
fi

port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')

(
  cd "$run_dir"
  # Credentials match e2e/helpers/auth.ts TEST_CREDENTIALS; global-setup logs in
  # once with them. The rate limit is raised, not disabled — the whole suite
  # drives one daemon from one IP across two browsers, and the compiled-in
  # default (100/min) runs dry mid-run. It cannot go below the default, so the
  # limiter middleware is still exercised.
  STEM_AUTH_USERNAME=admin \
  STEM_AUTH_PASSWORD=admin \
  STEM_API_RATE_LIMIT=5000 \
    exec "$repo_dir/bin/stem" web -p "$port"
) >"$server_log" 2>&1 &
server_pid=$!

base_url=
attempt=0
while [ "$attempt" -lt 120 ]; do
  # Port fallback (#69) walks +1..+9 when the chosen port is taken, so scan the
  # range rather than assuming the daemon landed on the port we picked.
  offset=0
  while [ "$offset" -le 9 ]; do
    candidate_port=$((port + offset))
    if curl -skf "https://127.0.0.1:$candidate_port/__version" >/dev/null 2>&1; then
      base_url="https://127.0.0.1:$candidate_port"
      break 2
    fi
    offset=$((offset + 1))
  done
  if ! kill -0 "$server_pid" 2>/dev/null; then
    wait "$server_pid" || true
    printf '%s\n' 'Stem exited before becoming ready' >&2
    exit 1
  fi
  attempt=$((attempt + 1))
  sleep 0.25
done

if [ -z "$base_url" ]; then
  printf '%s\n' 'Stem did not become ready within 30 seconds' >&2
  exit 1
fi

# HTTPS-only is a product invariant, and it is what makes the origin secure
# enough for WebKit to send the session cookie. If a plaintext listener ever
# comes back, the suite would silently return to the #959 failure mode.
plain_url="http://${base_url#https://}"
if curl -sf --max-time 2 "$plain_url/__version" >/dev/null 2>&1; then
  printf '%s\n' "Stem served application content over plaintext HTTP at $plain_url" >&2
  exit 1
fi

cd "$repo_dir/ui"
E2E_BASE_URL="$base_url" \
PLAYWRIGHT_IGNORE_HTTPS_ERRORS=true \
  node --disable-warning=DEP0205 ./node_modules/playwright/cli.js test "$@"
