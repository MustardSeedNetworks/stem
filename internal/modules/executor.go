// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import (
	"fmt"
)

// ErrTestNotImplemented is returned when a test type is not yet implemented.
var ErrTestNotImplemented = fmt.Errorf("test type not implemented")

// ErrModuleNotExecutor is returned when trying to execute on a non-executor module.
var ErrModuleNotExecutor = fmt.Errorf("module does not support execution")

// ErrInvalidConfig is returned when test configuration is invalid.
var ErrInvalidConfig = fmt.Errorf("invalid test configuration")
