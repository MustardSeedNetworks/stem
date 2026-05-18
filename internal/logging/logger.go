// SPDX-License-Identifier: BUSL-1.1

// Package logging provides structured logging with automatic redaction of sensitive data.
//
// This package wraps Go's log/slog with automatic sensitive data redaction,
// request ID correlation, and configurable output formats (text/JSON).
//
// JSON Log Format:
// When Format is "json", logs are output in structured JSON format:
//
//	{"time":"2026-01-05T10:00:00Z","level":"INFO","msg":"Request processed","component":"server","request_id":"abc123","duration_ms":45}
//
// Environment Variables:
//   - LOG_FORMAT: Set to "json" to enable JSON mode (also respects STEM_LOG_FORMAT)
//   - STEM_LOG_LEVEL: Set log level (debug, info, warn, error)
//   - STEM_LOG_FORMAT: Set log format (text, json)
//
// Context-Aware Logging:
// Use FromContext(ctx) to get a logger with request_id and user_id automatically included.
// Use WithComponent(logger, "server") to add a component field.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config contains logging configuration options.
type Config struct {
	Level      string `yaml:"level"`       // DEBUG, INFO, WARN, ERROR (default: INFO).
	Format     string `yaml:"format"`      // text or json (default: json).
	AddSource  bool   `yaml:"add_source"`  // Include file:line in logs.
	File       string `yaml:"file"`        // Log file path (empty = stdout only).
	MaxSize    int    `yaml:"max_size"`    // Max MB per log file before rotation.
	MaxBackups int    `yaml:"max_backups"` // Number of old files to keep.
	MaxAge     int    `yaml:"max_age"`     // Days to keep old files.
	Compress   bool   `yaml:"compress"`    // Compress rotated files.
	Component  string `yaml:"component"`   // Default component name for all logs.
}

// Log file rotation defaults.
const (
	defaultMaxSizeMB  = 100 // Default max MB per log file before rotation.
	defaultMaxBackups = 5   // Default number of old files to keep.
	defaultMaxAgeDays = 30  // Default days to keep old files.
)

// DefaultConfig returns sensible defaults for logging.
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "json",
		AddSource:  false,
		File:       "",
		MaxSize:    defaultMaxSizeMB,
		MaxBackups: defaultMaxBackups,
		MaxAge:     defaultMaxAgeDays,
		Compress:   true,
		Component:  "",
	}
}

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for request IDs.
	requestIDKey contextKey = "request_id"
	// userIDKey is the context key for user IDs.
	userIDKey contextKey = "user_id"
	// componentKey is the context key for component names.
	componentKey contextKey = "component"
)

// Standard field names for JSON logging.
const (
	// FieldTimestamp is the JSON field name for timestamps.
	FieldTimestamp = "timestamp"
	// FieldLevel is the JSON field name for log level.
	FieldLevel = "level"
	// FieldMessage is the JSON field name for log messages.
	FieldMessage = "message"
	// FieldComponent is the JSON field name for component names.
	FieldComponent = "component"
	// FieldRequestID is the JSON field name for request IDs.
	FieldRequestID = "request_id"
	// FieldDurationMS is the JSON field name for duration in milliseconds.
	FieldDurationMS = "duration_ms"
)

// parseLevel converts a string level to [slog.Level].
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// jsonReplaceAttr returns a ReplaceAttr function for JSON log formatting.
// It customizes field names: time->timestamp, msg->message, and lowercases level values.
func jsonReplaceAttr(isJSON bool) func([]string, slog.Attr) slog.Attr {
	return func(_ []string, a slog.Attr) slog.Attr {
		if !isJSON {
			return a
		}
		switch a.Key {
		case slog.TimeKey:
			if t, ok := a.Value.Any().(time.Time); ok {
				return slog.String(FieldTimestamp, t.UTC().Format(time.RFC3339))
			}
			return slog.Attr{Key: FieldTimestamp, Value: a.Value}
		case slog.MessageKey:
			return slog.Attr{Key: FieldMessage, Value: a.Value}
		case slog.LevelKey:
			if level, ok := a.Value.Any().(slog.Level); ok {
				return slog.String(FieldLevel, strings.ToLower(level.String()))
			}
			return a
		default:
			return a
		}
	}
}

