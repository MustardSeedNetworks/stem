// SPDX-License-Identifier: BUSL-1.1

package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/metrics"
)

// -----------------------------------------------------------------------------
// Middleware Tests
// -----------------------------------------------------------------------------

func TestMiddleware_CallsNextHandler(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("Expected next handler to be called")
	}
}

func TestMiddleware_PassesRequestToNextHandler(t *testing.T) {
	t.Parallel()

	var receivedMethod, receivedPath string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if receivedMethod != http.MethodPost {
		t.Errorf("Expected method POST, got %s", receivedMethod)
	}
	if receivedPath != "/api/v1/test" {
		t.Errorf("Expected path /api/v1/test, got %s", receivedPath)
	}
}

func TestMiddleware_RecordsMetrics(t *testing.T) {
	t.Parallel()

	// Get initial counter value.
	uniquePath := "/middleware/test/" + time.Now().Format("20060102150405.000000000")
	initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", uniquePath, "200")

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, uniquePath, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify counter increased.
	newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", uniquePath, "200")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1, got %f -> %f", initialValue, newValue)
	}
}

func TestMiddleware_RecordsDuration(t *testing.T) {
	t.Parallel()

	// Get initial histogram count.
	uniquePath := "/middleware/duration/" + time.Now().Format("20060102150405.000000000")
	initialCount := getHistogramCount(t, metrics.GetMetrics().HTTPRequestDuration, "GET", uniquePath)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, uniquePath, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify histogram count increased.
	newCount := getHistogramCount(t, metrics.GetMetrics().HTTPRequestDuration, "GET", uniquePath)
	if newCount != initialCount+1 {
		t.Errorf("Expected histogram count to increase by 1, got %d -> %d", initialCount, newCount)
	}
}

func TestMiddleware_CapturesStatusCode(t *testing.T) {
	t.Parallel()

	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		t.Run(strconv.Itoa(statusCode), func(t *testing.T) {
			t.Parallel()
			uniquePath := "/middleware/status/" + strconv.Itoa(statusCode) + "/" + time.Now().Format("150405.000000000")

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(statusCode)
			})

			handler := metrics.Middleware(next)
			req := httptest.NewRequest(http.MethodGet, uniquePath, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, w.Code)
			}
		})
	}
}

func TestMiddleware_DefaultsTo200WhenWriteHeaderNotCalled(t *testing.T) {
	t.Parallel()

	uniquePath := "/middleware/default200/" + time.Now().Format("20060102150405.000000000")
	initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", uniquePath, "200")

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Don't call WriteHeader - should default to 200.
		_, _ = w.Write([]byte("OK"))
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, uniquePath, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify counter increased for status 200.
	newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", uniquePath, "200")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter for 200 to increase by 1")
	}
}

func TestMiddleware_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			uniquePath := "/middleware/methods/" + method + "/" + time.Now().Format("150405.000000000")

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := metrics.Middleware(next)
			req := httptest.NewRequest(method, uniquePath, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestMiddleware_PreservesResponseBody(t *testing.T) {
	t.Parallel()

	expectedBody := "Hello, World!"
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestMiddleware_PreservesHeaders(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header custom-value, got %s", w.Header().Get("X-Custom-Header"))
	}
}

func TestMiddleware_ChainedMiddleware(t *testing.T) {
	t.Parallel()

	var order []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware1-before")
			next.ServeHTTP(w, r)
			order = append(order, "middleware1-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// Chain: middleware1 -> metrics.Middleware -> handler
	fullHandler := middleware1(metrics.Middleware(handler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	fullHandler.ServeHTTP(w, req)

	expectedOrder := []string{"middleware1-before", "handler", "middleware1-after"}
	if len(order) != len(expectedOrder) {
		t.Errorf("Expected %d operations, got %d", len(expectedOrder), len(order))
	}
	for i, expected := range expectedOrder {
		if i < len(order) && order[i] != expected {
			t.Errorf("Expected order[%d] = %q, got %q", i, expected, order[i])
		}
	}
}

// -----------------------------------------------------------------------------
// Path Normalization Tests (via Middleware behavior)
// -----------------------------------------------------------------------------

func TestMiddleware_NormalizesModulesPath(t *testing.T) {
	// Serialized against TestMiddleware_ExactPrefixPaths and
	// TestMiddleware_NormalizesTestPath because all four read+increment the
	// shared /api/modules or /api/test Prometheus counters; running them
	// in parallel causes flaky 'expected +1' assertions when concurrent
	// requests increment between the before/after reads.

	// Paths like /api/modules/123 should be normalized to /api/modules.
	// We verify by checking the counter increases by at least 1 for each request.
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	// Test various /api/modules paths.
	paths := []string{
		"/api/modules/benchmark",
		"/api/modules/servicetest",
		"/api/modules/123",
		"/api/modules/some-id/tests",
	}

	for _, path := range paths {
		// Get value before each request to verify individual normalization.
		beforeValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/modules", "200")

		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		afterValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/modules", "200")
		if afterValue <= beforeValue {
			t.Errorf("Expected /api/modules counter to increase for path %s", path)
		}
	}
}

func TestMiddleware_NormalizesTestPath(t *testing.T) {
	// Serialized — see TestMiddleware_NormalizesModulesPath for rationale.

	// Paths like /api/test/123 should be normalized to /api/test.
	// We verify by checking the counter increases by at least 1 for each request.
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	// Test various /api/test paths.
	paths := []string{
		"/api/test/throughput",
		"/api/test/latency",
		"/api/test/123",
		"/api/test/some-id/results",
	}

	for _, path := range paths {
		// Get value before each request to verify individual normalization.
		beforeValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/test", "200")

		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		afterValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/test", "200")
		if afterValue <= beforeValue {
			t.Errorf("Expected /api/test counter to increase for path %s", path)
		}
	}
}

func TestMiddleware_PreservesNonNormalizedPaths(t *testing.T) {
	t.Parallel()

	// Paths that don't match /api/modules or /api/test should be preserved.
	// Use unique paths to avoid interference from other parallel tests.
	paths := []string{
		"/preserve/test/1",
		"/preserve/test/2",
		"/preserve/test/3",
		"/preserve/test/4",
		"/preserve/test/5",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", path, "200")

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", path, "200")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter for %s to increase by 1", path)
			}
		})
	}
}

