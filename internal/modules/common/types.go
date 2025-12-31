// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package common provides shared types used across all module executors.
// This package exists to avoid import cycles between modules and sub-modules.
package common

import "fmt"

// ErrTestNotImplemented is returned when a test type is not yet implemented.
var ErrTestNotImplemented = fmt.Errorf("test type not implemented")

// ErrModuleNotExecutor is returned when trying to execute on a non-executor module.
var ErrModuleNotExecutor = fmt.Errorf("module does not support execution")

// ErrInvalidConfig is returned when test configuration is invalid.
var ErrInvalidConfig = fmt.Errorf("invalid test configuration")

// Result is a generic test result returned by all module executors.
type Result struct {
	TestType   string      `json:"testType"`
	ModuleName string      `json:"module"`
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// TestConfig holds configuration for test execution.
type TestConfig struct {
	Interface string
	FrameSize uint32
	Duration  int
	Params    map[string]interface{}
}

// Parameter extraction helpers for safe type conversion.
// JSON decoding converts all numbers to float64, so we need to handle both
// native types and float64 conversions.

// GetFloat64Param extracts a float64 parameter from a map, handling both float64 and int types.
func GetFloat64Param(params map[string]interface{}, key string, defaultVal float64) float64 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
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

// GetUint64Param extracts a uint64 parameter from a map, handling float64 and int types.
func GetUint64Param(params map[string]interface{}, key string, defaultVal uint64) uint64 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case float64:
		if val >= 0 {
			return uint64(val)
		}
		return defaultVal
	case uint64:
		return val
	case int64:
		if val >= 0 {
			return uint64(val)
		}
		return defaultVal
	case int:
		if val >= 0 {
			return uint64(val)
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// GetUint32Param extracts a uint32 parameter from a map, handling float64 and int types.
func GetUint32Param(params map[string]interface{}, key string, defaultVal uint32) uint32 {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case float64:
		if val >= 0 && val <= float64(^uint32(0)) {
			return uint32(val)
		}
		return defaultVal
	case uint32:
		return val
	case int:
		if val >= 0 && val <= int(^uint32(0)) {
			return uint32(val)
		}
		return defaultVal
	case int64:
		if val >= 0 && val <= int64(^uint32(0)) {
			return uint32(val)
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// GetIntParam extracts an int parameter from a map, handling float64 type.
func GetIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
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
