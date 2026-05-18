//go:build !cgo


package platform

// cgoEnabled returns false when CGO is disabled at build time.
func cgoEnabled() bool {
	return false
}
