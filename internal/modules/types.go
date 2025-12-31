// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package modules provides the module system for The Stem.
// This file re-exports shared types from the common package for convenience.
package modules

import "github.com/krisarmstrong/stem/internal/modules/common"

// Type aliases for backward compatibility - re-export from common package.
type (
	Result     = common.Result
	TestConfig = common.TestConfig
)

// Re-export helper functions from common package.
var (
	GetFloat64Param = common.GetFloat64Param
	GetUint64Param  = common.GetUint64Param
	GetUint32Param  = common.GetUint32Param
	GetIntParam     = common.GetIntParam
)
