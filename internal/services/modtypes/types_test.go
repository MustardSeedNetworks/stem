// SPDX-License-Identifier: BUSL-1.1

// Black-box tests for the parameter helpers every module executor uses to read
// its configuration. These moved here from internal/services/trafficgen, which
// owns none of this code.
package modtypes_test

import (
	"math"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// TestModtypesSafeIntToUint32 tests the modtypes.SafeIntToUint32 helper function.
func TestModtypesSafeIntToUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected uint32
	}{
		{
			name:     "positive value within range",
			value:    100,
			expected: 100,
		},
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "negative value returns 0",
			value:    -1,
			expected: 0,
		},
		{
			name:     "large negative value returns 0",
			value:    -1000000,
			expected: 0,
		},
		{
			name:     "max int32 value",
			value:    math.MaxInt32,
			expected: math.MaxInt32,
		},
		{
			name:     "max uint32 as int (if int is 64-bit)",
			value:    math.MaxUint32,
			expected: math.MaxUint32,
		},
		{
			name:     "value just above max uint32 returns max",
			value:    math.MaxUint32 + 1,
			expected: math.MaxUint32,
		},
		{
			name:     "large positive value over uint32 max returns max",
			value:    int(math.MaxInt64 / 2),
			expected: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.SafeIntToUint32(tt.value)
			if result != tt.expected {
				t.Errorf("modtypes.SafeIntToUint32(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint16Param tests the modtypes.GetUint16Param helper function.
func TestModtypesGetUint16Param(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		key        string
		defaultVal uint16
		expected   uint16
	}{
		{
			name:       "zero value",
			params:     map[string]any{"val": float64(0)},
			key:        "val",
			defaultVal: 100,
			expected:   0,
		},
		{
			name:       "small value",
			params:     map[string]any{"val": float64(100)},
			key:        "val",
			defaultVal: 0,
			expected:   100,
		},
		{
			name:       "mid range value",
			params:     map[string]any{"val": float64(32000)},
			key:        "val",
			defaultVal: 0,
			expected:   32000,
		},
		{
			name:       "max uint16 value",
			params:     map[string]any{"val": float64(math.MaxUint16)},
			key:        "val",
			defaultVal: 0,
			expected:   math.MaxUint16,
		},
		{
			name:       "value just above max uint16 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 1)},
			key:        "val",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "large value returns default",
			params:     map[string]any{"val": float64(math.MaxUint32)},
			key:        "val",
			defaultVal: 99,
			expected:   99,
		},
		{
			name:       "typical VLAN ID",
			params:     map[string]any{"val": float64(4094)},
			key:        "val",
			defaultVal: 0,
			expected:   4094,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint16Param(tt.params, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint16Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint8Param tests the modtypes.GetUint8Param helper function.
func TestModtypesGetUint8Param(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		key        string
		defaultVal uint8
		expected   uint8
	}{
		{
			name:       "zero value",
			params:     map[string]any{"val": float64(0)},
			key:        "val",
			defaultVal: 100,
			expected:   0,
		},
		{
			name:       "small value",
			params:     map[string]any{"val": float64(7)},
			key:        "val",
			defaultVal: 0,
			expected:   7,
		},
		{
			name:       "max uint8 value",
			params:     map[string]any{"val": float64(math.MaxUint8)},
			key:        "val",
			defaultVal: 0,
			expected:   math.MaxUint8,
		},
		{
			name:       "value just above max uint8 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 1)},
			key:        "val",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "large value returns default",
			params:     map[string]any{"val": float64(math.MaxUint32)},
			key:        "val",
			defaultVal: 99,
			expected:   99,
		},
		{
			name:       "typical VLAN priority",
			params:     map[string]any{"val": float64(7)},
			key:        "val",
			defaultVal: 0,
			expected:   7,
		},
		{
			name:       "CoS value",
			params:     map[string]any{"val": float64(6)},
			key:        "val",
			defaultVal: 0,
			expected:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint8Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint8Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestModtypesSafeIntToUint32Boundary tests boundary conditions for modtypes.SafeIntToUint32.
func TestModtypesSafeIntToUint32Boundary(t *testing.T) {
	// Test the exact boundaries.
	tests := []struct {
		name     string
		value    int
		expected uint32
	}{
		{
			name:     "value at -1",
			value:    -1,
			expected: 0,
		},
		{
			name:     "value at 0",
			value:    0,
			expected: 0,
		},
		{
			name:     "value at 1",
			value:    1,
			expected: 1,
		},
		{
			name:     "value at MaxUint32",
			value:    math.MaxUint32,
			expected: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.SafeIntToUint32(tt.value)
			if result != tt.expected {
				t.Errorf("modtypes.SafeIntToUint32(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint16ParamBoundary tests boundary conditions for modtypes.GetUint16Param.
func TestModtypesGetUint16ParamBoundary(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		defaultVal uint16
		expected   uint16
	}{
		{
			name:       "value at MaxUint16 - 1",
			params:     map[string]any{"val": float64(math.MaxUint16 - 1)},
			defaultVal: 0,
			expected:   math.MaxUint16 - 1,
		},
		{
			name:       "value at MaxUint16",
			params:     map[string]any{"val": float64(math.MaxUint16)},
			defaultVal: 0,
			expected:   math.MaxUint16,
		},
		{
			name:       "value at MaxUint16 + 1 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 1)},
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "value at MaxUint16 + 2 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 2)},
			defaultVal: 99,
			expected:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint16Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint16Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint8ParamBoundary tests boundary conditions for modtypes.GetUint8Param.
func TestModtypesGetUint8ParamBoundary(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		defaultVal uint8
		expected   uint8
	}{
		{
			name:       "value at MaxUint8 - 1",
			params:     map[string]any{"val": float64(math.MaxUint8 - 1)},
			defaultVal: 0,
			expected:   math.MaxUint8 - 1,
		},
		{
			name:       "value at MaxUint8",
			params:     map[string]any{"val": float64(math.MaxUint8)},
			defaultVal: 0,
			expected:   math.MaxUint8,
		},
		{
			name:       "value at MaxUint8 + 1 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 1)},
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "value at MaxUint8 + 2 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 2)},
			defaultVal: 99,
			expected:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint8Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint8Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}
