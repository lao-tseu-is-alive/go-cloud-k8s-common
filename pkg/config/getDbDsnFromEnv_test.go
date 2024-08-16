package config

import (
	"fmt"
	"os"
	"testing"
)

func TestGetPgDbDsnUrlFromEnvOrPanic(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		defaultIP      string
		defaultPort    int
		defaultDbName  string
		defaultDbUser  string
		defaultSSL     string
		expectedResult string
		shouldPanic    bool
	}{
		{
			name: "All environment variables set",
			envVars: map[string]string{
				"DB_HOST":     "192.168.1.1",
				"DB_PORT":     "5432",
				"DB_NAME":     "testdb",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_SSL_MODE": "disable",
			},
			defaultIP:      "127.0.0.1",
			defaultPort:    5433,
			defaultDbName:  "defaultdb",
			defaultDbUser:  "defaultuser",
			defaultSSL:     "prefer",
			expectedResult: "postgres://testuser:testpass@192.168.1.1:5432/testdb?sslmode=disable",
		},
		{
			name:           "Using default values",
			envVars:        map[string]string{"DB_PASSWORD": "testpass"},
			defaultIP:      "127.0.0.1",
			defaultPort:    5433,
			defaultDbName:  "defaultdb",
			defaultDbUser:  "defaultuser",
			defaultSSL:     "prefer",
			expectedResult: "postgres://defaultuser:testpass@127.0.0.1:5433/defaultdb?sslmode=prefer",
		},
		{
			name: "Invalid DB_PORT",
			envVars: map[string]string{
				"DB_PORT":     "invalid",
				"DB_PASSWORD": "testpass",
			},
			defaultIP:   "127.0.0.1",
			defaultPort: 5433,
			shouldPanic: true,
		},
		{
			name: "DB_PORT out of range",
			envVars: map[string]string{
				"DB_PORT":     "70000",
				"DB_PASSWORD": "testpass",
			},
			defaultIP:   "127.0.0.1",
			defaultPort: 5433,
			shouldPanic: true,
		},
		{
			name: "Invalid DB_HOST",
			envVars: map[string]string{
				"DB_HOST":     "invalid-ip",
				"DB_PASSWORD": "testpass",
			},
			defaultIP:   "127.0.0.1",
			defaultPort: 5433,
			shouldPanic: true,
		},
		{
			name:        "Missing DB_PASSWORD",
			envVars:     map[string]string{},
			defaultIP:   "127.0.0.1",
			defaultPort: 5433,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				err := os.Setenv(k, v)
				if err != nil {
					panic(fmt.Errorf("💥💥 unable to do a os.Setenv. error: %v", err))
				}
			}

			// Defer cleanup of environment variables
			defer func() {
				for k := range tt.envVars {
					err := os.Unsetenv(k)
					if err != nil {
						panic(fmt.Errorf("💥💥 unable to do a os.Unsetenv. error: %v", err))
					}
				}
			}()

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but didn't get one")
					}
				}()
			}

			result := GetPgDbDsnUrlFromEnvOrPanic(tt.defaultIP, tt.defaultPort, tt.defaultDbName, tt.defaultDbUser, tt.defaultSSL)

			if !tt.shouldPanic && result != tt.expectedResult {
				t.Errorf("Expected %s, but got %s", tt.expectedResult, result)
			}
		})
	}
}
