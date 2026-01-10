// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package metrics_test

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/krisarmstrong/stem/internal/metrics"
)

// -----------------------------------------------------------------------------
// Test Helpers
// -----------------------------------------------------------------------------

// getCounterValue extracts the value from a CounterVec for given labels.
func getCounterValue(t *testing.T, counter *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()
	metric, err := counter.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("Failed to get metric with labels %v: %v", labels, err)
	}
	m := &dto.Metric{} //nolint:exhaustruct // Only need Counter field populated by Write
	writeErr := metric.Write(m)
	if writeErr != nil {
		t.Fatalf("Failed to write metric: %v", writeErr)
	}
	return m.GetCounter().GetValue()
}

// getHistogramCount extracts the sample count from a HistogramVec for given labels.
func getHistogramCount(t *testing.T, histogram *prometheus.HistogramVec, labels ...string) uint64 {
	t.Helper()
	observer, err := histogram.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("Failed to get metric with labels %v: %v", labels, err)
	}
	// Type assert to Histogram to access the Desc method.
	hist, ok := observer.(prometheus.Histogram)
	if !ok {
		t.Fatalf("Failed to cast observer to Histogram")
	}
	m := &dto.Metric{} //nolint:exhaustruct // Only need Histogram field populated by Write
	writeErr := hist.Write(m)
	if writeErr != nil {
		t.Fatalf("Failed to write metric: %v", writeErr)
	}
	return m.GetHistogram().GetSampleCount()
}

// getGaugeValue extracts the value from a Gauge.
func getGaugeValue(t *testing.T, gauge prometheus.Gauge) float64 {
	t.Helper()
	m := &dto.Metric{} //nolint:exhaustruct // Only need Gauge field populated by Write
	writeErr := gauge.Write(m)
	if writeErr != nil {
		t.Fatalf("Failed to write metric: %v", writeErr)
	}
	return m.GetGauge().GetValue()
}

// -----------------------------------------------------------------------------
// RecordHTTPRequest Tests
// -----------------------------------------------------------------------------

func TestRecordHTTPRequest_IncreasesCounter(t *testing.T) {
	t.Parallel()

	// Get initial value.
	initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/test/record", "200")

	// Record a request.
	metrics.RecordHTTPRequest("GET", "/test/record", "200")

	// Verify counter increased.
	newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/test/record", "200")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1, got %f -> %f", initialValue, newValue)
	}
}

func TestRecordHTTPRequest_MultipleCalls(t *testing.T) {
	t.Parallel()

	// Get initial value.
	initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "POST", "/test/multi", "201")

	// Record multiple requests.
	const numRequests = 5
	for range numRequests {
		metrics.RecordHTTPRequest("POST", "/test/multi", "201")
	}

	// Verify counter increased by numRequests.
	newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "POST", "/test/multi", "201")
	if newValue != initialValue+numRequests {
		t.Errorf("Expected counter to increase by %d, got %f -> %f", numRequests, initialValue, newValue)
	}
}

func TestRecordHTTPRequest_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			path := "/test/methods/" + strings.ToLower(method)
			initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, method, path, "200")

			metrics.RecordHTTPRequest(method, path, "200")

			newValue := getCounterValue(t, metrics.HTTPRequestsTotal, method, path, "200")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for %s, got %f -> %f", method, initialValue, newValue)
			}
		})
	}
}

func TestRecordHTTPRequest_DifferentStatusCodes(t *testing.T) {
	t.Parallel()

	statusCodes := []string{"200", "201", "400", "401", "403", "404", "500", "502", "503"}

	for _, status := range statusCodes {
		t.Run("status_"+status, func(t *testing.T) {
			t.Parallel()
			path := "/test/status/" + status
			initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", path, status)

			metrics.RecordHTTPRequest("GET", path, status)

			newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", path, status)
			if newValue != initialValue+1 {
				t.Errorf(
					"Expected counter to increase by 1 for status %s, got %f -> %f",
					status, initialValue, newValue,
				)
			}
		})
	}
}

func TestRecordHTTPRequest_DifferentPaths(t *testing.T) {
	t.Parallel()

	// Use unique paths to avoid interference from other parallel tests.
	paths := []string{
		"/unique/paths/test/1",
		"/unique/paths/test/2",
		"/unique/paths/test/3",
		"/unique/paths/test/4",
		"/unique/paths/test/5",
	}

	for _, path := range paths {
		t.Run(strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", path, "200")

			metrics.RecordHTTPRequest("GET", path, "200")

			newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", path, "200")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for path %s, got %f -> %f", path, initialValue, newValue)
			}
		})
	}
}

