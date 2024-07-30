package main

import (
	"embed"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/version"
	"io/fs"
	"log"
	"net/http"
)

const (
	APP               = "goCloudK8sCommonDemoServer"
	defaultPort       = 9999
	defaultServerIp   = ""
	defaultServerPath = "/"
	defaultWebRootDir = "front/dist/"
)

// content holds our static web server content.
//
//go:embed all:front/dist
var content embed.FS

func GetMyDefaultHandler(s *gohttp.Server, webRootDir string, content embed.FS) http.HandlerFunc {
	handlerName := "GetMyDefaultHandler"
	logger := s.GetLog()
	logger.Debug("Initial call to %s with webRootDir:%s", handlerName, webRootDir)
	RootPathGetCounter := s.RootPathGetCounter
	// Create a subfolder filesystem to serve only the content of front/dist
	subFS, err := fs.Sub(content, "front/dist")
	if err != nil {
		logger.Fatal("Error creating sub-filesystem: %v", err)
	}

	// Create a file server handler for the embed filesystem
	handler := http.FileServer(http.FS(subFS))

	return func(w http.ResponseWriter, r *http.Request) {
		gohttp.TraceRequest(handlerName, r, logger)
		RootPathGetCounter.Inc()
		handler.ServeHTTP(w, r)
	}
}

func GetProtectedHandler(server *gohttp.Server, l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetProtectedHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		gohttp.TraceRequest(handlerName, r, l)
		// get the user from the context
		claims, err := gohttp.GetJwtCustomClaims(r)
		if err != nil {
			l.Error("Error getting user from context: %v", err)
			http.Error(w, "Error getting user from context", http.StatusInternalServerError)
			return
		}
		currentUserId := claims.Id
		// check if user is admin
		if !claims.IsAdmin {
			l.Error("User %d is not admin: %+v", currentUserId, claims)
			http.Error(w, "User is not admin", http.StatusForbidden)
			return
		}
		// respond with protected data
		err = server.JsonResponse(w, claims)
		if err != nil {
			http.Error(w, "Error responding with protected data", http.StatusInternalServerError)
			return
		}
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
	l.Info("HTTP server will listen : %s", listenAddr)

	jwtInfo := gohttp.JwtInfo{
		Secret:   config.GetJwtSecretFromEnvOrPanic(),
		Duration: config.GetJwtDurationFromEnvOrPanic(60),
	}

	// set local admin user for test
	myAuthenticator := gohttp.NewSimpleAdminAuthenticator(
		config.GetAdminUserFromFromEnvOrPanic("goadmin"),
		config.GetAdminPasswordFromFromEnvOrPanic())
	myVersionWriter := gohttp.NewSimpleVersionWriter(APP, version.VERSION, version.REVISION)

	server := gohttp.NewGoHttpServer(listenAddr, myAuthenticator, myVersionWriter, l)
	// curl -vv  -X GET  -H 'Content-Type: application/json'  http://localhost:9999/time	==>200 OK , {"time":"2024-07-15T15:30:21+02:00"}
	server.AddRoute("GET /hello", gohttp.GetStaticPageHandler("Hello", "Hello World!", l))
	mux := server.GetRouter()
	mux.Handle("POST /login", gohttp.GetLoginPostHandler(server, jwtInfo))
	// Protected endpoint (using jwtMiddleware)
	mux.Handle("GET /protected", gohttp.JwtMiddleware(GetProtectedHandler(server, l), jwtInfo.Secret, l))

	mux.Handle("GET /*", gohttp.NewPrometheusMiddleware(
		server.GetPrometheusRegistry(), nil).
		WrapHandler("GET /*", GetMyDefaultHandler(server, defaultWebRootDir, content)),
	)

	server.StartServer()
}
