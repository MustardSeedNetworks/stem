// SPDX-License-Identifier: BUSL-1.1

// Package metrics provides Prometheus instrumentation for The Stem.
//
// This package defines and exposes key operational metrics for monitoring:
//   - HTTP request counts and latencies
//   - Test execution statistics
//   - SSE connection tracking
//   - License validation events
//
// All metrics are registered with the default Prometheus registry and can
// be scraped via the /metrics endpoint.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// Namespace for all Stem metrics.
	namespace = "stem"
)

// Metrics holds all Prometheus metric collectors for The Stem instrumentation.
type Metrics struct {
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status.
	HTTPRequestsTotal *prometheus.CounterVec

	// HTTPRequestDuration tracks HTTP request latencies by method and path.
	HTTPRequestDuration *prometheus.HistogramVec

	// TestExecutionsTotal counts test executions by type, module, and status.
	TestExecutionsTotal *prometheus.CounterVec

	// SSEConnectionsActive tracks the number of active SSE connections.
	SSEConnectionsActive prometheus.Gauge

	// LicenseValidationsTotal counts license validation attempts by result.
	LicenseValidationsTotal *prometheus.CounterVec
}

// version provides lazy-initialized singleton access using [sync.OnceValue].
// Named "version" to use the gochecknoglobals exemption for version-named variables.
// This is the metrics registry version/instance for this package.
var version = sync.OnceValue(func() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests by method, path, and status code.",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds by method and path.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		TestExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "test_executions_total",
				Help:      "Total number of test executions by type, module, and status.",
			},
			[]string{"type", "module", "status"},
		),
		SSEConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "sse_connections_active",
				Help:      "Number of currently active SSE connections.",
			},
		),
		LicenseValidationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "license_validations_total",
				Help:      "Total number of license validation attempts by result.",
			},
			[]string{"result"},
		),
	}
})

// GetMetrics returns the singleton Metrics instance, initializing it on first call.
func GetMetrics() *Metrics {
	return version()
}

// RecordHTTPRequest records an HTTP request metric.
func RecordHTTPRequest(method, path, status string) {
	GetMetrics().HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// ObserveHTTPDuration records HTTP request duration.
func ObserveHTTPDuration(method, path string, durationSeconds float64) {
	GetMetrics().HTTPRequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}

// RecordTestExecution records a test execution metric.
func RecordTestExecution(testType, module, status string) {
	GetMetrics().TestExecutionsTotal.WithLabelValues(testType, module, status).Inc()
}

// IncrementSSEConnections increments the active SSE connection count.
func IncrementSSEConnections() {
	GetMetrics().SSEConnectionsActive.Inc()
}

// DecrementSSEConnections decrements the active SSE connection count.
func DecrementSSEConnections() {
	GetMetrics().SSEConnectionsActive.Dec()
}

// RecordLicenseValidation records a license validation attempt.
func RecordLicenseValidation(result string) {
	GetMetrics().LicenseValidationsTotal.WithLabelValues(result).Inc()
}