func configureDefaultLogger(cfg *Config, output io.Writer) {
	isJSON := strings.EqualFold(cfg.Format, "json")
	opts := &slog.HandlerOptions{
		Level:       parseLevel(cfg.Level),
		AddSource:   cfg.AddSource,
		ReplaceAttr: jsonReplaceAttr(isJSON),
	}

	var baseHandler slog.Handler
	if isJSON {
		baseHandler = slog.NewJSONHandler(output, opts)
	} else {
		baseHandler = slog.NewTextHandler(output, opts)
	}

	redactingHandler := NewRedactingHandler(baseHandler)

	logger := slog.New(redactingHandler)
	if cfg.Component != "" {
		logger = logger.With(FieldComponent, cfg.Component)
	}

	slog.SetDefault(logger)
}

// Init initializes the global structured logger with the given configuration.
// It sets up output writers (file and/or stdout), log rotation, and the redacting handler.
//
// For JSON format, the output uses customized field names:
//   - "time" -> "timestamp"
//   - "msg" -> "message"
//   - level values are lowercase (e.g., "info" instead of "INFO")
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Check environment variables for format override.
	// LOG_FORMAT takes precedence, then STEM_LOG_FORMAT, then config value.
	if envFormat := os.Getenv("LOG_FORMAT"); envFormat != "" {
		cfg.Format = envFormat
	} else if stemFormat := os.Getenv("STEM_LOG_FORMAT"); stemFormat != "" {
		cfg.Format = stemFormat
	}

	// Determine output writers
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	// Add file writer with rotation if configured
	// Log rotation policy:
	// - Rotates when file reaches MaxSize MB
	// - Keeps up to MaxBackups old log files
	// - Deletes files older than MaxAge days
	// - Compresses rotated files if Compress is true
	if cfg.File != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,    // MB per log file before rotation.
			MaxBackups: cfg.MaxBackups, // Number of old files to keep.
			MaxAge:     cfg.MaxAge,     // Days to keep old files.
			Compress:   cfg.Compress,   // Compress rotated files.
			LocalTime:  false,          // Use UTC for log rotation timestamps.
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer for both stdout and file
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	configureDefaultLogger(cfg, output)

	return nil
}

// Get returns the global logger instance.
// If Init hasn't been called, returns [slog.Default].
func Get() *slog.Logger {
	return slog.Default()
}

// InitWithWriter initializes the logger with a custom writer.
// This is primarily useful for testing to capture log output.
func InitWithWriter(cfg *Config, w io.Writer) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	configureDefaultLogger(cfg, w)

	return nil
}

// Reset resets the global logger to the default slog instance. Used for testing.
func Reset() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// WithRequestID returns a new context with the given request ID.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithUserID returns a new context with the given user ID.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from the context.
func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// WithComponent returns a new context with the given component name.
func WithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, componentKey, component)
}

// ComponentFromContext extracts the component name from the context.
func ComponentFromContext(ctx context.Context) string {
	if comp, ok := ctx.Value(componentKey).(string); ok {
		return comp
	}
	return ""
}

// FromContext returns a logger with contextual information (request ID, user ID, component).
// This is the preferred way to get a logger in HTTP handlers.
func FromContext(ctx context.Context) *slog.Logger {
	logger := Get()
	if component := ComponentFromContext(ctx); component != "" {
		logger = logger.With(FieldComponent, component)
	}
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		logger = logger.With(FieldRequestID, requestID)
	}
	if userID := UserIDFromContext(ctx); userID != "" {
		logger = logger.With("user_id", userID)
	}
	return logger
}

// WithComponentLogger returns a new logger with the specified component name.
// This is useful when you want a component-specific logger without using context.
func WithComponentLogger(component string) *slog.Logger {
	return Get().With(FieldComponent, component)
}

// LogWithDuration logs a message with duration in milliseconds.
// This is a convenience function for timing operations.
func LogWithDuration(ctx context.Context, level slog.Level, msg string, start time.Time, attrs ...any) {
	duration := time.Since(start)
	allAttrs := append([]any{FieldDurationMS, duration.Milliseconds()}, attrs...)
	FromContext(ctx).Log(ctx, level, msg, allAttrs...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// DebugContext logs a debug message with context (includes request_id if present).
func DebugContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context (includes request_id if present).
func InfoContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context (includes request_id if present).
func WarnContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context (includes request_id if present).
func ErrorContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}
