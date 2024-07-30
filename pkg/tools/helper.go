package tools

import (
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"io"
)

func PrintWantedReceived(wantBody string, receivedJson []byte, l golog.MyLogger) {
	l.Info("WANTED   :%T - %#v\n", wantBody, wantBody)
	l.Info("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
}

func CloseBody(Body io.ReadCloser, msg string, logger golog.MyLogger) {
	err := Body.Close()
	if err != nil {
		logger.Error("Error: %v in %s doing Body.Close().\n", err, msg)
	}
}
