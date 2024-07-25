package tools

import (
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

func PrintWantedReceived(wantBody string, receivedJson []byte, l golog.MyLogger) {
	l.Info("WANTED   :%T - %#v\n", wantBody, wantBody)
	l.Info("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
}
