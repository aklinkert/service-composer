package servicecomposer

import (
	"testing"
	"time"
)

func Test_moduleInitTimeoutOption_apply(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		wantTimeout time.Duration
	}{
		{
			name:        "should set timeout",
			timeout:     5 * time.Second,
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "should set zero timeout",
			timeout:     0,
			wantTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &moduleInitTimeoutOption{
				timeout: tt.timeout,
			}
			eco := &composer{}

			opt.apply(eco)

			if eco.moduleInitTimeout != tt.wantTimeout {
				t.Errorf("moduleInitTimeoutOption.apply() = %v, want %v", eco.moduleInitTimeout, tt.wantTimeout)
			}
		})
	}
}
