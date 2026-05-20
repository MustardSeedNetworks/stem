// SPDX-License-Identifier: BUSL-1.1

//go:build !linux && !darwin && !windows

// Lifted from the seed project (internal/truststore); keep in sync.

package truststore

import "context"

func installPlatform(_ context.Context, _ string) (Result, error) {
	return Result{}, ErrUnsupportedPlatform
}

func uninstallPlatform(_ context.Context, _ string) (Result, error) {
	return Result{}, ErrUnsupportedPlatform
}
