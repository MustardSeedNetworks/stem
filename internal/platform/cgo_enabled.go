//go:build cgo

package platform

// cgoEnabled returns true when CGO is enabled at build time.
func cgoEnabled() bool {
	return true
}
