package servicecomposer

import (
	"time"
)

const DefaultModuleInitTimeout = 10 * time.Second

type moduleInitTimeoutOption struct {
	timeout time.Duration
}

func (m *moduleInitTimeoutOption) apply(c *composer) {
	c.moduleInitTimeout = m.timeout
}

// WithModuleInitTimeout sets the timeout for module initialization.
// Defaults to 10 minutes as of [DefaultModuleInitTimeout]
func WithModuleInitTimeout(timeout time.Duration) Option {
	return &moduleInitTimeoutOption{
		timeout: timeout,
	}
}