func TestMiddleware_ShortPaths(t *testing.T) {
	t.Parallel()

	// Short paths that are shorter than the prefix checks should be preserved.
	paths := []string{
		"/",
		"/a",
		"/api",
		"/api/",
		"/api/m",
		"/api/t",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", path, "200")

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", path, "200")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter for %s to increase by 1", path)
			}
		})
	}
}

func TestMiddleware_ExactPrefixPaths(t *testing.T) {
	// Serialized — see TestMiddleware_NormalizesModulesPath for rationale.
	// Sub-tests also run sequentially because both touch shared counters.

	// Paths that are exactly the prefix should be normalized.
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	// Test exact /api/modules path.
	t.Run("/api/modules", func(t *testing.T) {
		// Note: /api/modules is exactly 12 characters, so length check passes.
		// But the path[:12] check requires len > 12.
		// So /api/modules itself should NOT be normalized.
		initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/modules", "200")

		req := httptest.NewRequest(http.MethodGet, "/api/modules", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/modules", "200")
		if newValue != initialValue+1 {
			t.Errorf("Expected counter to increase by 1 for exact /api/modules")
		}
	})

	// Test exact /api/test path.
	t.Run("/api/test", func(t *testing.T) {
		// /api/test is exactly 9 characters.
		initialValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/test", "200")

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		newValue := getCounterValue(t, metrics.GetMetrics().HTTPRequestsTotal, "GET", "/api/test", "200")
		if newValue != initialValue+1 {
			t.Errorf("Expected counter to increase by 1 for exact /api/test")
		}
	})
}

// -----------------------------------------------------------------------------
// responseWriter Tests (via Middleware behavior)
// -----------------------------------------------------------------------------

func TestMiddleware_WriteHeaderCalledOnce(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// Second call should be ignored.
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// First WriteHeader should win.
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestMiddleware_WriteImplicitlyCallsWriteHeader(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Write without WriteHeader should implicitly set 200.
		_, _ = w.Write([]byte("data"))
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// -----------------------------------------------------------------------------
// Concurrency Tests
// -----------------------------------------------------------------------------

func TestMiddleware_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	for range numGoroutines {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/concurrent", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}
}

func TestMiddleware_ConcurrentDifferentPaths(t *testing.T) {
	t.Parallel()

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	paths := []string{
		"/api/modules/1",
		"/api/modules/2",
		"/api/test/1",
		"/api/test/2",
		"/health/live",
	}

	for i := range numGoroutines {
		go func(id int) {
			path := paths[id%len(paths)]
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			done <- true
		}(i)
	}

	for range numGoroutines {
		<-done
	}
}

// -----------------------------------------------------------------------------
// Error Scenarios
// -----------------------------------------------------------------------------

func TestMiddleware_HandlerPanics(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	// Should panic (middleware doesn't recover).
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic to propagate")
		}
	}()

	handler.ServeHTTP(w, req)
}

func TestMiddleware_HandlerWritesPartialResponse(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		// Simulate partial write (no error handling needed for test).
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/partial", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "partial" {
		t.Errorf("Expected body 'partial', got %q", w.Body.String())
	}
}

// -----------------------------------------------------------------------------
// Edge Cases
// -----------------------------------------------------------------------------

func TestMiddleware_EmptyPath(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	// httptest.NewRequest requires a path, so use "/" as minimum.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Manually set empty path.
	req.URL.Path = ""
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_LongPath(t *testing.T) {
	t.Parallel()

	// Very long path.
	longPath := "/api/modules" + strings.Repeat("/segment", 100)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, longPath, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_SpecialCharactersInPath(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/api/test/with%20space",
		"/api/test/with+plus",
		"/api/test/with-dash",
		"/api/test/with_underscore",
		"/api/test/with.dot",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", path, w.Code)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Benchmarks
// -----------------------------------------------------------------------------

func BenchmarkMiddleware(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_WithBody(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_NormalizedPath(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/modules/123", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_Parallel(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := metrics.Middleware(next)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/benchmark/parallel", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}
