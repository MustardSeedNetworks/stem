// SPDX-License-Identifier: BUSL-1.1

package modules

import "github.com/krisarmstrong/stem/internal/services/modtypes"

// Re-export error types from common package.
var (
	ErrTestNotImplemented = modtypes.ErrTestNotImplemented
	ErrModuleNotExecutor  = modtypes.ErrModuleNotExecutor
	ErrInvalidConfig      = modtypes.ErrInvalidConfig
)
