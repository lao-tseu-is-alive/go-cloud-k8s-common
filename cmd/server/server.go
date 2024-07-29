package main

import (
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/info"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/version"
	"github.com/rs/xid"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	APP               = "goCloudK8sCommonDemoServer"
	defaultPort       = 9999
	defaultServerIp   = ""
	defaultServerPath = "/"
)

func GetMyDefaultHandler(s *gohttp.Server) http.HandlerFunc {
	handlerName := "GetMyDefaultHandler"
	logger := s.GetLog()
	logger.Debug("Initial call to %s", handlerName)

	data := info.CollectRuntimeInfo(APP, version.VERSION, logger)

	return func(w http.ResponseWriter, r *http.Request) {
		gohttp.TraceRequest(handlerName, r, logger)
		query := r.URL.Query()
		nameValue := query.Get("name")
		if nameValue != "" {
			data.ParamName = nameValue
		}
		data.Hostname, _ = os.Hostname()
		data.RemoteAddr = r.RemoteAddr
		data.Headers = r.Header
		data.Uptime = fmt.Sprintf("%s", time.Since(s.GetStartTime()))
		uptimeOS, err := info.GetOsUptime()
		if err != nil {
			logger.Error("💥💥 ERROR: 'GetOsUptime() returned an error : %+#v'", err)
		}
		data.UptimeOs = uptimeOS
		guid := xid.New()
		data.RequestId = guid.String()
		gohttp.RootPathGetCounter.Inc()
		err = s.JsonResponse(w, data)
		if err != nil {
			logger.Error("ERROR:  %v doing JsonResponse in %s, from IP: [%s]\n", err, handlerName, r.RemoteAddr)
			return
		}
		logger.Info("SUCCESS: [%s] from IP: [%s]\n", handlerName, r.RemoteAddr)
	}
}

func main() {

	l, err := golog.NewLogger("zap", golog.DebugLevel, APP)
	if err != nil {
		log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting App %s version:%s from %s", APP, version.VERSION, version.REPOSITORY)
	listenPort := config.GetPortFromEnvOrPanic(defaultPort)
	listenAddr := fmt.Sprintf("%s:%d", defaultServerIp, listenPort)
	l.Info("HTTP server listening %s'", listenAddr)
	server := gohttp.NewGoHttpServer(listenAddr, l)
	// curl -vv  -X POST -H 'Content-Type: application/json'  http://localhost:9999/time   ==> 405 Method Not Allowed,
	// curl -vv  -X GET  -H 'Content-Type: application/json'  http://localhost:9999/time	==>200 OK , {"time":"2024-07-15T15:30:21+02:00"}
	server.AddRoute("GET /hello", gohttp.GetHandlerStaticPage("Hello", "Hello World!", l))
	// using new server Mux in Go 1.22 https://pkg.go.dev/net/http#ServeMux
	mux := server.GetRouter()
	mux.Handle("GET /{$}", gohttp.NewMiddleware(
		server.GetRegistry(), nil).
		WrapHandler("GET /$", GetMyDefaultHandler(server)),
	)
	server.StartServer()
}
