// SPDX-License-Identifier: BUSL-1.1

package daemonclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// ErrUnreachable reports that nothing accepted the connection: the daemon is
// not running, or not where the target says. It is deliberately narrower than
// "the request failed" — a certificate the client does not trust, a refused
// credential or a daemon answering 500 all mean something IS listening, and
// `stem status` must not report those as a stopped daemon.
var ErrUnreachable = errors.New("no daemon is listening")

const pathHealthReady = "/health/ready"

// Target names a daemon to probe.
type Target struct {
	// URL is the daemon's base URL, scheme included.
	URL string

	// CAFile is the certificate to trust. Empty means the system roots.
	CAFile string
}

// Health is what a daemon says about itself without a credential.
type Health struct {
	// Build is the /__version document: version, commit, buildTime,
	// uiBuildHash, releaseTrain and tlsFingerprint.
	Build map[string]string

	// Ready is the /health/ready document. It is filled on a 503 too: a daemon
	// that is up but not ready says which check failed, and that is the part
	// worth showing.
	Ready api.ReadinessResponse

	// ReadyStatus is the HTTP status /health/ready answered with.
	ReadyStatus int
}

// Healthy reports whether the daemon answered that it is ready.
func (h Health) Healthy() bool {
	return h.ReadyStatus == http.StatusOK && h.Ready.Status == "ready"
}

// Probe asks a daemon for /__version and /health/ready. Both are served
// without authentication, so no token leaves this host and none is needed: a
// status check works for an account that cannot read the descriptor, given the
// daemon's URL and certificate.
//
// The error wraps [ErrUnreachable] only when the TCP connection itself failed.
// The test is the dial operation rather than an errno, because a refused
// connection is ECONNREFUSED on unix and WSAECONNREFUSED on Windows.
func Probe(ctx context.Context, target Target) (Health, error) {
	tlsConfig, err := clientTLSConfig(target.CAFile)
	if err != nil {
		return Health{}, err
	}
	client := &http.Client{
		Timeout:   requestTimeout,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	defer client.CloseIdleConnections()

	var health Health
	status, body, err := get(ctx, client, target.URL+pathBuildVersion)
	if err != nil {
		return Health{}, err
	}
	if status != http.StatusOK {
		return Health{}, fmt.Errorf("%s answered %d", pathBuildVersion, status)
	}
	if decodeErr := json.Unmarshal(body, &health.Build); decodeErr != nil {
		return Health{}, fmt.Errorf("decode %s: %w", pathBuildVersion, decodeErr)
	}

	health.ReadyStatus, body, err = get(ctx, client, target.URL+pathHealthReady)
	if err != nil {
		return Health{}, err
	}
	if health.ReadyStatus != http.StatusOK && health.ReadyStatus != http.StatusServiceUnavailable {
		return Health{}, fmt.Errorf("%s answered %d", pathHealthReady, health.ReadyStatus)
	}
	if decodeErr := json.Unmarshal(body, &health.Ready); decodeErr != nil {
		return Health{}, fmt.Errorf("decode %s: %w", pathHealthReady, decodeErr)
	}
	return health, nil
}

func get(ctx context.Context, client *http.Client, url string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		if opErr, ok := errors.AsType[*net.OpError](err); ok && opErr.Op == "dial" {
			return 0, nil, fmt.Errorf("%w at %s: %w", ErrUnreachable, url, err)
		}
		return 0, nil, fmt.Errorf("reach the daemon at %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, fmt.Errorf("read daemon response: %w", err)
	}
	return resp.StatusCode, body, nil
}
