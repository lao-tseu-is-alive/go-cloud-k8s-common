package go_http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/version"
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

// GoHttpServer is a struct type to store information related to all handlers of web server
type GoHttpServer struct {
	listenAddress string
	// later we will store here the connection to database
	//DB  *db.Conn
	logger     *log.Logger
	router     *http.ServeMux
	registry   *prometheus.Registry
	startTime  time.Time
	httpServer http.Server
}

// NewGoHttpServer is a constructor that initializes the server mux (routes) and all fields of the  GoHttpServer type
func NewGoHttpServer(listenAddress string, logger *log.Logger) *GoHttpServer {
	myServerMux := http.NewServeMux()
	// Create non-global registry.
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	registry.MustRegister(RootPathGetCounter)
	registry.MustRegister(rootPathNotFoundCounter)

	myServer := GoHttpServer{
		listenAddress: listenAddress,
		logger:        logger,
		router:        myServerMux,
		registry:      registry,
		startTime:     time.Now(),
		httpServer: http.Server{
			Addr:         listenAddress,       // configure the bind address
			Handler:      myServerMux,         // set the http mux
			ErrorLog:     logger,              // set the logger for the server
			ReadTimeout:  defaultReadTimeout,  // max time to read request from the client
			WriteTimeout: defaultWriteTimeout, // max time to write response to the client
			IdleTimeout:  defaultIdleTimeout,  // max time for connections using TCP Keep-Alive
		},
	}
	myServer.routes()

	return &myServer
}

// (*GoHttpServer) routes initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) routes() {

	s.router.Handle("GET /time", GetTimeHandler(s.logger))
	s.router.Handle("GET /wait", GetWaitHandler(defaultSecondsToSleep, s.logger))
	s.router.Handle("GET /readiness", GetReadinessHandler(s.logger))
	s.router.Handle("GET /health", GetHealthHandler(s.logger))
	//expose the default prometheus metrics for Go applications
	s.router.Handle("GET /metrics", NewMiddleware(
		s.registry, nil).
		WrapHandler("GET /metrics", promhttp.HandlerFor(
			s.registry,
			promhttp.HandlerOpts{}),
		))

	s.router.Handle("GET /...", GetHandlerNotFound(s.logger))
}

// AddRoute   adds a handler for this web server
func (s *GoHttpServer) AddRoute(pathPattern string, handler http.Handler) {
	s.router.Handle(pathPattern, handler)
}

// GetRouter returns the ServeMux of this web server
func (s *GoHttpServer) GetRouter() *http.ServeMux {
	return s.router
}

// GetRegistry returns the Prometheus registry of this web server
func (s *GoHttpServer) GetRegistry() *prometheus.Registry {
	return s.registry
}

// GetLog returns the log of this web server
func (s *GoHttpServer) GetLog() *log.Logger {
	return s.logger
}

// GetStartTime returns the start time of this web server
func (s *GoHttpServer) GetStartTime() time.Time {
	return s.startTime
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) StartServer() {

	// Starting the web server in his own goroutine
	go func() {
		s.logger.Printf("INFO: Starting http server listening at %s://%s/", defaultProtocol, s.listenAddress)
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalf("💥💥 ERROR: 'Could not listen on %q: %s'\n", s.listenAddress, err)
		}
	}()
	s.logger.Printf("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)

}

func (s *GoHttpServer) JsonResponse(w http.ResponseWriter, result interface{}) error {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Printf("ERROR: 'JSON marshal failed. Error: %v'", err)
		return err
	}
	var prettyOutput bytes.Buffer
	err = json.Indent(&prettyOutput, body, "", "  ")
	if err != nil {
		s.logger.Printf("ERROR: 'JSON Indent failed. Error: %v'", err)
		return err
	}
	w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(prettyOutput.Bytes())
	if err != nil {
		s.logger.Printf("ERROR: 'w.Write failed. Error: %v'", err)
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

var RootPathGetCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_root_get_request_count", version.APP),
		Help: fmt.Sprintf("Number of GET request handled by %s default root handler", version.APP),
	},
)

var rootPathNotFoundCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_root_not_found_request_count", version.APP),
		Help: fmt.Sprintf("Number of page not found handled by %s default root handler", version.APP),
	},
)

func getHtmlHeader(title string, description string) string {
	return fmt.Sprintf("%s<meta name=\"description\" content=\"%s\"><title>%s</title></head>", htmlHeaderStart, description, title)
}

func getHtmlPage(title string, description string) string {
	return getHtmlHeader(title, description) +
		fmt.Sprintf("\n<body><div class=\"container\"><h4>%s</h4></div></body></html>", title)
}
func TraceRequest(handlerName string, r *http.Request, l *log.Logger) {
	const formatTraceRequest = "TraceRequest:[%s] %s '%s', RemoteIP: [%s],id:%s\n"
	remoteIp := r.RemoteAddr // ip address of the original request or the last proxy
	requestedUrlPath := r.URL.Path
	guid := xid.New()
	l.Printf(formatTraceRequest, handlerName, r.Method, requestedUrlPath, remoteIp, guid.String())
}
