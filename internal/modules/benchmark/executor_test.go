// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package benchmark

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/modules/common"
)

// Test parameter extraction helpers (issue #24)
func TestGetFloat64Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		defVal   float64
		expected float64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
		{
			name:     "missing key returns default",
			params:   map[string]interface{}{"other": 5.0},
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
		{
			name:     "float64 value",
			params:   map[string]interface{}{"test": 25.5},
			key:      "test",
			defVal:   10.0,
			expected: 25.5,
		},
		{
			name:     "int value converts to float64",
			params:   map[string]interface{}{"test": 42},
			key:      "test",
			defVal:   10.0,
			expected: 42.0,
		},
		{
			name:     "int64 value converts to float64",
			params:   map[string]interface{}{"test": int64(100)},
			key:      "test",
			defVal:   10.0,
			expected: 100.0,
		},
		{
			name:     "string value returns default",
			params:   map[string]interface{}{"test": "not a number"},
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetFloat64Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetFloat64Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUint64Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		defVal   uint64
		expected uint64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "missing key returns default",
			params:   map[string]interface{}{"other": uint64(500)},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "uint64 value",
			params:   map[string]interface{}{"test": uint64(5000)},
			key:      "test",
			defVal:   1000,
			expected: 5000,
		},
		{
			name:     "float64 value converts to uint64",
			params:   map[string]interface{}{"test": 12345.0},
			key:      "test",
			defVal:   1000,
			expected: 12345,
		},
		{
			name:     "int value converts to uint64",
			params:   map[string]interface{}{"test": 999},
			key:      "test",
			defVal:   1000,
			expected: 999,
		},
		{
			name:     "negative float64 returns default",
			params:   map[string]interface{}{"test": -10.0},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "negative int returns default",
			params:   map[string]interface{}{"test": -5},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetUint64Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetUint64Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUint32Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		defVal   uint32
		expected uint32
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   100,
			expected: 100,
		},
		{
			name:     "uint32 value",
			params:   map[string]interface{}{"test": uint32(50)},
			key:      "test",
			defVal:   100,
			expected: 50,
		},
		{
			name:     "float64 value converts to uint32",
			params:   map[string]interface{}{"test": 75.0},
			key:      "test",
			defVal:   100,
			expected: 75,
		},
		{
			name:     "int value converts to uint32",
			params:   map[string]interface{}{"test": 200},
			key:      "test",
			defVal:   100,
			expected: 200,
		},
		{
			name:     "negative value returns default",
			params:   map[string]interface{}{"test": -1},
			key:      "test",
			defVal:   100,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetUint32Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetUint32Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		defVal   int
		expected int
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   10,
			expected: 10,
		},
		{
			name:     "int value",
			params:   map[string]interface{}{"test": 42},
			key:      "test",
			defVal:   10,
			expected: 42,
		},
		{
			name:     "float64 value converts to int",
			params:   map[string]interface{}{"test": 99.9},
			key:      "test",
			defVal:   10,
			expected: 99,
		},
		{
			name:     "negative int works",
			params:   map[string]interface{}{"test": -5},
			key:      "test",
			defVal:   10,
			expected: -5,
		},
		{
			name:     "int64 converts to int",
			params:   map[string]interface{}{"test": int64(1000)},
			key:      "test",
			defVal:   10,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetIntParam(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetIntParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test JSON-like scenarios (all numbers come as float64)
func TestJSONDecodedParams(t *testing.T) {
	// Simulate JSON-decoded params where all numbers are float64
	jsonParams := map[string]interface{}{
		"resolution":    0.1,
		"max_loss":      0.001,
		"warmup":        30.0,
		"initial_burst": 10000.0,
		"trials":        5.0,
	}

	// All these should work with our type-safe helpers
	resolution := common.GetFloat64Param(jsonParams, "resolution", 1.0)
	if resolution != 0.1 {
		t.Errorf("resolution = %v, want 0.1", resolution)
	}

	maxLoss := common.GetFloat64Param(jsonParams, "max_loss", 0.0)
	if maxLoss != 0.001 {
		t.Errorf("max_loss = %v, want 0.001", maxLoss)
	}

	warmup := common.GetIntParam(jsonParams, "warmup", 0)
	if warmup != 30 {
		t.Errorf("warmup = %v, want 30", warmup)
	}

	initialBurst := common.GetUint64Param(jsonParams, "initial_burst", 1000)
	if initialBurst != 10000 {
		t.Errorf("initial_burst = %v, want 10000", initialBurst)
	}

	trials := common.GetUint32Param(jsonParams, "trials", 3)
	if trials != 5 {
		t.Errorf("trials = %v, want 5", trials)
	}
}

func TestModuleInfo(t *testing.T) {
	mod := New()

	if mod.Name() != ModuleName {
		t.Errorf("Expected name '%s', got '%s'", ModuleName, mod.Name())
	}

	if mod.Color() != "#dc2626" {
		t.Errorf("Expected color '#dc2626', got '%s'", mod.Color())
	}

	testTypes := mod.TestTypes()
	if len(testTypes) != 6 {
		t.Errorf("Expected 6 test types, got %d", len(testTypes))
	}

	// Verify all RFC 2544 test types are present
	expectedTypes := []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}
	for _, expected := range expectedTypes {
		found := false
		for _, actual := range testTypes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected test type: %s", expected)
		}
	}
}

func TestCanRun(t *testing.T) {
	mod := New()

	validTests := []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}
	for _, test := range validTests {
		if !mod.CanRun(test) {
			t.Errorf("CanRun(%s) = false, want true", test)
		}
	}

	invalidTests := []string{"invalid", "y1564", "rfc2889"}
	for _, test := range invalidTests {
		if mod.CanRun(test) {
			t.Errorf("CanRun(%s) = true, want false", test)
		}
	}
}
