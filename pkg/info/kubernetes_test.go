package info

import (
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

const (
	DEBUG = true
)

var l golog.MyLogger

func TestGetJsonFromUrl(t *testing.T) {
	type args struct {
		url           string
		bearerToken   string
		caCert        []byte
		allowInsecure bool
		readTimeout   time.Duration
		logger        golog.MyLogger
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJsonFromUrl(tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.allowInsecure, tt.args.readTimeout, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetJsonFromUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetJsonFromUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKubernetesApiUrlFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetK8SApiUrlFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetK8SApiUrlFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetK8SApiUrlFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKubernetesConnInfo(t *testing.T) {

	type args struct {
		logger golog.MyLogger
	}
	tests := []struct {
		name    string
		args    args
		want    *K8sInfo
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should return empty strings and an error when K8S_SERVICE_HOST is not set",
			args: args{logger: l},
			want: &K8sInfo{
				CurrentNamespace: "",
				Version:          "",
				Token:            "",
				CaCert:           "",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetK8SConnInfo(tt.args.logger)
			if !tt.wantErr(t, err, fmt.Sprintf("GetK8SConnInfo() %s", tt.name)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetK8SConnInfo(%v)", tt.args.logger)
		})
	}
}

func TestGetKubernetesInfo(t *testing.T) {
	type args struct {
		l golog.MyLogger
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name:  "should return empty strings when not inside a k8s cluster",
			args:  args{l},
			want:  "",
			want1: "",
			want2: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := GetK8sInfo(tt.args.l)
			if got != tt.want {
				t.Errorf("GetK8sInfo() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetK8sInfo() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetK8sInfo() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetKubernetesLatestVersion(t *testing.T) {
	type args struct {
		logger golog.MyLogger
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetK8SLatestVersion(tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetK8SLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetK8SLatestVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func init() {
	var err error
	if DEBUG {
		l, err = golog.NewLogger("zap", golog.DebugLevel, "test_kubernetes")
		if err != nil {
			log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
		}
	} else {
		l, err = golog.NewLogger("zap", golog.ErrorLevel, "test_kubernetes")
		if err != nil {
			log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
		}
	}
}