func TestRecordHTTPRequest_EmptyLabels(t *testing.T) {
	t.Parallel()

	// Record with empty method.
	initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "", "/test/empty", "200")
	metrics.RecordHTTPRequest("", "/test/empty", "200")
	newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "", "/test/empty", "200")

	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1 with empty method")
	}

	// Record with empty path.
	initialValue2 := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "", "200")
	metrics.RecordHTTPRequest("GET", "", "200")
	newValue2 := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "", "200")

	if newValue2 != initialValue2+1 {
		t.Errorf("Expected counter to increase by 1 with empty path")
	}

	// Record with empty status.
	initialValue3 := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/test/emptystatus", "")
	metrics.RecordHTTPRequest("GET", "/test/emptystatus", "")
	newValue3 := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/test/emptystatus", "")

	if newValue3 != initialValue3+1 {
		t.Errorf("Expected counter to increase by 1 with empty status")
	}
}

// -----------------------------------------------------------------------------
// ObserveHTTPDuration Tests
// -----------------------------------------------------------------------------

func TestObserveHTTPDuration_RecordsDuration(t *testing.T) {
	t.Parallel()

	// Get initial count.
	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/duration")

	// Record a duration.
	metrics.ObserveHTTPDuration("GET", "/test/duration", 0.5)

	// Verify sample count increased.
	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/duration")
	if newCount != initialCount+1 {
		t.Errorf("Expected histogram count to increase by 1, got %d -> %d", initialCount, newCount)
	}
}

func TestObserveHTTPDuration_MultipleDurations(t *testing.T) {
	t.Parallel()

	// Get initial count.
	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "POST", "/test/multiduration")

	// Record multiple durations.
	durations := []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.5}
	for _, d := range durations {
		metrics.ObserveHTTPDuration("POST", "/test/multiduration", d)
	}

	// Verify sample count increased by len(durations).
	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "POST", "/test/multiduration")
	if newCount != initialCount+uint64(len(durations)) {
		t.Errorf("Expected histogram count to increase by %d, got %d -> %d", len(durations), initialCount, newCount)
	}
}

func TestObserveHTTPDuration_ZeroDuration(t *testing.T) {
	t.Parallel()

	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/zeroduration")

	metrics.ObserveHTTPDuration("GET", "/test/zeroduration", 0.0)

	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/zeroduration")
	if newCount != initialCount+1 {
		t.Errorf("Expected histogram count to increase by 1 for zero duration")
	}
}

func TestObserveHTTPDuration_VerySmallDuration(t *testing.T) {
	t.Parallel()

	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/smallduration")

	metrics.ObserveHTTPDuration("GET", "/test/smallduration", 0.000001) // 1 microsecond

	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/smallduration")
	if newCount != initialCount+1 {
		t.Errorf("Expected histogram count to increase by 1 for very small duration")
	}
}

func TestObserveHTTPDuration_LargeDuration(t *testing.T) {
	t.Parallel()

	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/largeduration")

	metrics.ObserveHTTPDuration("GET", "/test/largeduration", 60.0) // 60 seconds

	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/test/largeduration")
	if newCount != initialCount+1 {
		t.Errorf("Expected histogram count to increase by 1 for large duration")
	}
}

func TestObserveHTTPDuration_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			path := "/test/methods/duration/" + strings.ToLower(method)
			initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, method, path)

			metrics.ObserveHTTPDuration(method, path, 0.1)

			newCount := getHistogramCount(t, metrics.HTTPRequestDuration, method, path)
			if newCount != initialCount+1 {
				t.Errorf("Expected histogram count to increase for %s", method)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// RecordTestExecution Tests
// -----------------------------------------------------------------------------

func TestRecordTestExecution_IncreasesCounter(t *testing.T) {
	t.Parallel()

	initialValue := getCounterValue(t, metrics.TestExecutionsTotal, "throughput", "benchmark", "success")

	metrics.RecordTestExecution("throughput", "benchmark", "success")

	newValue := getCounterValue(t, metrics.TestExecutionsTotal, "throughput", "benchmark", "success")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1, got %f -> %f", initialValue, newValue)
	}
}

