package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetPortFromEnv(t *testing.T) {
	type args struct {
		defaultPort int
	}
	tests := []struct {
		name          string
		args          args
		envPORT       string
		want          string
		wantErr       assert.ErrorAssertionFunc
		wantErrPrefix string
	}{
		{
			name: "should return the default values when env variables are not set",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "",
			want:          ":8080",
			wantErr:       assert.NoError,
			wantErrPrefix: "",
		},
		{
			name: "should return SERVERIP:PORT when env variables are set to valid values",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "3333",
			want:          ":3333",
			wantErr:       assert.NoError,
			wantErrPrefix: "",
		},
		{
			name: "should return an empty string and report an error when PORT is not a number",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "aBigOne",
			want:          "",
			wantErr:       assert.Error,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain a valid integer.",
		},
		{
			name: "should return an empty string and report an error when PORT is < 1",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "0",
			want:          "",
			wantErr:       assert.Error,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
		},
		{
			name: "should return an empty string and report an error when PORT is > 65535",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "70000",
			want:          "",
			wantErr:       assert.Error,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envPORT) > 0 {
				err := os.Setenv("PORT", tt.envPORT)
				if err != nil {
					t.Errorf("Unable to set env variable PORT")
					return
				}
			} else {
				// we do not want that an external setting of PORT breaks this test
				err := os.Unsetenv("PORT")
				if err != nil {
					t.Errorf("Unable to unset env variable PORT")
					return
				}
			}
			got, err := GetPortFromEnv(tt.args.defaultPort)
			if !tt.wantErr(t, err, fmt.Sprintf("GetPortFromEnv() error = %v, wantErr %v", err, tt.wantErrPrefix)) {
				return
			}

			if got != tt.want {
				t.Errorf("GetPortFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
