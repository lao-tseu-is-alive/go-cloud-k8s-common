package gohttp

import (
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
		err := s.Json(w, s.VersionReader.GetVersionInfo(), http.StatusOK, "  ")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
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

// GetJwtChecker returns the jwt checker of this web server
func (s *Server) GetJwtChecker() JwtChecker {
	return s.JwtCheck
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

// Json writes a JSON response with the given status code
func (s *Server) Json(w http.ResponseWriter, result interface{}, statusCode int, indent string) error {
	// Set headers before any potential error occurs
	w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Create JSON encoder
	enc := json.NewEncoder(w)
	if indent != "" {
		enc.SetIndent("", indent)
	}

	// Set status code before encoding
	w.WriteHeader(statusCode)

	// Encode and write response
	if err := enc.Encode(result); err != nil {
		s.logger.Error("JSON encoding failed. Error: %v", err)
		// Note: If encoding fails after partial write, the response may already be sent.
		// Log the error and return it, but the client may see a partial response.
		return err
	}

	return nil
}

// waitForShutdownToExit will wait for interrupt signal SIGINT or SIGTERM and gracefully shutdown the server after secondsToWait seconds.
func waitForShutdownToExit(srv *http.Server, secondsToWait time.Duration) {
	// Create a channel to receive OS signals
	interruptChan := make(chan os.Signal, 1)

	// Register for SIGINT and SIGTERM signals
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to signal when shutdown is complete
	shutdownComplete := make(chan struct{})

	// Start a goroutine to handle shutdown
	go func() {
		// Block until a signal is received
		sig := <-interruptChan
		srv.ErrorLog.Printf("INFO: Signal %v received, initiating graceful shutdown (timeout: %v seconds)...\n",
			sig, secondsToWait.Seconds())

		// Create a context with timeout for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
		defer cancel()

		// Attempt graceful shutdown
		// This stops accepting new connections and waits for existing connections to complete
		if err := srv.Shutdown(ctx); err != nil {
			srv.ErrorLog.Printf("ERROR: Problem during shutdown: %v\n", err)
		}

		// Wait for context to be done (either timeout or cancel)
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				srv.ErrorLog.Println("WARNING: Shutdown timed out, some connections may have been terminated")
			} else {
				srv.ErrorLog.Println("INFO: Shutdown completed successfully")
			}
		}

		// Signal that shutdown is complete
		close(shutdownComplete)
	}()

	// Wait for shutdown to complete
	<-shutdownComplete
	srv.ErrorLog.Println("INFO: Server gracefully stopped, exiting")
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
