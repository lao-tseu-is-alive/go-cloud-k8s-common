package config

import (
	"os"
	"reflect"
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

func TestGetListenIpFromEnvOrPanic(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultSrvIp string
		expected     string
		shouldPanic  bool
	}{
		{"Default Localhost IP", "", "127.0.0.1", "127.0.0.1", false},
		{"Default IP", "", "0.0.0.0", "0.0.0.0", false},
		{"Valid IP from env", "192.168.1.1", "127.0.0.1", "192.168.1.1", false},
		{"Invalid IP", "invalid_ip", "127.0.0.1", "", true},
		{"IPv6 address", "2001:db8::1", "127.0.0.1", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("SRV_IP", tt.envValue)
				defer os.Unsetenv("SRV_IP")
			} else {
				os.Unsetenv("SRV_IP")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but didn't get one")
					}
				}()
			}

			result := GetListenIpFromEnvOrPanic(tt.defaultSrvIp)

			if !tt.shouldPanic {
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetAllowedIpsFromEnvOrPanic(t *testing.T) {
	tests := []struct {
		name              string
		envValue          string
		defaultAllowedIps []string
		expected          []string
		shouldPanic       bool
	}{
		{
			name:              "Default IPs",
			envValue:          "",
			defaultAllowedIps: []string{"127.0.0.1", "192.168.1.1"},
			expected:          []string{"127.0.0.1", "192.168.1.1"},
			shouldPanic:       false,
		},
		{
			name:              "Valid IPs from env",
			envValue:          "10.0.0.1, 172.16.0.1",
			defaultAllowedIps: []string{"127.0.0.1"},
			expected:          []string{"10.0.0.1", "172.16.0.1"},
			shouldPanic:       false,
		},
		{
			name:              "Invalid IP in env",
			envValue:          "10.0.0.1, invalid_ip",
			defaultAllowedIps: []string{"127.0.0.1"},
			expected:          nil,
			shouldPanic:       true,
		},
		{
			name:              "Invalid default IP",
			envValue:          "",
			defaultAllowedIps: []string{"invalid_ip"},
			expected:          nil,
			shouldPanic:       true,
		},
		{
			name:              "Empty env and default",
			envValue:          "",
			defaultAllowedIps: []string{},
			expected:          nil,
			shouldPanic:       true,
		},
		{
			name:              "IPv6 addresses",
			envValue:          "2001:db8::1, 2001:db8::2",
			defaultAllowedIps: []string{"::1"},
			expected:          []string{"2001:db8::1", "2001:db8::2"},
			shouldPanic:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ALLOWED_IP", tt.envValue)
				defer os.Unsetenv("ALLOWED_IP")
			} else {
				os.Unsetenv("ALLOWED_IP")
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but didn't get one")
					}
				}()
			}

			result := GetAllowedIpsFromEnvOrPanic(tt.defaultAllowedIps)

			if !tt.shouldPanic {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("Expected %v, but got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestGetAllowedHostsFromEnvOrPanic(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    []string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:     "Valid single IP",
			envValue: "192.168.1.1",
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "Valid multiple IPs",
			envValue: "192.168.1.1, 10.0.0.1, 172.16.0.1",
			expected: []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
		},
		{
			name:     "IPs with extra spaces and empty entries",
			envValue: " 192.168.1.1 ,  , 10.0.0.1 , ",
			expected: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:        "Empty env variable",
			envValue:    "",
			shouldPanic: true,
			panicMsg:    "💥💥 ERROR: ENV ALLOWED_HOSTS should contain your allowed hosts.",
		},
		{
			name:        "Env variable not set",
			shouldPanic: true,
			panicMsg:    "💥💥 ERROR: ENV ALLOWED_HOSTS should contain your allowed hosts.",
		},
		{
			name:        "Only empty entries",
			envValue:    " , , ",
			shouldPanic: true,
			panicMsg:    "💥💥 ERROR: CONFIG ENV ALLOWED_HOSTS should contain at least one valid Host.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ALLOWED_HOSTS", tt.envValue)
				defer os.Unsetenv("ALLOWED_HOSTS")
			} else if !tt.shouldPanic {
				os.Unsetenv("ALLOWED_HOSTS")
			}

			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("GetAllowedHostsFromEnvOrPanic() should have panicked")
					} else if r != tt.panicMsg {
						t.Errorf("GetAllowedHostsFromEnvOrPanic() panicked with %v, want %v", r, tt.panicMsg)
					}
				}()
			}

			result := GetAllowedHostsFromEnvOrPanic()

			if !tt.shouldPanic {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("GetAllowedHostsFromEnvOrPanic() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}
