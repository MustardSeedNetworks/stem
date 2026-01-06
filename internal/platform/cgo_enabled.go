//go:build cgo

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package platform

// cgoEnabled returns true when CGO is enabled at build time.
func cgoEnabled() bool {
	return true
}
