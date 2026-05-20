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
6. [Kubernetes Deployment](#kubernetes-deployment)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Storage | 1 GB | 10+ GB |
| Network | 1 Gbps NIC | 10+ Gbps NIC |

### Software Requirements

| Software | Version | Notes |
|----------|---------|-------|
| Linux | Kernel 5.4+ | Required for DPDK/XDP |
| macOS | 12+ | Development/testing only |
| Go | 1.25+ | For building from source |
| Node.js | 25+ | For building WebUI |

### Network Requirements

- Network interface with promiscuous mode support
- Root/sudo access for packet capture
- Firewall rules allowing UDP traffic for testing

---

## Installation

### Option 1: Pre-built Binary

```bash
# Download latest release
curl -LO https://github.com/krisarmstrong/stem/releases/latest/download/stem-linux-amd64

# Make executable
chmod +x stem-linux-amd64

# Move to PATH
sudo mv stem-linux-amd64 /usr/local/bin/stem

# Verify installation
stem version
```

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/krisarmstrong/stem.git
cd stem

# Build UI
cd ui && npm ci && npm run build && cd ..

# Build binary
make build

# Install
sudo cp bin/stem /usr/local/bin/
```

### Option 3: Docker

```bash
# Pull image
docker pull ghcr.io/krisarmstrong/stem:latest

# Run container
docker run -d \
  --name stem \
  --network host \
  --cap-add NET_ADMIN \
  -e STEM_AUTH_USERNAME=admin \
  -e STEM_AUTH_PASSWORD=your-secure-password \
  -e STEM_JWT_SECRET=your-256-bit-secret \
  ghcr.io/krisarmstrong/stem:latest web -p 8444
```

---

## Configuration

### Environment Variables (Required)

| Variable | Description | Example |
|----------|-------------|---------|
| `STEM_AUTH_USERNAME` | WebUI username | `admin` |
| `STEM_AUTH_PASSWORD` | WebUI password | `secure-password-here` |
| `STEM_JWT_SECRET` | JWT signing secret (256-bit) | Auto-generated if not set |

### Environment Variables (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `STEM_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `STEM_LOG_FORMAT` | Log format (json, text) | `json` |
| `STEM_DATA_DIR` | Data directory path | `~/.stem` |

### Setting Credentials

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

Create `/etc/stem/stem.env`:

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

# Start with host binding
stem web --host 0.0.0.0 -p 8444

# Plaintext HTTP only (legacy)
stem web --http -p 8444
```

Access at: `https://localhost:8444` (self-signed cert; run
`sudo stem install-ca` once to trust it system-wide).

### Reflector Mode (CLI)

```bash
# Basic reflector
stem reflect -i eth0

# With profile
stem reflect -i eth0 --profile all

# With TUI dashboard
stem reflect -i eth0 --tui
```

### Test Mode (CLI)

```bash
# RFC 2544 throughput test
stem test -i eth0 -t throughput

# Multiple tests
stem test -i eth0 -t throughput,latency,frame_loss

# Y.1564 service activation
stem test -i eth0 -t y1564 --cir 100 --eir 50
```

### TUI Mode

```bash
# Reflector TUI
stem tui --mode reflect -i eth0

# Test Master TUI
stem tui --mode test -i eth0
```

---

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/stem.service`:

```ini
[Unit]
Description=The Stem Network Testing
After=network.target

[Service]
Type=simple
User=stem
Group=stem
EnvironmentFile=/etc/stem/stem.env
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

## Kubernetes Deployment

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stem
  labels:
    app: stem
spec:
  replicas: 1
  selector:
    matchLabels:
      app: stem
  template:
    metadata:
      labels:
        app: stem
    spec:
      containers:
      - name: stem
        image: ghcr.io/krisarmstrong/stem:latest
        ports:
        - containerPort: 8444
        env:
        - name: STEM_AUTH_USERNAME
          valueFrom:
            secretKeyRef:
              name: stem-secrets
              key: username
        - name: STEM_AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: stem-secrets
              key: password
        - name: STEM_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: stem-secrets
              key: jwt-secret
        securityContext:
          capabilities:
            add: ["NET_ADMIN", "NET_RAW"]
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8444
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8444
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: stem
spec:
  selector:
    app: stem
  ports:
  - port: 80
    targetPort: 8444
  type: ClusterIP
```

### Secret

```bash
kubectl create secret generic stem-secrets \
  --from-literal=username=admin \
  --from-literal=password="$(openssl rand -base64 32)" \
  --from-literal=jwt-secret="$(openssl rand -base64 32)"
```

---

## Monitoring

### Health Endpoints

| Endpoint | Purpose | Expected Response |
|----------|---------|-------------------|
| `/health/live` | Liveness probe | 200 OK |
| `/health/ready` | Readiness probe | 200 OK |
| `/api/v1/health` | Detailed health | JSON status |

### Metrics

Prometheus metrics available at `/metrics` (when enabled):

```
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
  "msg": "Test started",
  "test_type": "throughput",
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

- GitHub Issues: https://github.com/krisarmstrong/stem/issues
- Documentation: https://github.com/krisarmstrong/stem/docs

---

*The Stem - Network Performance Testing*
*Mustard Seed Networks*
