// pkg/config/getTlsInfoFromEnv_test.go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTlsMode(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    string
		shouldPanic bool
	}{
		{"Mode None", "none", "none", false},
		{"Mode Manual", "manual", "manual", false},
		{"Mode Autocert", "autocert", "autocert", false},
		{"Mode Empty (defaults to none)", "", "none", false},
		{"Mode Invalid", "invalid-mode", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TLS_MODE", tt.envValue)
			defer os.Unsetenv("TLS_MODE")

			if tt.shouldPanic {
				assert.Panics(t, func() {
					GetTlsMode()
				})
			} else {
				assert.Equal(t, tt.expected, GetTlsMode())
			}
		})
	}
}