func TestRecordTestExecution_DifferentTestTypes(t *testing.T) {
	t.Parallel()

	testTypes := []string{"throughput", "latency", "frame_loss", "back_to_back", "y1564"}

	for _, testType := range testTypes {
		t.Run(testType, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.TestExecutionsTotal, testType, "module", "success")

			metrics.RecordTestExecution(testType, "module", "success")

			newValue := getCounterValue(t, metrics.TestExecutionsTotal, testType, "module", "success")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for %s", testType)
			}
		})
	}
}

func TestRecordTestExecution_DifferentModules(t *testing.T) {
	t.Parallel()

	modules := []string{"benchmark", "servicetest", "trafficgen", "measure", "certify"}

	for _, module := range modules {
		t.Run(module, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.TestExecutionsTotal, "test", module, "success")

			metrics.RecordTestExecution("test", module, "success")

			newValue := getCounterValue(t, metrics.TestExecutionsTotal, "test", module, "success")
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for %s", module)
			}
		})
	}
}

func TestRecordTestExecution_DifferentStatuses(t *testing.T) {
	t.Parallel()

	statuses := []string{"success", "failure", "error", "timeout", "cancelled"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.TestExecutionsTotal, "test", "module", status)

			metrics.RecordTestExecution("test", "module", status)

			newValue := getCounterValue(t, metrics.TestExecutionsTotal, "test", "module", status)
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for status %s", status)
			}
		})
	}
}

func TestRecordTestExecution_MultipleCalls(t *testing.T) {
	t.Parallel()

	initialValue := getCounterValue(t, metrics.TestExecutionsTotal, "multi_test", "multi_module", "success")

	const numExecutions = 10
	for range numExecutions {
		metrics.RecordTestExecution("multi_test", "multi_module", "success")
	}

	newValue := getCounterValue(t, metrics.TestExecutionsTotal, "multi_test", "multi_module", "success")
	if newValue != initialValue+numExecutions {
		t.Errorf("Expected counter to increase by %d, got %f -> %f", numExecutions, initialValue, newValue)
	}
}

// -----------------------------------------------------------------------------
// WebSocket Connection Tests
// -----------------------------------------------------------------------------

func TestIncrementWebSocketConnections(t *testing.T) {
	t.Parallel()

	// Instead of checking exact values (which is flaky with parallel tests),
	// verify that the function doesn't panic and the gauge increases.
	beforeValue := getGaugeValue(t, metrics.WebSocketConnectionsActive)

	metrics.IncrementWebSocketConnections()

	afterValue := getGaugeValue(t, metrics.WebSocketConnectionsActive)
	if afterValue <= beforeValue {
		t.Errorf("Expected gauge to increase, got %f -> %f", beforeValue, afterValue)
	}

	// Clean up by decrementing.
	metrics.DecrementWebSocketConnections()
}

func TestDecrementWebSocketConnections(t *testing.T) {
	t.Parallel()

	// Test that DecrementWebSocketConnections doesn't panic and the gauge reflects changes.
	// We can't reliably check exact values with parallel tests affecting the shared gauge.
	// Instead, verify the increment/decrement pattern works atomically.
	metrics.IncrementWebSocketConnections()
	metrics.IncrementWebSocketConnections() // Ensure gauge is positive
	metrics.DecrementWebSocketConnections()

	// If we get here without panic, the decrement function works.
	// The gauge value is indeterminate due to parallel test interference.
}

func TestWebSocketConnections_IncrementDecrement(t *testing.T) {
	t.Parallel()

	// Test that increment/decrement functions work correctly.
	// Due to parallel test interference, we can't check exact values.
	// Instead, verify the functions don't panic and clean up properly.

	// Simulate opening 5 connections.
	for range 5 {
		metrics.IncrementWebSocketConnections()
	}

	// Simulate closing 5 connections.
	for range 5 {
		metrics.DecrementWebSocketConnections()
	}

	// If we get here without panic, the functions work correctly.
}

func TestWebSocketConnections_MultipleIncrements(t *testing.T) {
	t.Parallel()

	initialValue := getGaugeValue(t, metrics.WebSocketConnectionsActive)

	const numConnections = 100
	for range numConnections {
		metrics.IncrementWebSocketConnections()
	}

	newValue := getGaugeValue(t, metrics.WebSocketConnectionsActive)
	// With parallel tests, we might have interference, so just check it increased significantly.
	if newValue <= initialValue {
		t.Errorf("Expected gauge to increase significantly, got %f -> %f", initialValue, newValue)
	}

	// Clean up.
	for range numConnections {
		metrics.DecrementWebSocketConnections()
	}
}

