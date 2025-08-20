package config

import (
	"os"
	"testing"
)

func TestGetAdminUserFromEnv(t *testing.T) {
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

	tests := []struct {
		name             string
		envValue         string
		defaultAdminUser string
		expected         string
		shouldPanic      bool
	}{
		{"Use default", "", "defaultAdmin", "defaultAdmin", false},
		{"Use env variable", "envAdmin", "defaultAdmin", "envAdmin", false},
		{"UserLogin too short", "a", "adm", "", true},
		{"UserLogin exactly minimum length", "exact", "defaultAdmin", "exact", false},
		{"emoticons characters should be counted as one", "💥⭐🌀🚩", "defaultAdmin", "", true},
		{"emoticons characters should be accepted", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "adm", "🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_USER", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_USER")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetAdminUserFromEnvOrPanic(tt.defaultAdminUser)

			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestGetAdminEmailFromEnv(t *testing.T) {
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

	tests := []struct {
		name              string
		envValue          string
		defaultAdminEmail string
		expected          string
		shouldPanic       bool
	}{
		{"Use default", "", "defaultAdminEmail@toto.ch", "defaultAdminEmail@toto.ch", false},
		{"Use env variable", "envAdmin@toto.ch", "defaultAdminEmail@toto.ch", "envAdmin@toto.ch", false},
		{"UserEmail too short", "a", "adm", "", true},
		{"UserEmail exactly minimum length", "adm@totos.ch", "defaultAdmin", "adm@totos.ch", false},
		{"emoticons should not be allowed", "💥⭐🌀🚩@toto.ch", "defaultAdmin", "", true},
		{"invalid email should not be allowed", "blairoATtoto.ch", "defaultAdmin", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_EMAIL", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_EMAIL")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetAdminEmailFromEnvOrPanic(tt.defaultAdminEmail)

			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestGetAdminIdFromEnvOrPanic(t *testing.T) {
	tests := []struct {
		name            string
		defaultAdminId  int
		envValue        string
		expectedAdminId int
		shouldPanic     bool
	}{
		{
			name:            "Use default when env not set",
			defaultAdminId:  1000,
			envValue:        "",
			expectedAdminId: 1000,
			shouldPanic:     false,
		},
		{
			name:            "Use env value when set",
			defaultAdminId:  1000,
			envValue:        "2000",
			expectedAdminId: 2000,
			shouldPanic:     false,
		},
		{
			name:            "Panic on invalid env value",
			defaultAdminId:  1000,
			envValue:        "invalid",
			expectedAdminId: 0,
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ADMIN_ID", tt.envValue)
				defer os.Unsetenv("ADMIN_ID")
			} else {
				os.Unsetenv("ADMIN_ID")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but didn't get one")
					}
				}()
			}

			result := GetAdminIdFromEnvOrPanic(tt.defaultAdminId)

			if !tt.shouldPanic && result != tt.expectedAdminId {
				t.Errorf("Expected admin ID %d, but got %d", tt.expectedAdminId, result)
			}
		})
	}
}

func TestGetAdminPasswordFromEnvOrPanic(t *testing.T) {
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

	tests := []struct {
		name        string
		envValue    string
		expected    string
		shouldPanic bool
	}{
		{"Valid password", "ValidP@ssw0rd", "ValidP@ssw0rd", false},
		{"Missing env variable", "", "", true},
		{"Password too short", "Short1!", "", true},
		{"Password without lowercase", "PASSWORD123!", "", true},
		{"Password without uppercase", "password123!", "", true},
		{"Password without number", "Password!", "", true},
		{"Password without special char", "Password123", "", true},
		{"Password with invalid char #", "Password123#", "", true},
		{"Password with invalid char |", "Password123|", "", true},
		{"Password with invalid char '", "Password123'", "", true},
		{"Password with space", "Password 123!", "", true},
		{"emoticons characters should be counted as one", "💥⭐🌀🚩✅📣🔆", "", true},
		{"emoticons characters should be accepted", "1Aa🏁❗‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", "1Aa🏁❗‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				setEnv("ADMIN_PASSWORD", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_PASSWORD")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but function did not panic")
					}
				}()
			}

			result := GetAdminPasswordFromEnvOrPanic()

			if !tt.shouldPanic && result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestVerifyPasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid password", "ValidP@ssw0rd", true},
		{"No lowercase", "PASSWORD123!", false},
		{"No uppercase", "password123!", false},
		{"No number", "Password!", false},
		{"No special char", "Password123", false},
		{"With invalid char #", "Password123#", false},
		{"With invalid char |", "Password123|", false},
		{"With invalid char '", "Password123'", false},
		{"With space", "Password 123!", false},
		{"With forbidden chars", "&nrt+\"]or_.r#''!s<Bbv", false},
		{"With special allowed chars", "&1nrt+\"]or_.r!s<Bbv", true},
		{"Too short but complex", "P@ss1", true},
		{"unicode spaces should be banned ", "Aa1💥⭐ 🌀🚩", false},
		{"all kind of emoticons characters should be accepted", "Aa1🏁❗️‼️⁉️⚠️✅❎🔺🔻🔸🔹🔶🔴🔴🔵🔷🔔🔕🚩 🔅🔆✍️✌️👍👆🚀🛎👉🎁📣☀️🔥", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPasswordComplexity(tt.password)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v for password: %s", tt.expected, result, tt.password)
			}
		})
	}
}
