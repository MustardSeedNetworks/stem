//go:build !cgo

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package platform

// cgoEnabled returns false when CGO is disabled at build time.
func cgoEnabled() bool {
	return false
}
