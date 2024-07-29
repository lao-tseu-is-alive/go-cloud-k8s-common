package config

import (
	"os"
	"testing"
)

func TestGetPortFromEnvOrPanic(t *testing.T) {
	// Helper function to set and unset environment variables
	setEnv := func(key, value string) {
		oldValue, exists := os.LookupEnv(key)
		os.Setenv(key, value)
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	// Test cases
	tests := []struct {
		name        string
		envValue    string
		defaultPort int
		expected    int
		shouldPanic bool
	}{
		{"Default port", "", 8080, 8080, false},
		{"Valid port from env", "3000", 8080, 3000, false},
		{"Invalid port (non-integer)", "abc", 8080, 0, true},
		{"Invalid port (too low)", "0", 8080, 0, true},
		{"Invalid port (too high)", "65536", 8080, 0, true},
		{"Valid port (min)", "1", 8080, 1, false},
		{"Valid port (max)", "65535", 8080, 65535, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("PORT", tt.envValue)
			} else {
				os.Unsetenv("PORT")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetPortFromEnvOrPanic(tt.defaultPort)

			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("Expected %d, but got %d", tt.expected, result)
			}
		})
	}
}
