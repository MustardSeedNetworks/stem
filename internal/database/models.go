// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

import (
	"time"
)

// Pagination defaults.
const (
	// defaultPaginationLimit is the default number of items returned per page.
	defaultPaginationLimit = 100
)

// User represents an authenticated user.
type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	PasswordHash   string     `json:"-"` // Never expose in JSON
	Role           string     `json:"role"`
	IsActive       bool       `json:"isActive"`
	LastLogin      *time.Time `json:"lastLogin,omitempty"`
	FailedAttempts int        `json:"failedAttempts"`
	LockedUntil    *time.Time `json:"lockedUntil,omitempty"`
	TokenVersion   int        `json:"tokenVersion"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// TestRun represents a test execution instance.
type TestRun struct {
	ID            string     `json:"id"`
	Module        string     `json:"module"`        // benchmark, servicetest, trafficgen, measure, certify
	TestType      string     `json:"testType"`      // throughput, latency, frame_loss, etc.
	Status        string     `json:"status"`        // pending, running, completed, failed, cancelled
	ConfigJSON    string     `json:"config"`        // JSON string of test configuration
	InterfaceName string     `json:"interfaceName"` // e.g., eth0
	TargetAddress string     `json:"targetAddress"` // Target IP or hostname
	StartedAt     time.Time  `json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	DurationMs    *int64     `json:"durationMs,omitempty"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
	Metadata      string     `json:"metadata,omitempty"` // JSON string for extra data
}

// TestRunStatus constants.
const (
	TestRunStatusPending   = "pending"
	TestRunStatusRunning   = "running"
	TestRunStatusCompleted = "completed"
	TestRunStatusFailed    = "failed"
	TestRunStatusCancelled = "cancelled"
)

// TestResult represents a single test metric data point.
type TestResult struct {
	ID         int64     `json:"id"`
	RunID      string    `json:"runId"`
	MetricType string    `json:"metricType"` // throughput, latency, frame_loss, jitter, etc.
	FrameSize  *int      `json:"frameSize,omitempty"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit,omitempty"` // Mbps, us, %, etc.
	Timestamp  time.Time `json:"timestamp"`
	Metadata   string    `json:"metadata,omitempty"` // JSON string for extra data
}

// MetricType constants for common metric types.
const (
	MetricTypeThroughput = "throughput"
	MetricTypeLatency    = "latency"
	MetricTypeLatencyMin = "latency_min"
	MetricTypeLatencyMax = "latency_max"
	MetricTypeLatencyAvg = "latency_avg"
	MetricTypeFrameLoss  = "frame_loss"
	MetricTypeJitter     = "jitter"
	MetricTypeBackToBack = "back_to_back"
)

// TestSummary represents aggregated test results.
type TestSummary struct {
	ID             int64     `json:"id"`
	RunID          string    `json:"runId"`
	Module         string    `json:"module"`
	TestType       string    `json:"testType"`
	Pass           bool      `json:"pass"`
	ThroughputMbps *float64  `json:"throughputMbps,omitempty"`
	LatencyAvgUs   *float64  `json:"latencyAvgUs,omitempty"`
	LatencyMinUs   *float64  `json:"latencyMinUs,omitempty"`
	LatencyMaxUs   *float64  `json:"latencyMaxUs,omitempty"`
	JitterUs       *float64  `json:"jitterUs,omitempty"`
	FrameLossPct   *float64  `json:"frameLossPct,omitempty"`
	FramesSent     *int64    `json:"framesSent,omitempty"`
	FramesReceived *int64    `json:"framesReceived,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Profile represents a saved test configuration.
type Profile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Module      string    `json:"module"`
	ConfigJSON  string    `json:"config"` // JSON string of config
	IsDefault   bool      `json:"isDefault"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Setting represents a key-value setting.
type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AuditLogEntry represents an audit log entry.
type AuditLogEntry struct {
	ID           int64     `json:"id"`
	Action       string    `json:"action"`
	User         string    `json:"user,omitempty"`
	ResourceType string    `json:"resourceType,omitempty"`
	ResourceID   string    `json:"resourceId,omitempty"`
	OldValueJSON string    `json:"oldValue,omitempty"`
	NewValueJSON string    `json:"newValue,omitempty"`
	IPAddress    string    `json:"ipAddress,omitempty"`
	UserAgent    string    `json:"userAgent,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}


// Session represents a blacklisted/invalidated token session.
type Session struct {
	ID           int64     `json:"id"`
	TokenID      string    `json:"tokenId"`      // JWT ID (jti claim)
	Username     string    `json:"username"`
	Reason       string    `json:"reason"`       // logout, password_change, forced_logout
	BlacklistedAt time.Time `json:"blacklistedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`    // When the token would have expired
}

// SessionReason constants.
const (
	SessionReasonLogout         = "logout"
	SessionReasonPasswordChange = "password_change"
	SessionReasonForcedLogout   = "forced_logout"
)

// AuditAction constants for common audit actions.
const (
	AuditActionLogin        = "login"
	AuditActionLogout       = "logout"
	AuditActionLoginFailed  = "login_failed"
	AuditActionTestStarted  = "test_started"
	AuditActionTestComplete = "test_completed"
	AuditActionTestFailed   = "test_failed"
	AuditActionConfigChange = "config_change"
	AuditActionUserCreated  = "user_created"
	AuditActionUserUpdated  = "user_updated"
)

// TimeRange represents a time range for queries.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Pagination represents pagination parameters.
type Pagination struct {
	Offset int
	Limit  int
}

// DefaultPagination returns default pagination (first 100 items).
func DefaultPagination() Pagination {
	return Pagination{Offset: 0, Limit: defaultPaginationLimit}
}

// TestRunQueryOptions specifies criteria for querying test runs.
type TestRunQueryOptions struct {
	Module    string
	TestType  string
	Status    string
	TimeRange TimeRange
	Limit     int
	Offset    int
}

// TestResultQueryOptions specifies criteria for querying test results.
type TestResultQueryOptions struct {
	RunID      string
	MetricType string
	FrameSize  *int
	TimeRange  TimeRange
	Limit      int
	Offset     int
}
