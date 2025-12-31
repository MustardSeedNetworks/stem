// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import "github.com/krisarmstrong/stem/internal/modules/common"

// Re-export error types from common package.
var (
	ErrTestNotImplemented = common.ErrTestNotImplemented
	ErrModuleNotExecutor  = common.ErrModuleNotExecutor
	ErrInvalidConfig      = common.ErrInvalidConfig
)
