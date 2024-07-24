package go_http

import (
	"io"
	"log"
)

func CloseBody(Body io.ReadCloser, msg string, logger *log.Logger) {
	err := Body.Close()
	if err != nil {
		logger.Printf("Error %v in %s doing Body.Close().\n", err, msg)
	}
}
