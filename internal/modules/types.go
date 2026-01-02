// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import "github.com/krisarmstrong/stem/internal/modules/modtypes"

type (
	// Result is a type alias for modtypes.Result for backward compatibility.
	Result = modtypes.Result
	// TestConfig is a type alias for modtypes.TestConfig for backward compatibility.
	TestConfig = modtypes.TestConfig
)

// Re-export helper functions from common package.
//
//nolint:gochecknoglobals // Re-exported helper functions for convenience.
var (
	GetFloat64Param = modtypes.GetFloat64Param
	GetUint64Param  = modtypes.GetUint64Param
	GetUint32Param  = modtypes.GetUint32Param
	GetIntParam     = modtypes.GetIntParam
)
