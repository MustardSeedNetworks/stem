# The Stem - Load Testing

Performance and load testing using [k6](https://k6.io/).

## Prerequisites

### Install k6

```bash
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6
```

### Running Server

Ensure The Stem server is running with authentication configured:

```bash
export STEM_AUTH_USERNAME=admin
export STEM_AUTH_PASSWORD=your-secure-password
export STEM_JWT_SECRET=your-jwt-secret
stem web -p 8080
```

## Test Scripts

| Script | Purpose | Duration | VUs |
| -------- | --------- | ---------- | ----- |
| `auth.js` | Authentication flow testing | ~7 min | 100 |
| `api.js` | API endpoint stress testing | ~9 min | 100 |
| `full.js` | Combined production simulation | ~10 min | 100+ |

## Running Tests

### Basic Usage

```bash
# Set environment variables
export STEM_URL=http://localhost:8080
export STEM_USER=admin
export STEM_PASS=your-password

# Run individual tests
k6 run auth.js
k6 run api.js

# Run full test suite
k6 run full.js
```

### With Docker

```bash
docker run --rm -i \
  -e STEM_URL=http://host.docker.internal:8080 \
  -e STEM_USER=admin \
  -e STEM_PASS=your-password \
  grafana/k6 run - < auth.js
```

### Output Options

```bash
# JSON output for analysis
k6 run --out json=results.json auth.js

# CSV output
k6 run --out csv=results.csv auth.js

# InfluxDB for Grafana dashboards
k6 run --out influxdb=http://localhost:8086/k6 auth.js
```

## Performance Targets

### Authentication (`auth.js`)

- Login: p99 < 100ms
- Token refresh: p99 < 50ms
- Error rate: < 1%

### API (`api.js`)

- Health check: p99 < 50ms
- Modules list: p99 < 100ms
- Overall: p95 < 200ms, p99 < 500ms
- Error rate: < 1%

### Full Suite (`full.js`)

- Overall: p95 < 300ms, p99 < 1s
- SSE connection: p99 < 2s
- Error rate: < 2%

## Test Scenarios

### auth.js

1. **Auth Load Test**: Ramps up to 100 concurrent users performing login/refresh/logout cycles
2. **Rate Limit Test**: Validates rate limiting by exceeding 5 req/min threshold

### api.js

1. **API Load Test**: Ramps up to 100 users hitting various API endpoints
2. **Throughput Test**: Constant 100 req/s for 1 minute

### full.js

1. **API Users**: Simulates typical web dashboard usage
2. **SSE Users**: Long-running SSE connections (dashboard monitors)
3. **Auth Stress**: Continuous authentication operations

## Interpreting Results

### Key Metrics

```text
http_req_duration.............: avg=45ms   min=5ms   med=30ms   max=500ms  p(90)=80ms   p(95)=120ms
http_req_failed...............: 0.50%   ✓ 5      ✗ 995
http_reqs.....................: 1000    83.333/s
```

- **avg**: Average response time
- **p(95)/p(99)**: 95th/99th percentile (important for SLAs)
- **http_req_failed**: Percentage of failed requests
- **http_reqs**: Total requests and rate

### Threshold Failures

```text
✗ http_req_duration..............: avg=250ms min=50ms med=200ms max=5s p(95)=500ms p(99)=1500ms
    ✓ p(95)<300
    ✗ p(99)<1000
```

A ✗ next to a threshold indicates failure. Review the specific metric to identify bottlenecks.

## Troubleshooting

### Connection Refused

```text
ERRO[0001] request failed: dial tcp 127.0.0.1:8080: connect: connection refused
```

Ensure the server is running and accessible at STEM_URL.

### Authentication Failed

```text
WARN[0005] login status is 200: false
```

Check STEM_USER and STEM_PASS environment variables match server configuration.

### Rate Limited

```text
WARN[0030] Request rate limited
```

Expected during rate limit testing. Not an error in normal operation.

### SSE Connection Failed

```text
ERRO[0010] SSE connection failed
```

Check firewall settings and ensure the SSE endpoint is accessible.

## Continuous Integration

Add load tests to CI pipeline (run on schedule, not every commit):

```yaml
# .github/workflows/load-test.yml
name: Load Test
on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly Monday 2am
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start server
        run: |
          export STEM_AUTH_USERNAME=admin
          export STEM_AUTH_PASSWORD=testpass123
          export STEM_JWT_SECRET=testsecret123
          ./bin/stem-linux-amd64 web -p 8080 &
          sleep 5

      - name: Run k6
        uses: grafana/k6-action@v0.3.1
        with:
          filename: tests/load/api.js
        env:
          STEM_URL: http://localhost:8080
          STEM_USER: admin
          STEM_PASS: testpass123
```

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Thresholds](https://k6.io/docs/using-k6/thresholds/)
- [k6 Scenarios](https://k6.io/docs/using-k6/scenarios/)
- [Grafana Cloud k6](https://grafana.com/products/cloud/k6/)
