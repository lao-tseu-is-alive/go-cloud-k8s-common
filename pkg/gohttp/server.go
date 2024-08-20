package gohttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/xid"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is a struct type to store information related to all handlers of web server
type Server struct {
	listenAddress       string
	logger              golog.MyLogger
	router              *http.ServeMux
	registry            *prometheus.Registry
	RootPathGetCounter  prometheus.Counter
	PathNotFoundCounter prometheus.Counter
	startTime           time.Time
	Authenticator       Authentication
	JwtCheck            JwtChecker
	VersionReader       VersionReader
	httpServer          http.Server
}

// NewGoHttpServer is a constructor that initializes the server mux (routes) and all fields of the  Server type
func NewGoHttpServer(listenAddress string, Auth Authentication, JwtCheck JwtChecker, Ver VersionReader, logger golog.MyLogger) *Server {
	myServerMux := http.NewServeMux()
	// Create non-global registry.
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	v := Ver.GetVersionInfo()
	appName := v.App
	RootPathGetCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_root_get_request_count", appName),
			Help: fmt.Sprintf("Number of GET request handled by %s default root handler", appName),
		},
	)

	rootPathNotFoundCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_root_not_found_request_count", appName),
			Help: fmt.Sprintf("Number of page not found handled by %s default root handler", appName),
		},
	)

	registry.MustRegister(RootPathGetCounter)
	registry.MustRegister(rootPathNotFoundCounter)

	var defaultHttpLogger *log.Logger
	defaultHttpLogger, err := logger.GetDefaultLogger()
	if err != nil {
		// in case we cannot get a valid log.Logger for http let's create a reasonable one
		defaultHttpLogger = log.New(os.Stderr, "NewGoHttpServer::defaultHttpLogger", log.Ldate|log.Ltime|log.Lshortfile)
	}

	myServer := Server{
		listenAddress:       listenAddress,
		logger:              logger,
		router:              myServerMux,
		registry:            registry,
		startTime:           time.Now(),
		RootPathGetCounter:  RootPathGetCounter,
		PathNotFoundCounter: rootPathNotFoundCounter,
		Authenticator:       Auth,
		JwtCheck:            JwtCheck,
		VersionReader:       Ver,
		httpServer: http.Server{
			Addr:         listenAddress,       // configure the bind address
			Handler:      myServerMux,         // set the http mux
			ErrorLog:     defaultHttpLogger,   // set the logger for the server
			ReadTimeout:  defaultReadTimeout,  // max time to read request from the client
			WriteTimeout: defaultWriteTimeout, // max time to write response to the client
			IdleTimeout:  defaultIdleTimeout,  // max time for connections using TCP Keep-Alive
		},
	}
	myServer.routes()

	return &myServer
}

// CreateNewServerFromEnvOrFail creates a new server from environment variables or fails
func CreateNewServerFromEnvOrFail(
	defaultPort int,
	defaultServerIp string,
	myAuthenticator Authentication,
	myJwt JwtChecker,
	myVersionReader VersionReader,
	l golog.MyLogger,
) *Server {
	listenPort := config.GetPortFromEnvOrPanic(defaultPort)
	listenAddr := fmt.Sprintf("%s:%d", defaultServerIp, listenPort)
	l.Info("HTTP server will listen : %s", listenAddr)

	server := NewGoHttpServer(listenAddr, myAuthenticator, myJwt, myVersionReader, l)
	return server

}

// (*Server) routes initializes all the default handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *Server) routes() {

	s.router.Handle("GET /time", GetTimeHandler(s.logger))
	s.router.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		TraceRequest("GetVersionHandler", r, s.logger)
		err := s.JsonResponse(w, s.VersionReader.GetVersionInfo())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
	s.router.Handle("GET /readiness", GetReadinessHandler(s.logger))
	s.router.Handle("GET /health", GetHealthHandler(s.logger))
	//expose the default prometheus metrics for Go applications
	s.router.Handle("GET /metrics", NewPrometheusMiddleware(
		s.registry, nil).
		WrapHandler("GET /metrics", promhttp.HandlerFor(
			s.registry,
			promhttp.HandlerOpts{}),
		))

	s.router.Handle("GET /...", GetHandlerNotFound(s.logger, s.PathNotFoundCounter))
}