// -----------------------------------------------------------------------------
// RecordLicenseValidation Tests
// -----------------------------------------------------------------------------

func TestRecordLicenseValidation_IncreasesCounter(t *testing.T) {
	t.Parallel()

	initialValue := getCounterValue(t, metrics.LicenseValidationsTotal, "success")

	metrics.RecordLicenseValidation("success")

	newValue := getCounterValue(t, metrics.LicenseValidationsTotal, "success")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1, got %f -> %f", initialValue, newValue)
	}
}

func TestRecordLicenseValidation_DifferentResults(t *testing.T) {
	t.Parallel()

	results := []string{"success", "failure", "expired", "invalid", "missing"}

	for _, result := range results {
		t.Run(result, func(t *testing.T) {
			t.Parallel()
			initialValue := getCounterValue(t, metrics.LicenseValidationsTotal, result)

			metrics.RecordLicenseValidation(result)

			newValue := getCounterValue(t, metrics.LicenseValidationsTotal, result)
			if newValue != initialValue+1 {
				t.Errorf("Expected counter to increase by 1 for result %s", result)
			}
		})
	}
}

func TestRecordLicenseValidation_MultipleCalls(t *testing.T) {
	t.Parallel()

	initialValue := getCounterValue(t, metrics.LicenseValidationsTotal, "multi_test")

	const numValidations = 15
	for range numValidations {
		metrics.RecordLicenseValidation("multi_test")
	}

	newValue := getCounterValue(t, metrics.LicenseValidationsTotal, "multi_test")
	if newValue != initialValue+numValidations {
		t.Errorf("Expected counter to increase by %d, got %f -> %f", numValidations, initialValue, newValue)
	}
}

func TestRecordLicenseValidation_EmptyResult(t *testing.T) {
	t.Parallel()

	initialValue := getCounterValue(t, metrics.LicenseValidationsTotal, "")

	metrics.RecordLicenseValidation("")

	newValue := getCounterValue(t, metrics.LicenseValidationsTotal, "")
	if newValue != initialValue+1 {
		t.Errorf("Expected counter to increase by 1 with empty result")
	}
}

// -----------------------------------------------------------------------------
// Metric Variable Tests
// -----------------------------------------------------------------------------

func TestHTTPRequestsTotal_IsNotNil(t *testing.T) {
	t.Parallel()

	if metrics.HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal should not be nil")
	}
}

func TestHTTPRequestDuration_IsNotNil(t *testing.T) {
	t.Parallel()

	if metrics.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration should not be nil")
	}
}

func TestTestExecutionsTotal_IsNotNil(t *testing.T) {
	t.Parallel()

	if metrics.TestExecutionsTotal == nil {
		t.Error("TestExecutionsTotal should not be nil")
	}
}

func TestWebSocketConnectionsActive_IsNotNil(t *testing.T) {
	t.Parallel()

	if metrics.WebSocketConnectionsActive == nil {
		t.Error("WebSocketConnectionsActive should not be nil")
	}
}

func TestLicenseValidationsTotal_IsNotNil(t *testing.T) {
	t.Parallel()

	if metrics.LicenseValidationsTotal == nil {
		t.Error("LicenseValidationsTotal should not be nil")
	}
}

// -----------------------------------------------------------------------------
// Concurrency Tests
// -----------------------------------------------------------------------------

func TestRecordHTTPRequest_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	initialValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/concurrent/test", "200")

	for range numGoroutines {
		go func() {
			metrics.RecordHTTPRequest("GET", "/concurrent/test", "200")
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	newValue := getCounterValue(t, metrics.HTTPRequestsTotal, "GET", "/concurrent/test", "200")
	if newValue != initialValue+numGoroutines {
		t.Errorf("Expected counter to be %f, got %f", initialValue+numGoroutines, newValue)
	}
}

func TestObserveHTTPDuration_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	initialCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/concurrent/duration")

	for i := range numGoroutines {
		go func(id int) {
			metrics.ObserveHTTPDuration("GET", "/concurrent/duration", float64(id)*0.001)
			done <- true
		}(i)
	}

	for range numGoroutines {
		<-done
	}

	newCount := getHistogramCount(t, metrics.HTTPRequestDuration, "GET", "/concurrent/duration")
	if newCount != initialCount+numGoroutines {
		t.Errorf("Expected histogram count to be %d, got %d", initialCount+numGoroutines, newCount)
	}
}

