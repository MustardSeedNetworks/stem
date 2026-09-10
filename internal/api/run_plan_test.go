// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
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

func TestRunPlanStopsAfterFailedStep(t *testing.T) {
	s := setupTestingTestServer(t)
	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("no network interface available")
	}
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) { return &planExecutor{}, nil }, true
	})
	token := getTestingAuthToken(t, s)
	body := bytes.NewBufferString(`{"peer":"192.0.2.1","tests":[
		{"testType":"rfc2544_throughput"},
		{"testType":"rfc2544_latency"},
		{"testType":"rfc2544_frame_loss"}
	],"interface":"` + ifaces[0].Name + `"}`)
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
			got := []string{stats.Steps[0].Status, stats.Steps[1].Status, stats.Steps[2].Status}
			want := []string{"passed", "failed", "skipped"}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("step statuses = %v, want %v", got, want)
				}
			}
			assertPlanResultPreservesMeasurements(t, s, token)
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("run plan did not reach failed state")
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
