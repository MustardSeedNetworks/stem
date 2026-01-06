// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package modtypes provides shared types used across all module executors.
// This package exists to avoid import cycles between modules and sub-modules.
package modtypes

import "errors"

// Error sentinels for module operations.
var (
	// ErrTestNotImplemented is returned when a test type is not yet implemented.
	ErrTestNotImplemented = errors.New("test type not implemented")

	// ErrModuleNotExecutor is returned when trying to execute on a non-executor module.
	ErrModuleNotExecutor = errors.New("module does not support execution")

	// ErrInvalidConfig is returned when test configuration is invalid.
	ErrInvalidConfig = errors.New("invalid test configuration")
)

// Result is a generic test result returned by all module executors.
type Result struct {
	TestType   string `json:"testType"`
	ModuleName string `json:"module"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Data       any    `json:"data,omitempty"`
}

// TestConfig holds configuration for test execution.
type TestConfig struct {
	Interface string
	FrameSize uint32
	Duration  int
	Params    map[string]any
}

// Parameter extraction helpers for safe type conversion.
// JSON decoding converts all numbers to float64, so we need to handle both
// native types and float64 conversions.

// GetFloat64Param extracts a float64 parameter from a map.
func GetFloat64Param(params map[string]any, key string, defaultVal float64) float64 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	return convertToFloat64(v, defaultVal)
}

// convertToFloat64 converts various numeric types to float64.
func convertToFloat64(v any, defaultVal float64) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return defaultVal
	}
}

// GetUint64Param extracts a uint64 parameter from a map.
func GetUint64Param(params map[string]any, key string, defaultVal uint64) uint64 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	return convertToUint64(v, defaultVal)
}

// convertToUint64 converts various numeric types to uint64.
func convertToUint64(v any, defaultVal uint64) uint64 {
	switch val := v.(type) {
	case float64:
		if val >= 0 {
			return uint64(val)
		}
	case uint64:
		return val
	case int64:
		if val >= 0 {
			return uint64(val)
		}
	case int:
		if val >= 0 {
			return uint64(val)
		}
	}
	return defaultVal
}

// GetUint32Param extracts a uint32 parameter from a map.
func GetUint32Param(params map[string]any, key string, defaultVal uint32) uint32 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	return convertToUint32(v, defaultVal)
}

// convertToUint32 converts various numeric types to uint32.
func convertToUint32(v any, defaultVal uint32) uint32 {
	const maxUint32 = 4294967295 // ^uint32(0)

	switch val := v.(type) {
	case float64:
		if val >= 0 && val <= maxUint32 {
			return uint32(val)
		}
	case uint32:
		return val
	case int:
		if val >= 0 && val <= maxUint32 {
			return uint32(val)
		}
	case int64:
		if val >= 0 && val <= maxUint32 {
			return uint32(val)
		}
	}
	return defaultVal
}

// GetIntParam extracts an int parameter from a map.
func GetIntParam(params map[string]any, key string, defaultVal int) int {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	return convertToInt(v, defaultVal)
}

// convertToInt converts various numeric types to int.
func convertToInt(v any, defaultVal int) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	default:
		return defaultVal
	}
}

// GetBoolParam extracts a bool parameter from a map.
func GetBoolParam(params map[string]any, key string, defaultVal bool) bool {
	if params == nil {
		return defaultVal
	}
	v, exists := params[key]
	if !exists {
		return defaultVal
	}
	if b, isBool := v.(bool); isBool {
		return b
	}
	return defaultVal
}

// GetStringParam extracts a string parameter from a map.
func GetStringParam(params map[string]any, key string, defaultVal string) string {
	if params == nil {
		return defaultVal
	}
	v, exists := params[key]
	if !exists {
		return defaultVal
	}
	if s, isString := v.(string); isString {
		return s
	}
	return defaultVal
}
