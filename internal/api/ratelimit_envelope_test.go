// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
)

// A spent budget answers in the same JSON envelope as every other refusal on a
// registered route, which is the 429 docs/openapi.yaml documents (#1438). The
// request goes through the real router, so the limiter is the one the
// Registrar wires, not one built for the test.
func TestRateLimitedRequestAnswersInTheErrorEnvelope(t *testing.T) {
	s := setupTestServer(t)

	get := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil))
		return w
	}
	for i := range ratelimit.APIBurstLimit {
		if w := get(); w.Code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200 inside the burst", i+1, w.Code)
		}
	}

	w := get()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 once the burst is spent", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got != "60" {
		t.Errorf("Retry-After = %q, want 60", got)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var body api.HTTPErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not the error envelope: %v", w.Body.String(), err)
	}
	if body.Error != "Too Many Requests" || body.Code != api.ErrCodeRateLimited {
		t.Errorf("envelope = %+v, want error %q code %s", body, "Too Many Requests", api.ErrCodeRateLimited)
	}
}