// AddRoute   adds a handler for this web server
func (s *Server) AddRoute(pathPattern string, handler http.Handler) {
	s.router.Handle(pathPattern, handler)
}

// GetRouter returns the ServeMux of this web server
func (s *Server) GetRouter() *http.ServeMux {
	return s.router
}

// GetPrometheusRegistry returns the Prometheus registry of this web server
func (s *Server) GetPrometheusRegistry() *prometheus.Registry {
	return s.registry
}

// GetLog returns the log of this web server
func (s *Server) GetLog() golog.MyLogger {
	return s.logger
}

// GetStartTime returns the start time of this web server
func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *Server) StartServer() {

	// Starting the web server in his own goroutine
	go func() {
		s.logger.Debug("http server listening at %s://%s/", defaultProtocol, s.listenAddress)
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("💥💥 ERROR: 'Could not listen on %q: %s'\n", s.listenAddress, err)
		}
	}()
	s.logger.Debug("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)

}

func (s *Server) JsonResponse(w http.ResponseWriter, result interface{}) error {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("JSON marshal failed. Error: %v", err)
		return err
	}
	var prettyOutput bytes.Buffer
	err = json.Indent(&prettyOutput, body, "", "  ")
	if err != nil {
		s.logger.Error("JSON Indent failed. Error: %v", err)
		return err
	}
	w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(prettyOutput.Bytes())
	if err != nil {
		s.logger.Error("w.Write failed. Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	return nil
}

// WaitForHttpServer attempts to establish a TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(listenAddress string, waitDuration time.Duration, numRetries int) {
	log.Printf("INFO: 'WaitForHttpServer Will wait for server to be up at %s for %v seconds, with %d retries'\n", listenAddress, waitDuration.Seconds(), numRetries)
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		//conn, err := net.DialTimeout("tcp", listenAddress, dialTimeout)
		resp, err := httpClient.Get(listenAddress)
		if err != nil {
			fmt.Printf("\n[%d] Cannot make http get %s: %v\n", i, listenAddress, err)
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		fmt.Printf("OK: Server responded after %d retries, with status code %d ", i, resp.StatusCode)
		return
	}
	log.Fatalf("Server %s not ready up after %d attempts", listenAddress, numRetries)
}

// waitForShutdownToExit will wait for interrupt signal SIGINT or SIGTERM and gracefully shutdown the server after secondsToWait seconds.
func waitForShutdownToExit(srv *http.Server, secondsToWait time.Duration) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	// wait for SIGINT (interrupt) 	: ctrl + C keypress, or in a shell : kill -SIGINT processId
	sig := <-interruptChan
	srv.ErrorLog.Printf("INFO: 'SIGINT %d interrupt signal received, about to shut down server after max %v seconds...'\n", sig, secondsToWait.Seconds())

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
	defer cancel()
	// gracefully shuts down the server without interrupting any active connections
	// as long as the actives connections last less than shutDownTimeout
	// https://pkg.go.dev/net/http#Server.Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		srv.ErrorLog.Printf("💥💥 ERROR: 'Problem doing Shutdown %v'\n", err)
	}
	<-ctx.Done()
	srv.ErrorLog.Println("INFO: 'Server gracefully stopped, will exit'")
	os.Exit(0)
}

func getHtmlHeader(title string, description string) string {
	return fmt.Sprintf("%s<meta name=\"description\" content=\"%s\"><title>%s</title></head>", htmlHeaderStart, description, title)
}

func getHtmlPage(title string, description string) string {
	return getHtmlHeader(title, description) +
		fmt.Sprintf("\n<body><div class=\"container\"><h4>%s</h4></div></body></html>", title)
}
func TraceRequest(handlerName string, r *http.Request, l golog.MyLogger) {
	const formatTraceRequest = "TraceRequest:[%s] %s '%s', RemoteIP: [%s],id:%s\n"
	remoteIp := r.RemoteAddr // ip address of the original request or the last proxy
	requestedUrlPath := r.URL.Path
	guid := xid.New()
	l.Debug(formatTraceRequest, handlerName, r.Method, requestedUrlPath, remoteIp, guid.String())
}
