#!/bin/sh
#
# Run the stem daemon the E2E suite and the phone-width gate test against, from
# RUN_DIR, listening on PORT. Stays in the foreground (it execs stem), so a
# caller backgrounds it and owns its lifetime.
#
#   scripts/e2e-daemon.sh RUN_DIR PORT
#
# scripts/run-e2e.sh starts it on a free port and reaps it. The phone-width job
# in ci.yml starts it on a fixed one, since the gate's base-url is a workflow
# input.
set -eu

if [ "$#" -ne 2 ]; then
  printf 'usage: %s RUN_DIR PORT\n' "$0" >&2
  exit 2
fi
run_dir=$1
port=$2
repo_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)

# Port fallback (#69) walks +1..+9 when PORT is taken, so every origin the
# daemon can land on is a WebAuthn origin.
webauthn_origins=
offset=0
while [ "$offset" -le 9 ]; do
  webauthn_origins="${webauthn_origins:+$webauthn_origins,}https://localhost:$((port + offset))"
  offset=$((offset + 1))
done

mkdir -p "$run_dir"
cd "$run_dir"
# Credentials match e2e/helpers/auth.ts TEST_CREDENTIALS; global-setup logs in
# once with them. The rate limit is raised, not disabled — the whole suite
# drives one daemon from one IP across two browsers, and the compiled-in
# API (100/min) and authentication (5/min) defaults run dry mid-run. Neither
# override can go below its default, so both limiters remain exercised.
#
# HOME is the run directory because activation state is the one thing stem
# keeps under the home directory (~/.config/stem/.license). Without this a
# local run reads — and a started trial writes — the developer's real license
# file, so the license specs saw a different daemon locally than in CI, where
# the runner's home is always empty.
STEM_AUTH_USERNAME=admin \
STEM_AUTH_PASSWORD=admin \
STEM_API_RATE_LIMIT=5000 \
STEM_AUTH_RATE_LIMIT=200 \
STEM_WEBAUTHN_RPID=localhost \
STEM_WEBAUTHN_ORIGINS="$webauthn_origins" \
HOME="$run_dir" \
  exec "$repo_dir/bin/stem" web -p "$port"
