// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import "github.com/krisarmstrong/stem/internal/modules/modtypes"

// Re-export error types from common package.
var (
	ErrTestNotImplemented = modtypes.ErrTestNotImplemented
	ErrModuleNotExecutor  = modtypes.ErrModuleNotExecutor
	ErrInvalidConfig      = modtypes.ErrInvalidConfig
)
