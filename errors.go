package servicecomposer

import (
	"errors"
)

var (
	ErrModuleNotFound  = errors.New("module not found")
	ErrFailedStart     = errors.New("failed to start module")
	ErrDependencyCycle = errors.New("dependency cycle detected")
)
