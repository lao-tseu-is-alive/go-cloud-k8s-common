package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/version"
)

const (
	APP               = "goCloudK8sCommonDemoServer"
	defaultPort       = 9999
	defaultServerIp   = "0.0.0.0"
	defaultLogName    = "stderr"
	defaultServerPath = "/"
	defaultWebRootDir = "front/dist"
	defaultAdminId    = 99999
	defaultAdminUser  = "goadmin"
	defaultAdminEmail = "goadmin@lausanne.ch"
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
	subFS, err := fs.Sub(content, webRootDir)
	if err != nil {
		// Log error but don't terminate the program
		logger.Error("Error creating fs.Sub for content in %s: %v", webRootDir, err)
		// Return a handler that serves an error message
		return func(w http.ResponseWriter, r *http.Request) {
			gohttp.TraceRequest(handlerName, r, logger)
			RootPathGetCounter.Inc()
			http.Error(w, "Internal Server Error: Could not initialize file system", http.StatusInternalServerError)
		}
	}

	// Create a file server handler for the embed filesystem
	handler := http.FileServer(http.FS(subFS))

	return func(w http.ResponseWriter, r *http.Request) {
		gohttp.TraceRequest(handlerName, r, logger)
		RootPathGetCounter.Inc()
		handler.ServeHTTP(w, r)
	}
}

func GetProtectedHandler(server *gohttp.Server, jwtContextKey string, l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetProtectedHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		gohttp.TraceRequest(handlerName, r, l)
		// get the user from the context

		l.Debug("Get protected data from context %s :%v", jwtContextKey, r.Context().Value(jwtContextKey))
		// get the user from the context ( the user is already in the context thanks to the jwtMiddleware)
		claims := server.GetJwtChecker().GetJwtCustomClaimsFromContext(r.Context())

		currentUserId := claims.User.UserId
		// check if the user is admin
		if !claims.User.IsAdmin {
			l.Error("User %d is not admin: %+v", currentUserId, claims)
			http.Error(w, "User is not admin", http.StatusForbidden)
			return
		}
		// respond with protected data
		err := server.Json(w, claims, http.StatusOK, "")
		if err != nil {
			l.Error("error writing json response to client : %v", err)
			http.Error(w, "Error responding with protected data", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	l, err := golog.NewLogger("simple", config.GetLogWriterFromEnvOrPanic(defaultLogName), golog.DebugLevel, APP)
	if err != nil {
		log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
	}
	l.Info("🚀🚀 Starting App:'%s', ver:%s, build:%s, from: %s", APP, version.VERSION, version.BuildStamp, version.REPOSITORY)
	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl := config.GetJwtAuthUrlFromEnvOrPanic()
	jwtContextKey := config.GetJwtContextKeyFromEnvOrPanic()
	myVersionReader := gohttp.NewSimpleVersionReader(APP, version.VERSION, version.REPOSITORY, version.REVISION, version.BuildStamp, jwtAuthUrl)
	// Create a new JWT checker
	myJwt := gohttp.NewJwtChecker(
		config.GetJwtSecretFromEnvOrPanic(),
		config.GetJwtIssuerFromEnvOrPanic(),
		APP,
		jwtContextKey,
		config.GetJwtDurationFromEnvOrPanic(60),
		l)
	// Create a new Authenticator with a simple admin user
	myAuthenticator := gohttp.NewSimpleAdminAuthenticator(
		config.GetAdminUserFromEnvOrPanic(defaultAdminUser),
		config.GetAdminPasswordFromEnvOrPanic(),
		config.GetAdminEmailFromEnvOrPanic(defaultAdminEmail),
		config.GetAdminIdFromEnvOrPanic(defaultAdminId),
		myJwt)

	// Define your protected routes and their handlers in a map
	protectedAPI := map[string]http.HandlerFunc{
		"GET /api/v1/data": func(w http.ResponseWriter, r *http.Request) {
			claims := myJwt.GetJwtCustomClaimsFromContext(r.Context())
			// Your logic here...
			fmt.Fprintf(w, "Hello, %s! Here is your protected data.", claims.User.UserName)
		},
		"POST /api/v1/submit": func(w http.ResponseWriter, r *http.Request) {
			// Your logic here...
			fmt.Fprintln(w, "Data submitted successfully.")
		},
	}

	server := gohttp.CreateNewServerFromEnvOrFail(
		defaultPort,
		defaultServerIp, APP, l,
		gohttp.WithAuthentication(myAuthenticator),
		gohttp.WithJwtChecker(myJwt),
		gohttp.WithProtectedRoutes(myJwt, protectedAPI),
		gohttp.WithVersionReader(myVersionReader),
	)

	// curl -vv  -X GET  -H 'Content-Type: application/json'  http://localhost:9999/time	==>200 OK , {"time":"2024-07-15T15:30:21+02:00"}
	server.AddRoute("GET /hello", gohttp.GetStaticPageHandler("Hello", "Hello World!", l))
	server.AddRoute("GET /info", gohttp.GetInfoHandler(server))
	server.AddRoute("GET /goAppInfo", gohttp.GetAppInfoHandler(server))

	mux := server.GetRouter()
	mux.Handle(fmt.Sprintf("POST %s", jwtAuthUrl), gohttp.GetLoginPostHandler(server))
	// Protected endpoint (using jwtMiddleware)
	mux.Handle("GET /api/v1/secret", myJwt.JwtMiddleware(GetProtectedHandler(server, jwtContextKey, l)))

	mux.Handle("GET /", gohttp.NewPrometheusMiddleware(
		server.GetPrometheusRegistry(), nil).
		WrapHandler("GET /*", GetMyDefaultHandler(server, defaultWebRootDir, content)),
	)

	server.StartServer()
}
