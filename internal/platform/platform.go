// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package platform provides platform detection utilities for determining
// dataplane capabilities and providing user-friendly error messages.
package platform

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Platform requirements for different capabilities.
const (
	// RequiredOS is the operating system required for dataplane operations.
	RequiredOS = "linux"
)

// Error types for platform-related issues.
var (
	// ErrPlatformNotSupported is returned when the platform doesn't support dataplane.
	ErrPlatformNotSupported = errors.New("platform not supported for dataplane operations")

	// ErrCGORequired is returned when CGO is required but not available.
	ErrCGORequired = errors.New("CGO is required for dataplane operations")
)

// Info holds platform detection results.
type Info struct {
	OS           string
	Arch         string
	CGOEnabled   bool
	IsSupported  bool
	Requirements []string
}

// Detect returns information about the current platform.
func Detect() *Info {
	info := &Info{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		CGOEnabled:   cgoEnabled(),
		IsSupported:  false,
		Requirements: make([]string, 0),
	}

	// Check requirements.
	if info.OS != RequiredOS {
		info.Requirements = append(info.Requirements,
			fmt.Sprintf("Linux required (current: %s)", info.OS))
	}

	if !info.CGOEnabled {
		info.Requirements = append(info.Requirements,
			"CGO must be enabled (build with CGO_ENABLED=1)")
	}

	info.IsSupported = len(info.Requirements) == 0

	return info
}

// IsDataplaneSupported returns true if the current platform supports dataplane.
func IsDataplaneSupported() bool {
	return Detect().IsSupported
}

// CheckDataplaneSupport returns nil if supported, or a detailed error if not.
func CheckDataplaneSupport() error {
	info := Detect()
	if info.IsSupported {
		return nil
	}

	return &Error{
		Info:    info,
		Message: "dataplane operations require Linux with CGO enabled",
	}
}

// Error provides detailed platform compatibility information.
type Error struct {
	Info    *Info
	Message string
}

func (e *Error) Error() string {
	if len(e.Info.Requirements) == 0 {
		return e.Message
	}

	var sb strings.Builder
	sb.WriteString(e.Message)
	sb.WriteString("\n\nPlatform: ")
	sb.WriteString(e.Info.OS)
	sb.WriteString("/")
	sb.WriteString(e.Info.Arch)
	sb.WriteString("\nMissing requirements:")

	for _, req := range e.Info.Requirements {
		sb.WriteString("\n  - ")
		sb.WriteString(req)
	}

	sb.WriteString("\n\nFor development/testing on macOS, use mock mode or run tests in a Linux container.")

	return sb.String()
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *Error) Unwrap() error {
	return ErrPlatformNotSupported
}
