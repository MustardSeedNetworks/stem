// SPDX-License-Identifier: BUSL-1.1

package certify_test

import (
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/certify"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// TestNewExecutorWithContextExternal verifies executor creation with an existing context.
func TestNewExecutorWithContextExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)
	if exec == nil {
		t.Fatal("NewExecutorWithContext returned nil")
	}

	// Verify the executor has the correct module info.
	if exec.Name() != certify.ModuleName {
		t.Errorf("Expected module name %q, got %q", certify.ModuleName, exec.Name())
	}
}

// TestExecutorSupportsExecutionExternal verifies the executor supports execution.
func TestExecutorSupportsExecutionExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() returned false, expected true")
	}
}

// TestExecutorCloseExternal verifies the Close method doesn't panic.
func TestExecutorCloseExternal(_ *testing.T) {
	exec := certify.NewExecutorWithContext(nil)
	// Close should not panic with nil context.
	exec.Close()

	// Calling Close again should also not panic.
	exec.Close()
}

// TestExecuteInvalidTestTypeExternal verifies Execute rejects unknown test types.
func TestExecuteInvalidTestTypeExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	invalidTests := []string{
		"invalid_test",
		"rfc2544_throughput",
		"y1564",
		"",
		"rfc2889", // Missing suffix.
	}

	for _, testType := range invalidTests {
		t.Run(testType, func(t *testing.T) {
			_, err := exec.Execute(testType, cfg)
			if err == nil {
				t.Errorf("Execute(%q) should return error for invalid test type", testType)
			}
		})
	}
}

// TestExecuteNilConfigExternal verifies Execute rejects nil configuration.
func TestExecuteNilConfigExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)

	// All valid test types should fail with nil config.
	validTests := []string{
		"rfc2889_forwarding",
		"rfc6349_throughput",
		"tsn_timing",
	}

	for _, testType := range validTests {
		t.Run(testType, func(t *testing.T) {
			_, err := exec.Execute(testType, nil)
			if err == nil {
				t.Errorf("Execute(%q, nil) should return error", testType)
			}
			if !errors.Is(err, modtypes.ErrInvalidConfig) {
				t.Errorf("Expected ErrInvalidConfig, got: %v", err)
			}
		})
	}
}

// TestExecutorModuleMethodsExternal tests that Module methods work through Executor.
func TestExecutorModuleMethodsExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)

	// Name.
	if exec.Name() != certify.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), certify.ModuleName)
	}

	// DisplayName.
	if exec.DisplayName() != certify.DisplayName {
		t.Errorf("DisplayName() = %q, want %q", exec.DisplayName(), certify.DisplayName)
	}

	// Color.
	if exec.Color() != certify.ColorHex {
		t.Errorf("Color() = %q, want %q", exec.Color(), certify.ColorHex)
	}

	// Standard.
	if exec.Standard() != certify.StandardRef {
		t.Errorf("Standard() = %q, want %q", exec.Standard(), certify.StandardRef)
	}

	// TestTypes.
	testTypes := exec.TestTypes()
	expectedCount := 11
	if len(testTypes) != expectedCount {
		t.Errorf("TestTypes() returned %d types, want %d", len(testTypes), expectedCount)
	}

	// CanRun valid types.
	validTypes := []string{
		"rfc2889_forwarding",
		"rfc2889_caching",
		"rfc6349_throughput",
		"tsn_timing",
		"tsn",
	}
	for _, tt := range validTypes {
		if !exec.CanRun(tt) {
			t.Errorf("CanRun(%q) = false, want true", tt)
		}
	}

	// CanRun invalid types.
	invalidTypes := []string{"invalid", "rfc2544_throughput", "y1564"}
	for _, tt := range invalidTypes {
		if exec.CanRun(tt) {
			t.Errorf("CanRun(%q) = true, want false", tt)
		}
	}
}

// TestExecutorDescriptionExternal verifies Description is non-empty.
func TestExecutorDescriptionExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)
	desc := exec.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

// TestExecutorTestDescriptionExternal verifies TestDescription returns correct values.
func TestExecutorTestDescriptionExternal(t *testing.T) {
	exec := certify.NewExecutorWithContext(nil)

	testCases := []struct {
		testType    string
		shouldExist bool
	}{
		{"rfc2889_forwarding", true},
		{"rfc6349_throughput", true},
		{"tsn_timing", true},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.testType, func(t *testing.T) {
			desc := exec.TestDescription(tc.testType)
			if tc.shouldExist && desc == "" {
				t.Errorf("TestDescription(%q) returned empty string", tc.testType)
			}
			if !tc.shouldExist && desc != "" {
				t.Errorf("TestDescription(%q) = %q, want empty", tc.testType, desc)
			}
		})
	}
}
