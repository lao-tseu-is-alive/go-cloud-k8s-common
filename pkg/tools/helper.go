package tools

import (
	"log"
)

func PrintWantedReceived(wantBody string, receivedJson []byte, l *log.Logger) {
	l.Printf("WANTED   :%T - %#v\n", wantBody, wantBody)
	l.Printf("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
}
