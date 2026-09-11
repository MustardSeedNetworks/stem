// SPDX-License-Identifier: BUSL-1.1

// Package daemonclient talks to a running Stem daemon over its own HTTPS
// API (#1166).
//
// `stem test` and `stem reflect` are clients of the daemon, not a second
// orchestration path: the daemon owns every run, and the CLI and the web UI
// see the same run ID, progress, result and terminal state because they are
// reading the same run. There is deliberately no standalone fallback — a CLI
// that quietly opened its own dataplane when the daemon was unreachable
// would be the second path this package exists to remove.
//
// Request and response types are the daemon's own (internal/api), not copies:
// a second declaration of the wire contract is a second thing to drift.
package daemonclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

const (
	// requestTimeout bounds a single control request. Runs are observed by
	// polling, so no request here waits on a test to finish.
	requestTimeout = 30 * time.Second

	// maxResponseBytes caps what a client will read from one response.
	maxResponseBytes = 4 << 20

	csrfHeader = "X-Csrf-Token"

	pathStart  = "/api/v1/test/start"
	pathStop   = "/api/v1/test/stop"
	pathStats  = "/api/v1/stats"
	pathCSRF   = "/api/v1/auth/csrf-token"
	pathResult = "/api/v1/test/result"
)

// ErrRunInProgress reports that the daemon is already running something.
// Interface ownership is the daemon's to arbitrate, and it refuses a second
// concurrent run rather than letting two dataplanes fight over one NIC.
var ErrRunInProgress = errors.New("the daemon is already running a test")

// FeatureGateError reports a run the licence does not cover. It names the
// feature so the CLI can tell the operator what to buy rather than printing
// a bare 402.
type FeatureGateError struct {
	Feature string
	Message string
}

func (e *FeatureGateError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s requires a Professional licence: %s", e.Feature, e.Message)
	}
	return fmt.Sprintf("%s requires a Professional licence", e.Feature)
}

// Client is an authenticated connection to one daemon.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
	csrf    string
}

// Terminal reports whether a run state has stopped moving, so a watcher can
// stop polling. Anything not known to be in flight counts as terminal: a
// watcher that guesses wrong in the other direction never returns.
func Terminal(state string) bool {
	switch state {
	case "running", "starting":
		return false
	default:
		return true
	}
}

// Open connects to the daemon that published a descriptor in dataDir.
func Open(dataDir string) (*Client, error) {
	descriptor, err := daemonconn.Read(dataDir)
	if err != nil {
		return nil, err
	}
	return newClient(descriptor)
}

// Discover connects to the running daemon on this host, wherever it keeps
// its state. It fails rather than falling back to a local dataplane: the
// daemon is the single owner of a run (#1166).
func Discover() (*Client, error) {
	descriptor, err := daemonconn.Discover()
	if err != nil {
		return nil, err
	}
	return newClient(descriptor)
}

// OpenRemote connects to a daemon named explicitly rather than discovered,
// for the case where the CLI and the daemon are on different hosts. caFile
// may be empty when the daemon serves a certificate the system already
// trusts.
func OpenRemote(baseURL, token, caFile string) (*Client, error) {
	return newClient(daemonconn.Descriptor{URL: baseURL, Token: token, CAFile: caFile})
}

func newClient(d daemonconn.Descriptor) (*Client, error) {
	tlsConfig, err := clientTLSConfig(d.CAFile)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL: d.URL,
		token:   d.Token,
		http: &http.Client{
			Timeout:   requestTimeout,
			Transport: &http.Transport{TLSClientConfig: tlsConfig},
		},
	}, nil
}

// clientTLSConfig trusts exactly the certificate the daemon published. The
// default certificate is self-signed, so the system roots would reject it —
// but skipping verification would accept anything on the port, which is
// precisely what a pinned root prevents.
func clientTLSConfig(caFile string) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if caFile == "" {
		return cfg, nil
	}
	pemBytes, err := os.ReadFile(caFile) // #nosec G304 -- path comes from the daemon's own descriptor
	if err != nil {
		return nil, fmt.Errorf("read daemon certificate %s: %w", caFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("daemon certificate %s holds no usable certificate", caFile)
	}
	cfg.RootCAs = pool
	return cfg, nil
}

// Start submits a run plan and returns the run ID the daemon issued — the
// same ID the web UI shows for this run.
func (c *Client) Start(ctx context.Context, req api.TestStartRequest) (string, error) {
	var response struct {
		Status  string `json:"status"`
		SuiteID string `json:"suiteId"`
		Message string `json:"message"`
	}
	if err := c.do(ctx, http.MethodPost, pathStart, req, &response); err != nil {
		return "", err
	}
	return response.SuiteID, nil
}

// Status reports the daemon's current run state.
func (c *Client) Status(ctx context.Context) (api.Stats, error) {
	var status api.Stats
	if err := c.do(ctx, http.MethodGet, pathStats, nil, &status); err != nil {
		return api.Stats{}, err
	}
	return status, nil
}

// Result returns the outcome the daemon recorded for the last run.
func (c *Client) Result(ctx context.Context) (api.TestResultResponse, error) {
	var result api.TestResultResponse
	if err := c.do(ctx, http.MethodGet, pathResult, nil, &result); err != nil {
		return api.TestResultResponse{}, err
	}
	return result, nil
}

// Stop cancels the running test or reflector.
func (c *Client) Stop(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, pathStop, struct{}{}, nil)
}

// do performs one authenticated request, fetching a CSRF token first when
// the method is one the daemon protects.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	if method != http.MethodGet && c.csrf == "" {
		if err := c.fetchCSRF(ctx); err != nil {
			return err
		}
	}

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrf != "" {
		req.Header.Set(csrfHeader, c.csrf)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("reach the daemon at %s: %w", c.baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read daemon response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errorForStatus(resp.StatusCode, data)
	}
	if out == nil {
		return nil
	}
	if decodeErr := json.Unmarshal(data, out); decodeErr != nil {
		return fmt.Errorf("decode daemon response: %w", decodeErr)
	}
	return nil
}

func (c *Client) fetchCSRF(ctx context.Context) error {
	var response struct {
		Token string `json:"token"`
	}
	if err := c.do(ctx, http.MethodGet, pathCSRF, nil, &response); err != nil {
		return fmt.Errorf("fetch CSRF token: %w", err)
	}
	if response.Token == "" {
		return errors.New("daemon returned an empty CSRF token")
	}
	c.csrf = response.Token
	return nil
}

// apiError is the daemon's error envelope.
type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Feature string `json:"feature"`
	} `json:"error"`
}

// errorForStatus turns the daemon's refusal into something the CLI can act
// on, so an operator sees why a run did not start rather than a status code.
func errorForStatus(status int, body []byte) error {
	var envelope apiError
	_ = json.Unmarshal(body, &envelope)

	switch status {
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", ErrRunInProgress, envelope.Error.Message)
	case http.StatusPaymentRequired:
		return &FeatureGateError{Feature: envelope.Error.Feature, Message: envelope.Error.Message}
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("the daemon rejected this credential (%d): %s", status, envelope.Error.Message)
	default:
		if envelope.Error.Message != "" {
			return fmt.Errorf("daemon returned %d: %s", status, envelope.Error.Message)
		}
		return fmt.Errorf("daemon returned %d", status)
	}
}
