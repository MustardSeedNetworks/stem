// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package platform_test

import (
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/platform"
)

func TestDetect(t *testing.T) {
	info := platform.Detect()

	if info.OS != runtime.GOOS {
		t.Errorf("Detect().OS = %q, want %q", info.OS, runtime.GOOS)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("Detect().Arch = %q, want %q", info.Arch, runtime.GOARCH)
	}

	// On non-Linux, should have requirements.
	if runtime.GOOS != platform.RequiredOS {
		if len(info.Requirements) == 0 {
			t.Error("Detect().Requirements should not be empty on non-Linux")
		}
		if info.IsSupported {
			t.Error("Detect().IsSupported should be false on non-Linux")
		}
	}
}

func TestIsDataplaneSupported(t *testing.T) {
	// This will return false on macOS (development) and true on Linux with CGO.
	result := platform.IsDataplaneSupported()

	// On non-Linux, should be false.
	if runtime.GOOS != platform.RequiredOS && result {
		t.Error("IsDataplaneSupported() should return false on non-Linux")
	}
}

func TestCheckDataplaneSupport(t *testing.T) {
	err := platform.CheckDataplaneSupport()

	// On non-Linux, should return an error.
	if runtime.GOOS != platform.RequiredOS {
		if err == nil {
			t.Error("CheckDataplaneSupport() should return error on non-Linux")
		}

		// Check error unwrapping.
		if !errors.Is(err, platform.ErrPlatformNotSupported) {
			t.Error("CheckDataplaneSupport() error should unwrap to ErrPlatformNotSupported")
		}
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *platform.Error
		wantSub  string
		wantNoSb string
	}{
		{
			name: "with requirements",
			err: &platform.Error{
				Info: &platform.Info{
					OS:           "darwin",
					Arch:         "arm64",
					CGOEnabled:   false,
					IsSupported:  false,
					Requirements: []string{"Linux required", "CGO required"},
				},
				Message: "test error",
			},
			wantSub:  "test error",
			wantNoSb: "",
		},
		{
			name: "empty requirements",
			err: &platform.Error{
				Info: &platform.Info{
					OS:           "linux",
					Arch:         "amd64",
					CGOEnabled:   true,
					IsSupported:  true,
					Requirements: []string{},
				},
				Message: "simple error",
			},
			wantSub:  "simple error",
			wantNoSb: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if len(got) == 0 {
				t.Error("Error() returned empty string")
			}
			if tt.wantSub != "" && !strings.Contains(got, tt.wantSub) {
				t.Errorf("Error() = %q, should contain %q", got, tt.wantSub)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	err := &platform.Error{
		Info: &platform.Info{
			OS:           "darwin",
			Arch:         "arm64",
			CGOEnabled:   false,
			IsSupported:  false,
			Requirements: []string{},
		},
		Message: "test",
	}

	unwrapped := err.Unwrap()
	if !errors.Is(unwrapped, platform.ErrPlatformNotSupported) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, platform.ErrPlatformNotSupported)
	}
}

func TestCGOEnabled(_ *testing.T) {
	// CGO detection is tested indirectly through Detect().CGOEnabled.
	// The actual value depends on build configuration.
	info := platform.Detect()
	// Just verify it returns a valid value (not panicking).
	_ = info.CGOEnabled
}
