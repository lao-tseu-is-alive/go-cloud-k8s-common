package main

import (
	"log"
	"os"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

const (
	APP             = "simpleServer"
	defaultPort     = 9999
	defaultServerIp = "127.0.0.1"
)

func main() {
	l, err := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, APP)
	if err != nil {
		log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting App:'%s'", APP)
	server := gohttp.CreateNewServerFromEnvOrFail(defaultPort, defaultServerIp, APP, l)

	server.AddRoute("GET /", gohttp.GetStaticPageHandler("Welcome to SimpleServer", "a simple page for your simple server", l))

	server.StartServer()

}
