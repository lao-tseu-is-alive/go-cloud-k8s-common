package config

import (
	"os"
	"testing"
)

func TestGetJwtDurationFromEnvOrPanic(t *testing.T) {
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
		name            string
		envValue        string
		defaultDuration int
		expected        int
		shouldPanic     bool
	}{
		{"Default duration", "", 60, 60, false},
		{"Valid duration from env", "120", 60, 120, false},
		{"Invalid duration (non-integer)", "abc", 60, 0, true},
		{"Invalid duration (too low)", "0", 60, 0, true},
		{"Invalid duration (too high)", "14401", 60, 0, true},
		{"Valid duration (min)", "1", 60, 1, false},
		{"Valid duration (max)", "1440", 60, 1440, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_DURATION_MINUTES", tt.envValue)
			} else {
				os.Unsetenv("JWT_DURATION_MINUTES")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetJwtDurationFromEnvOrPanic(tt.defaultDuration)

			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("Expected %d, but got %d", tt.expected, result)
			}
		})
	}
}

func TestGetJwtSecretFromEnvOrPanic(t *testing.T) {
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
		expected    string
		shouldPanic bool
	}{
		{"Valid secret", "validSecretLongEnough", "validSecretLongEnough", false},
		{"Missing env variable", "", "", true},
		{"Secret too short", "short", "", true},
		{"Secret exactly minimum length", "a2b4c6t8a2b4c6t8", "a2b4c6t8a2b4c6t8", false}, // Assuming minSecretLength is 1
		{"Secret with not enough special characters", "!@#$'%^&*()", "!@#$'%^&*()", true},
		{"emoticons characters should be counted as one", "✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "", true},
		{"emoticons characters should be accepted", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_SECRET", tt.envValue)
			} else {
				os.Unsetenv("JWT_SECRET")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetJwtSecretFromEnvOrPanic()

			if !tt.shouldPanic {
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetJwtIssuerFromEnvOrPanic(t *testing.T) {
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
		expected    string
		shouldPanic bool
	}{
		{"Valid issuer", "validSecretLongEnough", "validSecretLongEnough", false},
		{"Missing env variable", "", "", true},
		{"issuer id too short", "short", "", true},
		{"issuer is exactly minimum length", "a2b4c6t8a2b4c6t8", "a2b4c6t8a2b4c6t8", false}, // Assuming minSecretLength is 1
		{"issuer with special characters", "!@#$'%^&*()", "!@#$'%^&*()", true},
		{"emoticons characters should be counted as one", "✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "", true},
		{"emoticons characters should be accepted", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_ISSUER_ID", tt.envValue)
			} else {
				os.Unsetenv("JWT_ISSUER_ID")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetJwtIssuerFromEnvOrPanic()

			if !tt.shouldPanic {
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetJwtContextKeyFromEnvOrPanic(t *testing.T) {
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
		expected    string
		shouldPanic bool
	}{
		{"Valid context key", "validContextKey", "validContextKey", false},
		{"Missing env variable", "", "", true},
		{"context key too short", "short", "", true},
		{"context key is exactly minimum length", "abcDef", "abcDef", false},   // Assuming minSecretLength is 1
		{"context key should contain only letters ", "a2b4c6", "a2b4c6", true}, // Assuming minSecretLength is 1
		{"issuer with special characters", "!@#$'%^&*()", "!@#$'%^&*()", true},
		{"emoticons characters should be refused", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("JWT_CONTEXT_KEY", tt.envValue)
			} else {
				os.Unsetenv("JWT_CONTEXT_KEY")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetJwtContextKeyFromEnvOrPanic()

			if !tt.shouldPanic {
				if result != tt.expected {
					t.Errorf("GetJwtContextKeyFromEnvOrPanic() Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}
