#!/usr/bin/env bash
# test-deploy-validate.sh — proves deploy-validate.sh passes the release it was
# given and fails every other build, against a local HTTPS /__version stub.
#
# The script before #1577 compared the tag name (`v0.25.9`) with the version a
# release build reports (`0.25.9`), so it could never pass a real install.
set -euo pipefail

cd "$(dirname "$0")/.."
readonly VALIDATE="$PWD/scripts/deploy-validate.sh"

work=$(mktemp -d)
server_pid=""
cleanup() {
    [ -n "$server_pid" ] && kill "$server_pid" 2>/dev/null
    rm -rf "$work"
}
trap cleanup EXIT

git -C "$work" init -q
git -C "$work" -c user.name=t -c user.email=t@t commit -q --allow-empty -m release
git -C "$work" tag v1.2.3
commit=$(git -C "$work" rev-parse HEAD)

openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj /CN=localhost \
    -keyout "$work/key.pem" -out "$work/cert.pem" 2>/dev/null

port_file="$work/port"
python3 - "$work" "$port_file" <<'EOF' &
import http.server, pathlib, ssl, sys
work = pathlib.Path(sys.argv[1])
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        body = (work / "version.json").read_bytes()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *args):
        pass
srv = http.server.HTTPServer(("127.0.0.1", 0), H)
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain(work / "cert.pem", work / "key.pem")
srv.socket = ctx.wrap_socket(srv.socket, server_side=True)
pathlib.Path(sys.argv[2]).write_text(str(srv.server_address[1]))
srv.serve_forever()
EOF
server_pid=$!
for _ in $(seq 50); do [ -s "$port_file" ] && break; sleep 0.1; done
port=$(cat "$port_file")

failures=0
# expect <want-exit> <name> <release> <served json>
expect() {
    printf '%s' "$4" >"$work/version.json"
    set +e
    (cd "$work" && "$VALIDATE" --host 127.0.0.1 --port "$port" --release "$3") >"$work/out" 2>&1
    got=$?
    set -e
    if [ "$got" -ne "$1" ]; then
        echo "FAIL: $2: exit $got, want $1"
        sed 's/^/  /' "$work/out"
        failures=$((failures + 1))
    fi
}

ok_json() { printf '{"version":"%s","commit":"%s","uiBuildHash":"%s"}' "$1" "$2" "$3"; }

expect 0 "matching release" v1.2.3 "$(ok_json 1.2.3 "$commit" abc123)"
expect 1 "other version" v1.2.3 "$(ok_json 1.2.4 "$commit" abc123)"
expect 1 "version reported with v" v1.2.3 "$(ok_json v1.2.3 "$commit" abc123)"
expect 1 "other commit" v1.2.3 "$(ok_json 1.2.3 "${commit:0:7}" abc123)"
expect 1 "empty uiBuildHash" v1.2.3 "$(ok_json 1.2.3 "$commit" "")"
expect 1 "unknown uiBuildHash" v1.2.3 "$(ok_json 1.2.3 "$commit" unknown)"
expect 1 "no uiBuildHash field" v1.2.3 "{\"version\":\"1.2.3\",\"commit\":\"$commit\"}"
expect 2 "release not tagged" v9.9.9 "$(ok_json 9.9.9 "$commit" abc123)"

if [ "$failures" -ne 0 ]; then
    echo "test-deploy-validate: $failures case(s) failed"
    exit 1
fi
echo "test-deploy-validate: PASS (8 cases)"
