# The Stem - Deployment Guide

**Version**: v0.2.3+
**Last Updated**: 2026-01-06

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Running The Stem](#running-the-stem)
5. [Production Deployment](#production-deployment)
6. [Monitoring](#monitoring)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended |
| ----------- | --------- | ------------- |
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Storage | 1 GB | 10+ GB |
| Network | 1 Gbps NIC | 10+ Gbps NIC |

### Software Requirements

| Software | Version | Notes |
| ---------- | --------- | ------- |
| Linux | Kernel 5.4+ | Required for AF_XDP |
| macOS | 12+ (Apple Silicon) | Runs as a launchd daemon; no dataplane |
| Go | 1.27.4 | For building from source |
| Node.js | 26.13.1 | For building WebUI |

### Network Requirements

- Network interface with promiscuous mode support
- Root/sudo access for packet capture
- Firewall rules allowing UDP traffic for testing

---

## Installation

### Option 1: Signed package from a GitHub release

Every tag publishes a macOS .pkg, .deb, .rpm, Linux/macOS tarballs and Windows zips to
[the releases page](https://github.com/MustardSeedNetworks/stem/releases),
each with a cosign bundle, an SBOM and a SLSA provenance attestation. Asset
names carry the version, so pin one rather than using `latest/download`.

```bash
VERSION=0.24.63   # or the tag you want

# Debian/Ubuntu (arm64: stem_${VERSION}_arm64.deb)
curl -LO https://github.com/MustardSeedNetworks/stem/releases/download/v${VERSION}/stem_${VERSION}_amd64.deb
sudo dpkg -i stem_${VERSION}_amd64.deb

# RHEL/Fedora (arm64: stem-${VERSION}-1.aarch64.rpm)
curl -LO https://github.com/MustardSeedNetworks/stem/releases/download/v${VERSION}/stem-${VERSION}-1.x86_64.rpm
sudo rpm -i stem-${VERSION}-1.x86_64.rpm

# Portable tarball (linux-amd64, linux-arm64, darwin-arm64)
curl -LO https://github.com/MustardSeedNetworks/stem/releases/download/v${VERSION}/stem-${VERSION}-linux-amd64.tar.gz

# macOS (Apple silicon only, arm64): the .pkg, signed and notarized; double-click it.
# See "launchd Service (macOS)" below
curl -LO https://github.com/MustardSeedNetworks/stem/releases/download/v${VERSION}/stem-${VERSION}-darwin-arm64.pkg

# macOS alternative: the tarball carries install.sh, upgrade.sh and uninstall.sh
curl -LO https://github.com/MustardSeedNetworks/stem/releases/download/v${VERSION}/stem-${VERSION}-darwin-arm64.tar.gz
mkdir stem-${VERSION} && tar xzf stem-${VERSION}-darwin-arm64.tar.gz -C stem-${VERSION} && cd stem-${VERSION}
sudo ./install.sh

# Verify installation
stem version
```

To confirm the service answering on a host is that release, run
`make deploy-validate HOST=<host> RELEASE=v${VERSION}` from a stem checkout
with the tag fetched. It checks `/__version` for the release's version and
commit and a non-empty `uiBuildHash`, and fails on any other build.

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/MustardSeedNetworks/stem.git
cd stem

# Build UI
cd ui && npm ci && npm run build && cd ..

# Build binary
make build

# Install
sudo cp bin/stem /usr/local/bin/
```

---

## Configuration

### First-Run Setup

A fresh install needs no credentials. The service starts unclaimed and writes
a one-time setup token to `.setup-token` in its data directory (mode 0600,
readable only by the service user); the startup log names the file. Open
`https://<host>:8444`: the setup page asks for that token, then for the
password of the administrator account `admin`. Read the token on the host, for
example `sudo cat /var/lib/stem/.setup-token`. Until setup completes every
login is refused. The account is kept in an encrypted credential store, the
`auth/` directory of the data directory (`credentials.sealed` and the key that
opens it, `credential.key`, both mode 0600 on Linux and macOS): the Argon2id
password hash, and any TOTP secret and passkeys enrolled later, so they survive
restarts and upgrades. The token file is deleted. A restart before setup
completes issues a new token. A password reset through password recovery is
stored the same way, and it also removes the TOTP secret and every passkey: the
operator signs in with the new password and enrols them again.

A `credentials.json` left by an earlier release is not read. Such an install
comes up unclaimed and is set up again (the file held only the password hash,
never a second factor), unless `STEM_AUTH_USERNAME` and `STEM_AUTH_PASSWORD` are
set, in which case they create the account in the store on that first start.

| Platform | Data directory |
| ---------- | ------------- |
| Linux `.deb` / `.rpm` | `/var/lib/stem` |
| macOS `.pkg` / tarball (`install.sh`) | `/usr/local/stem` |
| Windows zip (`install.ps1`) | `C:\ProgramData\stem` |

The license state (`.license`) lives in the same directory, so the daemon and
`stem license` agree on it; `STEM_DATA_DIR` moves both. With no daemon running,
`stem license` writes the file itself as the account that owns the directory,
so the service can read it: run it as root (it then becomes that account for the
write, and refuses a link at the license path) or as that account. Any other user is refused with
"run as root or as the stem user". On macOS the directory is root's, and on
Windows the directory's access list already limits it to SYSTEM and
Administrators, so no ownership change is made there.

A credential store that exists but cannot be read, fails authentication or has
lost its key is a start failure, not a return to setup. To reclaim the daemon
deliberately, stop it, delete the `auth/` directory and start it again. Back up
`auth/` together: the sealed file is unreadable without `credential.key`.

### Environment Variables

| Variable | Description | Default |
| ---------- | ------------- | --------- |
| `STEM_AUTH_USERNAME` | WebUI username, instead of first-run setup; set both or neither. Used once: it creates the account in the credential store, and the store wins afterwards | unset |
| `STEM_AUTH_PASSWORD` | WebUI password, instead of first-run setup; set both or neither. Used once, like the username | unset |
| `STEM_JWT_SECRET` | JWT signing secret (256-bit) | Auto-generated if not set |
| `STEM_ENABLE_HIBP` | Set to `1` to check new passwords against the Have I Been Pwned breach corpus at setup and recovery. Sends the first five characters of the password's SHA-1 to `api.pwnedpasswords.com`; the password never leaves the host. Off by default, so an air-gapped install never calls out. An unreachable service never blocks a password change | unset (offline) |
| `STEM_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `STEM_LOG_FORMAT` | Log format (json, text) | `json` |
| `STEM_DATA_DIR` | Data directory path | the working directory |
| `STEM_TRUSTED_PROXIES` | Comma-separated CIDRs whose `X-Forwarded-For` may key rate limits and failed-login counters | unset (loopback only) |
| `STEM_WEBAUTHN_ORIGIN` | The HTTPS origin operators browse to, such as `https://stem.example.com:8444`; passkeys work only from it, and its host must be a DNS name, not an IP address | `https://localhost:<bound port>` |

### Login lockout

Five consecutive failed logins for the administrator (a wrong password, or a
wrong second-factor code after a correct password) lock password sign-in for 15
minutes, in memory, per daemon start. A locked account answers exactly like a
wrong password and logs `event=auth.forbidden reason=locked`. A completed login,
or a password set through first-run setup or file-token recovery, clears it;
passkey sign-in is not affected. Failed attempts for an unknown username never
count. The per-address limit (5 per minute) applies in front of this.

### Reverse Proxies

Rate limiting and the failed-login tracker key on the immediate TCP peer, and
believe `X-Forwarded-For` / `X-Real-IP` only from a hop they trust. Loopback is
always trusted, so a reverse proxy on the same host needs no configuration.

A proxy on another host — an ingress controller, a load balancer — is not
trusted until it is named, and until then every client behind it shares one
bucket. Name it with `STEM_TRUSTED_PROXIES`:

```bash
STEM_TRUSTED_PROXIES=10.0.0.0/24,192.168.7.5/32
```

The header is read right to left, discarding entries that are themselves
trusted hops, because each proxy appends the address it saw and the leftmost
entry is whatever the client chose to send. List only the addresses your own
proxies connect from: `0.0.0.0/0` and `::/0` are refused at startup, since
trusting every peer lets any client set its own rate-limit key.

### Setting Credentials From the Environment

For unattended or scripted installs that must not wait on the setup page. A
password stored through setup or recovery takes precedence over these.

```bash
# Generate a secure password
openssl rand -base64 32

# Generate JWT secret
openssl rand -base64 32

# Set environment variables
export STEM_AUTH_USERNAME="admin"
export STEM_AUTH_PASSWORD="$(openssl rand -base64 32)"
export STEM_JWT_SECRET="$(openssl rand -base64 32)"
```

### Persistent Configuration

The packages create `/etc/stem/environment` and the shipped unit reads it.
Set the variables there:

```bash
STEM_AUTH_USERNAME=admin
STEM_AUTH_PASSWORD=your-secure-password
STEM_JWT_SECRET=your-256-bit-secret
STEM_LOG_LEVEL=info
STEM_LOG_FORMAT=json
```

---

## Running The Stem

### WebUI Mode

```bash
# Start web server on default port (HTTPS on 8444)
stem web

# Start on custom port
stem web -p 9443

# Listen on loopback only, keeping the web UI off the network
stem web --host 127.0.0.1
```

Without `--host` the server listens on every address. A port named with `-p`
must be free: `stem web` exits rather than move to another port. Without
`-p`, a busy 8444 falls back to the next free port up to 8453.

Access at: `https://localhost:8444` (self-signed cert; run
`sudo stem install-ca` once to trust it system-wide).

### Reflector Mode (CLI)

The daemon runs the reflector; `stem reflect` asks it to start one, so the
`stem` service must be running (`sudo systemctl start stem`; on macOS see
[launchd Service](#launchd-service-macos)).

```bash
# Basic reflector
stem reflect -i eth0

# With profile
stem reflect -i eth0 --profile all

# Live counters are in the web UI, which shows the same run
stem web -p 8444
```

#### A reflector-only host

For a box whose whole job is to reflect — a permanent test target — ask the
daemon to bring the reflector back whenever it starts:

```bash
sudo systemctl enable --now stem
sudo stem reflect -i eth0 --profile netally --at-boot
```

`--at-boot` records the interface and profile with the daemon, so the
reflector returns after a reboot without a second service. Do **not** add a
separate unit running `stem reflect`: the daemon owns the reflector, and two
owners would contend for the interface.

Check it came back with:

```bash
curl -sk https://localhost:8444/api/v1/reflector/stats
```

### Test Mode (CLI)

```bash
# RFC 2544 throughput test
stem test -i eth0 --peer 192.0.2.10 -t rfc2544_throughput

# Multiple tests
stem test -i eth0 --peer 192.0.2.10 -t rfc2544_throughput,rfc2544_latency,rfc2544_frame_loss

# Y.1564 service activation
stem test -i eth0 --peer 192.0.2.10 -t y1564 --cir 100 --eir 50
```

## Production Deployment

### Systemd Service

The `.deb` and `.rpm` install their own hardened unit at
`/usr/lib/systemd/system/stem.service`; a file at the path below overrides it,
so do not create one on a package install unless you mean to override. This
example is for a tarball install — start from `deploy/systemd/stem.service`
rather than this excerpt if you want the full hardening.

Create `/etc/systemd/system/stem.service`:

```ini
[Unit]
Description=The Stem Network Testing
After=network.target

[Service]
Type=simple
User=stem
Group=stem
EnvironmentFile=-/etc/stem/environment
ExecStart=/usr/local/bin/stem web -p 8444
Restart=always
RestartSec=5
LimitNOFILE=65535

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/stem

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable stem
sudo systemctl start stem
sudo systemctl status stem
```

### launchd Service (macOS)

macOS is Apple silicon only: releases publish an arm64 build and no Intel one.

The `.pkg` is the primary install. It installs `/usr/local/stem/stem` and
`/Library/LaunchDaemons/com.stem.plist` and `com.stem.update.plist`, then
bootstraps both. The `.pkg` is signed with a Developer ID Installer certificate,
notarized by Apple and stapled, so a double-click install passes Gatekeeper.
Check the download against the release's `checksums.txt` and `.cosign.bundle` first. Installing a newer `.pkg`
over an existing install stops the daemon, replaces the binary and plists,
and starts it again; `stem.yaml` (a default is written only if there is none),
`auth/`, `certs/`, `data/` and `.daemon.json` are never touched.
Remove Stem with `uninstall.sh` from the tarball.

The `darwin-arm64` tarball is the alternative. It holds the `stem` binary,
`stem.yaml`, the plists under `launchd/` and three scripts, run as root from
the unpacked directory:

| Script | Does |
| ------ | ---- |
| `sudo ./install.sh` | Copies the binary to `/usr/local/stem`, installs `/Library/LaunchDaemons/com.stem.plist` and `com.stem.update.plist`, and starts the daemon. Refuses if Stem is already installed. |
| `sudo ./upgrade.sh` | Stops the daemon, saves the old binary as `stem.bak`, installs the new binary and plists, and restarts. It restores `stem.bak` if the new binary does not start. It never touches `stem.yaml`, `credentials.json`, `certs/` or `data/`. |
| `sudo ./uninstall.sh [--purge]` | Removes the daemon and binary and keeps configuration and data; `--purge` deletes `/usr/local/stem` entirely. |

`com.stem.plist` runs `/usr/local/stem/stem web` as root and starts it at
boot. launchd restarts it after a crash, not after it exits with an error, so
a daemon that refused to start stays stopped until it is started again.

```bash
sudo launchctl kickstart system/com.stem      # start (add -k to restart)
sudo launchctl print system/com.stem          # state, last exit status
sudo launchctl bootout system/com.stem        # stop and unload
sudo launchctl bootstrap system /Library/LaunchDaemons/com.stem.plist  # load again after bootout
```

Logs are in `/usr/local/stem/logs/stem.log`, which stem writes and rotates itself
(the plist sets `STEM_LOG_FILE`), and `stem.error.log` (stderr, crash traces).

### Windows Service

The Windows `.zip` (`stem-<version>-windows-amd64.zip` or `-arm64`) holds
`stem.exe`, `stem.yaml` and `install.ps1`. Unzip it, then run the script from
an elevated PowerShell in the unpacked directory:

```powershell
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

The script refuses to run without Administrator rights, and:

- copies `stem.exe` to `C:\Program Files\Stem` and adds that directory to the
  system `PATH`;
- registers the `Stem` service (display name "The Stem", automatic delayed
  start), which runs `stem web` on port 8444 as `LocalSystem`;
- adds a Windows Firewall rule named "Stem Web UI" for inbound TCP 8444 if one
  does not exist;
- starts the service and waits until port 8444 accepts connections, so a
  daemon that refuses to start fails the install instead of printing success.

State lives in `C:\ProgramData\stem`: `auth\`, `certs\`, the license
state (`.license`), the data files and `logs\stem.log`. The service runs as
`LocalSystem` and reads this directory, so whoever can write it can run code as
SYSTEM. The script therefore fails closed and never tries to repair a tree it
cannot trust. It refuses, untouched and listing the offending paths, a
directory that is or contains a symbolic link or junction, anything under it
owned by an account other than SYSTEM, Administrators or TrustedInstaller, an
access entry of its own for any other account, or a directory DACL that lets
another account write, delete or change permissions. Otherwise it creates the
directory with, or replaces its access with, owner Administrators, inheritance
removed and full control for SYSTEM and Administrators only, lets everything
inside inherit that, and reads the result back, failing unless it holds.

The service never repairs anything. At every start it resolves the data
directory and checks it with foundation's `securedir`: it refuses to start if
the directory is missing, a link or junction, owned by anyone but SYSTEM,
Administrators or TrustedInstaller, if its parent lets another account delete
or replace it, or if anything under it, at any depth, is a link, has an
untrusted owner, or grants another account write, delete or permission-change
access. The repair is install.ps1. `stem service install` creates a missing
directory with the same protected access and leaves an existing one to
install.ps1 and the start-time check.

The same data directory holds the license for the service and for
`stem license` from a console.

`stem service install` also refuses to register a `stem.exe` outside
`C:\Program Files\Stem`, because a SYSTEM service runs whatever its path names.

Run install.ps1 again from a newer zip to upgrade: it stops the service,
replaces the binary and points the service at it, and keeps the contents of
`C:\ProgramData\stem`. `stem.yaml` is copied there only when it is missing.

Manage the service from an elevated prompt:

```powershell
stem service status
stem service stop
stem service start
stem service uninstall     # removes the service; the data directory stays
```

The service restarts a daemon that fails twice, five seconds apart, and then
stays failed, so a refused start (the data directory unusable, port 8444 in
use) shows up in `stem service status` and `logs\stem.log` instead of
respinning. The service binds 8444 only and never falls back to a neighbouring
port, because the firewall rule opens 8444 alone; to use another port, run
`stem web --port` in a console instead of the service.

`stem service` exists only on Windows. `stem update` replaces `stem.exe` but
does not restart the service; restart it with `stem service stop` and
`stem service start` afterwards.

### Updating

`sudo stem update` installs the newest release in place (Administrator on
Windows). It verifies the release's signature, checksum and build provenance
offline before touching anything, copies the state directory, and restores
both the state and the previous version if the new one is not healthy within
60 seconds. It refuses while a test or reflector run is active and only ever
installs a newer version. `--source` takes `github` (the default), an
`https://` mirror or an absolute directory for an air-gapped site; both lay
releases out as `<base>/vX.Y.Z/<asset>`.

Automatic updates are **off** until an operator turns them on, under
**Settings → Updates** in the web UI or from the CLI:

```bash
sudo stem update --automatic on                              # daily, from GitHub
sudo stem update --automatic on --source /srv/stem-releases  # daily, from a directory
sudo stem update --automatic off
```

Settings shows the source but does not change it; set a mirror or directory
with `--source` once, and the Settings switch keeps it.

A daily check runs `stem update --scheduled` around 03:00 local time. While
automatic updates are off it stops before any network call, so an install that
never turns them on never calls out. A check that finds a run active waits for
the next day.

| Install | Daily check | Its log |
| ------- | ----------- | ------- |
| `.deb` / `.rpm` | `stem-update.timer`, enabled by the package | `journalctl -u stem-update` |
| macOS `.pkg` / tarball | `/Library/LaunchDaemons/com.stem.update.plist`, installed by the `.pkg` or `install.sh` | `/usr/local/stem/logs/update.log` |
| Windows zip | register it once (below) | the task's history |

On Windows, register the task once from an elevated prompt, with the path
where `stem.exe` was unpacked:

```powershell
schtasks /Create /TN "Stem update" /SC DAILY /ST 03:00 /RU SYSTEM `
  /TR "\"C:\Program Files\Stem\stem.exe\" update --scheduled"
```

The `.deb` asks once at install whether to turn automatic updates on, and the
answer defaults to no. A noninteractive install (`DEBIAN_FRONTEND=noninteractive`,
or no terminal) takes that default without prompting. Preseed the answer for an
unattended install:

```bash
echo "stem stem/automatic-updates boolean true" | sudo debconf-set-selections
```

A choice already saved from Settings or the CLI wins over the answer when the
package is upgraded, and `dpkg --purge` forgets it. The `.rpm` cannot prompt,
so it prints the command that turns automatic updates on instead.

The macOS `.pkg` offers the same choice as **Automatic updates** in Installer's
choices pane, unselected. A scripted `sudo installer -pkg …` takes that
default. Select it for an unattended install with a choice-changes file:

```bash
cat > stem-choices.xml <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<array>
    <dict>
        <key>choiceIdentifier</key>
        <string>com.stem.autoupdate</string>
        <key>choiceAttribute</key>
        <string>selected</string>
        <key>attributeSetting</key>
        <integer>1</integer>
    </dict>
</array>
</plist>
XML
sudo installer -pkg stem-<version>-darwin-arm64.pkg -target / -applyChoiceChangesXML stem-choices.xml
```

As with the `.deb`, a choice already saved from Settings or the CLI wins.
`install.sh` leaves automatic updates off; turn them on with
`sudo stem update --automatic on` or from Settings.

### Create Service User

```bash
sudo useradd -r -s /bin/false stem
sudo mkdir -p /var/lib/stem
sudo chown stem:stem /var/lib/stem
```

### Network Capabilities

For packet capture without root:

```bash
sudo setcap cap_net_raw,cap_net_admin=eip /usr/local/bin/stem
```

### Reverse Proxy (nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name stem.example.com;

    ssl_certificate /etc/ssl/certs/stem.crt;
    ssl_certificate_key /etc/ssl/private/stem.key;

    location / {
        proxy_pass https://127.0.0.1:8444;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # SSE support (long-lived connections for real-time updates)
    location /api/v1/events {
        proxy_pass https://127.0.0.1:8444;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400;
    }
}
```

---

## Monitoring

### Health Endpoints

| Endpoint | Purpose | Expected Response |
| ---------- | --------- | ------------------- |
| `/health/live` | Liveness probe | 200 OK |
| `/health/ready` | Readiness probe | 200 OK |
| `/api/v1/health` | Detailed health | JSON status |

### Operations commands

| Command | What it does |
| ------- | ------------ |
| `stem status [--json]` | Reports whether the daemon is running and ready, from `/__version` and `/health/ready` |
| `stem logs [-f] [-n N]` | Shows the end of the service log: the systemd journal on Linux, `<data directory>/logs/stem.log` on macOS; `--file` names any other log |
| `stem support-bundle [-o FILE]` | Writes a redacted diagnostics archive for support; read it before sending |
| `stem backup [-o FILE]` | Archives the daemon's state: administrator credential store (sealed account and its key), reflector and update settings, TLS pair, activation state, published descriptor |
| `stem restore ARCHIVE` | Replaces the state from a backup; refused while the daemon runs, and the previous state is kept beside the data directory |
| `stem version [--json]` | Prints the build; `--json` is the document the daemon serves at `/__version` |

`stem status` finds the daemon through the descriptor in its data directory
(`STEM_DATA_DIR`, else the packaged location) and trusts the certificate that
descriptor names. For a daemon on another host, use `--api URL --cacert FILE`.

Exit status: `0` success, `1` failure. `stem status` reads them as niac does:
`0` running and ready, `1` not running (nothing listening, or up but not
ready), `2` the check could not be made, so a script can tell a stopped
daemon from a check that failed:

```bash
stem status > /dev/null; case $? in
  0) echo healthy ;;
  1) echo stopped ;;
  *) echo "could not check" ;;
esac
```

A backup holds credentials and the TLS private key. It is written readable by
its owner alone and is never overwritten; keep it as you would the data
directory itself.

### Metrics

Prometheus metrics available at `/metrics` (when enabled):

```text
stem_packets_received_total
stem_packets_sent_total
stem_bytes_received_total
stem_bytes_sent_total
stem_current_pps
stem_current_mbps
stem_test_duration_seconds
```

### Logging

Logs are output in JSON format by default:

```json
{
  "time": "2026-01-06T12:00:00Z",
  "level": "INFO",
  "msg": "Starting test via module system",
  "testType": "rfc2544_throughput",
  "module": "benchmark",
  "interface": "eth0"
}
```

Security events include:

- `auth_failure` - Failed login attempts
- `auth_success` - Successful logins
- `token_expired` - Token expiration
- `rate_limited` - Rate limit exceeded

---

## Troubleshooting

### Common Issues

#### Permission Denied on Interface

```bash
# Check capabilities
getcap /usr/local/bin/stem

# Add network capabilities
sudo setcap cap_net_raw,cap_net_admin=eip /usr/local/bin/stem
```

#### Port Already in Use

```bash
# Find process using port
sudo lsof -i :8444

# Kill process
sudo kill -9 <PID>
```

#### SSE Connection Failed

1. Check firewall allows long-lived HTTP connections
2. Verify nginx/proxy SSE configuration (buffering disabled)
3. Check browser console for errors

#### Authentication Failed

1. Verify environment variables are set
2. Check credentials match
3. Clear browser localStorage and retry

### Debug Mode

```bash
# Enable debug logging
export STEM_LOG_LEVEL=debug
stem web -p 8444
```

### Support

- GitHub Issues: https://github.com/MustardSeedNetworks/stem/issues
- Documentation: https://github.com/MustardSeedNetworks/stem/docs

---

_The Stem - Network Performance Testing_
_Mustard Seed Networks_
