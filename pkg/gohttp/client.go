package gohttp

import (
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"io"
)

func CloseBody(Body io.ReadCloser, msg string, logger golog.MyLogger) {
	err := Body.Close()
	if err != nil {
		logger.Error("Error: %v in %s doing Body.Close().\n", err, msg)
	}
}
