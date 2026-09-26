// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/servicetest"
)

type planExecutor struct{}

func (*planExecutor) Close() {}

func (*planExecutor) Execute(testType string, _ *modtypes.TestConfig) (*modtypes.Result, error) {
	if testType == "rfc2544_latency" {
		return nil, errors.New("injected failure")
	}
	return &modtypes.Result{
		TestType: testType,
		Success:  true,
		Data:     map[string]any{"throughputMbps": 941.2},
	}, nil
}

// failedServiceDataplane answers a Y.1564 configuration test with the
// verdict the ST-2 bench saw with its reflector stopped: every frame lost.
// The embedded interface is nil, so any other runner panics if called.
type failedServiceDataplane struct {
	servicetest.ServiceDataplane
}

func (failedServiceDataplane) Configure(*dataplane.Config) error { return nil }
func (failedServiceDataplane) Close()                            {}

func (failedServiceDataplane) RunY1564ConfigTest(*dataplane.Y1564Service) (*dataplane.Y1564ConfigResult, error) {
	return &dataplane.Y1564ConfigResult{ServiceID: 1, ServicePass: false}, nil
}

func TestRunPlanStopsAfterFailedStep(t *testing.T) {
	s, token, ifaceName := startPlanServer(t, func(string) (api.TestExecutor, error) {
		return &planExecutor{}, nil
	})
	stats := runPlanUntilFailed(t, s, token, ifaceName, `
		{"testType":"rfc2544_throughput"},
		{"testType":"rfc2544_latency"},
		{"testType":"rfc2544_frame_loss"}`)
	assertStepStatuses(t, stats, "passed", "failed", "skipped")
	assertPlanResultPreservesMeasurements(t, s, token)
}

// A Y.1564 service that fails its acceptance criteria fails its step and the
// plan, keeps its measurements, and skips what follows (#1463).
func TestRunPlanFailsOnFailedServiceVerdict(t *testing.T) {
	s, token, ifaceName := startPlanServer(t, func(string) (api.TestExecutor, error) {
		return servicetest.NewExecutorWithDataplane(failedServiceDataplane{}), nil
	})
	stats := runPlanUntilFailed(t, s, token, ifaceName, `
		{"testType":"y1564_config"},
		{"testType":"y1564_config"}`)
	assertStepStatuses(t, stats, "failed", "skipped")
	if stats.Steps[0].Error == "" {
		t.Error("the failed step carries no error for the operator")
	}
	if !strings.Contains(stats.ErrorMessage, "acceptance criteria") {
		t.Errorf("run error = %q, want the criteria cause, not a pointer at the daemon log", stats.ErrorMessage)
	}
	if stats.Steps[0].Result == nil || stats.Steps[0].Result.Data == nil {
		t.Errorf("the failed step dropped its measurements: %+v", stats.Steps[0].Result)
	}
}

func startPlanServer(t *testing.T, factory api.TestExecutorFactory) (*api.Server, string, string) {
	t.Helper()
	s := setupTestingTestServer(t)
	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("no network interface available")
	}
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) { return factory, true })
	return s, getTestingAuthToken(t, s), ifaces[0].Name
}

func runPlanUntilFailed(t *testing.T, s *api.Server, token, ifaceName, tests string) api.Stats {
	t.Helper()
	body := bytes.NewBufferString(`{"peer":"192.0.2.1","tests":[` + tests + `],"interface":"` + ifaceName + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start status = %d: %s", w.Code, w.Body.String())
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		statsReq := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
		statsReq.Header.Set("Authorization", "Bearer "+token)
		statsW := httptest.NewRecorder()
		s.ServeHTTP(statsW, statsReq)
		var stats api.Stats
		if decodeErr := json.Unmarshal(statsW.Body.Bytes(), &stats); decodeErr != nil {
			t.Fatalf("decode stats: %v", decodeErr)
		}
		if stats.TestStatus == "error" {
			return stats
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("run plan did not reach failed state")
	return api.Stats{}
}

func assertStepStatuses(t *testing.T, stats api.Stats, want ...string) {
	t.Helper()
	got := make([]string, len(stats.Steps))
	for i, step := range stats.Steps {
		got[i] = step.Status
	}
	if len(got) != len(want) {
		t.Fatalf("step statuses = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step statuses = %v, want %v", got, want)
		}
	}
}

func assertPlanResultPreservesMeasurements(t *testing.T, s *api.Server, token string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	var result api.TestResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if len(result.Steps) != 3 || result.Steps[0].Result == nil {
		t.Fatalf("plan result omitted completed step measurements: %+v", result.Steps)
	}
	data, ok := result.Steps[0].Result.Data.(map[string]any)
	if !ok || data["throughputMbps"] != 941.2 {
		t.Fatalf("completed step data = %#v", result.Steps[0].Result.Data)
	}
}
