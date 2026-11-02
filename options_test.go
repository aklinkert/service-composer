package servicecomposer

import (
	"log/slog"
	"testing"
	"time"
)

func TestWithModuleInitTimeout(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option
		wantTimeout time.Duration
	}{
		{
			name:        "should default without option",
			options:     nil,
			wantTimeout: DefaultModuleInitTimeout,
		},
		{
			name:        "should set timeout",
			options:     []Option{WithModuleInitTimeout(5 * time.Second)},
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "should set zero timeout",
			options:     []Option{WithModuleInitTimeout(0)},
			wantTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := New(t.Context(), slog.Default(), "test", tt.options...).(*composer)
			if !ok {
				t.Fatal("New() did not return a *composer")
			}

			if c.moduleInitTimeout != tt.wantTimeout {
				t.Errorf("moduleInitTimeout = %v, want %v", c.moduleInitTimeout, tt.wantTimeout)
			}
		})
	}
}