func TestRecordTestExecution_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	initialValue := getCounterValue(t, metrics.TestExecutionsTotal, "concurrent", "test", "success")

	for range numGoroutines {
		go func() {
			metrics.RecordTestExecution("concurrent", "test", "success")
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	newValue := getCounterValue(t, metrics.TestExecutionsTotal, "concurrent", "test", "success")
	if newValue != initialValue+numGoroutines {
		t.Errorf("Expected counter to be %f, got %f", initialValue+numGoroutines, newValue)
	}
}

func TestWebSocketConnections_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 50
	done := make(chan bool, numGoroutines*2)

	// Track the delta we apply (should be 0 after equal increments and decrements).
	// We can't check absolute values due to parallel test interference.
	beforeValue := getGaugeValue(t, metrics.WebSocketConnectionsActive)

	// Concurrent increments.
	for range numGoroutines {
		go func() {
			metrics.IncrementWebSocketConnections()
			done <- true
		}()
	}

	// Wait for all increments to complete before decrements.
	for range numGoroutines {
		<-done
	}

	// Verify gauge increased.
	afterIncrement := getGaugeValue(t, metrics.WebSocketConnectionsActive)
	if afterIncrement <= beforeValue {
		t.Errorf("Expected gauge to increase after increments, got %f -> %f", beforeValue, afterIncrement)
	}

	// Concurrent decrements.
	for range numGoroutines {
		go func() {
			metrics.DecrementWebSocketConnections()
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	// Verify gauge decreased after decrements.
	afterDecrement := getGaugeValue(t, metrics.WebSocketConnectionsActive)
	if afterDecrement >= afterIncrement {
		t.Errorf("Expected gauge to decrease after decrements, got %f -> %f", afterIncrement, afterDecrement)
	}
}

func TestRecordLicenseValidation_Concurrency(t *testing.T) {
	t.Parallel()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	initialValue := getCounterValue(t, metrics.LicenseValidationsTotal, "concurrent")

	for range numGoroutines {
		go func() {
			metrics.RecordLicenseValidation("concurrent")
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	newValue := getCounterValue(t, metrics.LicenseValidationsTotal, "concurrent")
	if newValue != initialValue+numGoroutines {
		t.Errorf("Expected counter to be %f, got %f", initialValue+numGoroutines, newValue)
	}
}

// -----------------------------------------------------------------------------
// Benchmarks
// -----------------------------------------------------------------------------

func BenchmarkRecordHTTPRequest(b *testing.B) {
	for b.Loop() {
		metrics.RecordHTTPRequest("GET", "/benchmark", "200")
	}
}

func BenchmarkObserveHTTPDuration(b *testing.B) {
	for b.Loop() {
		metrics.ObserveHTTPDuration("GET", "/benchmark", 0.1)
	}
}

func BenchmarkRecordTestExecution(b *testing.B) {
	for b.Loop() {
		metrics.RecordTestExecution("throughput", "benchmark", "success")
	}
}

func BenchmarkIncrementWebSocketConnections(b *testing.B) {
	for b.Loop() {
		metrics.IncrementWebSocketConnections()
	}
}

func BenchmarkDecrementWebSocketConnections(b *testing.B) {
	// Pre-increment to avoid negative values (setup phase, not the benchmark itself).
	for range b.N { //nolint:modernize // This is setup before ResetTimer, not the benchmark
		metrics.IncrementWebSocketConnections()
	}
	b.ResetTimer()
	for b.Loop() {
		metrics.DecrementWebSocketConnections()
	}
}

func BenchmarkRecordLicenseValidation(b *testing.B) {
	for b.Loop() {
		metrics.RecordLicenseValidation("success")
	}
}

func BenchmarkRecordHTTPRequest_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.RecordHTTPRequest("GET", "/benchmark/parallel", "200")
		}
	})
}

func BenchmarkObserveHTTPDuration_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.ObserveHTTPDuration("GET", "/benchmark/parallel", 0.1)
		}
	})
}
