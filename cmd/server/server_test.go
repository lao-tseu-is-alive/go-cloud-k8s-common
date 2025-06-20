package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/info"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/version"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	DEBUG                           = true
	assertCorrectStatusCodeExpected = "expected status code should be returned"
	fmtErrNewRequest                = "### ERROR http.NewRequest %s on [%s] error is :%v\n"
	fmtTraceInfo                    = "### %s : %s on %s\n"
	fmtErr                          = "### GOT ERROR : %s\n%s"
	msgRespNotExpected              = "Response should contain what was expected."
)

var l golog.MyLogger

type TestMainStruct struct {
	name                         string
	contentType                  string
	wantStatusCode               int
	wantBody                     string
	paramKeyValues               map[string]string
	httpMethod                   string
	url                          string
	useFormUrlencodedContentType bool
	body                         string
}

func TestGoHttpServerMyDefaultHandler(t *testing.T) {
	var nameParameter string
	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl := config.GetJwtAuthUrlFromEnvOrPanic()
	myVersionReader := gohttp.NewSimpleVersionReader(APP, version.VERSION, version.REPOSITORY, version.REVISION, version.BuildStamp, jwtAuthUrl)
	// Create a new JWT checker
	myJwt := gohttp.NewJwtChecker(
		config.GetJwtSecretFromEnvOrPanic(),
		config.GetJwtIssuerFromEnvOrPanic(),
		APP,
		config.GetJwtContextKeyFromEnvOrPanic(),
		config.GetJwtDurationFromEnvOrPanic(60),
		l)
	// Create a new Authenticator with a simple admin user
	myAuthenticator := gohttp.NewSimpleAdminAuthenticator(
		config.GetAdminUserFromFromEnvOrPanic(defaultAdminUser),
		config.GetAdminPasswordFromFromEnvOrPanic(),
		config.GetAdminEmailFromFromEnvOrPanic(defaultAdminEmail),
		config.GetAdminIdFromFromEnvOrPanic(defaultAdminId),
		myJwt)
	myServer := gohttp.CreateNewServerFromEnvOrFail(
		defaultPort,
		defaultServerIp,
		myAuthenticator,
		myJwt,
		myVersionReader,
		l)

	ts := httptest.NewServer(GetMyDefaultHandler(myServer, defaultWebRootDir, content))
	defer ts.Close()

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}
	type testStruct struct {
		name           string
		wantStatusCode int
		wantBody       string
		paramKeyValues map[string]string
		r              *http.Request
	}
	tests := []testStruct{

		{
			name:           "2: Get on default Server Path should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "goCloudK8sExampleFront",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, defaultServerPath, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(gohttp.HeaderContentType, gohttp.MIMEHtml)
			if len(tt.paramKeyValues) > 0 {
				parameters := tt.r.URL.Query()
				for paramName, paramValue := range tt.paramKeyValues {
					parameters.Add(paramName, paramValue)
					if paramName == "name" {
						nameParameter = paramValue
					}
				}
				tt.r.URL.RawQuery = parameters.Encode()
			}
			resp, err := http.DefaultClient.Do(tt.r)
			l.Debug(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			defer tools.CloseBody(resp.Body, tt.name, l)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			received, _ := io.ReadAll(resp.Body)
			l.Debug("param name : % v", nameParameter)
			tools.PrintWantedReceived(tt.wantBody, received, l)
			if tt.wantStatusCode == http.StatusOK {
				if tt.wantBody != "" {
					// check that received contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
					assert.Contains(t, string(received), tt.wantBody, msgRespNotExpected)
				}
			}
		})
	}
}

func TestGetJsonFromUrl(t *testing.T) {

	const authToken = "test-token"
	expectedBody := `{"key": "value"}`
	// Create a mock server
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", authToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(expectedBody))
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
			return
		}
	}))
	defer mockServer.Close()

	// Create a CA cert pool with the server's certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(mockServer.Certificate())

	// Create a mock server that will be closed to simulate a connection error
	mockServerClosed := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just an empty mock server that will be closed
	}))
	// Create a CA cert pool with the server's certificate
	caCertPoolClosed := x509.NewCertPool()
	caCertPoolClosed.AddCert(mockServerClosed.Certificate())
	mockServerClosed.Close()

	// Create a mock server ReadError
	mockServerReadError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a partial write followed by a read error
		w.Header().Set("Content-Length", "1024")
		_, err := w.Write([]byte(expectedBody))
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
			return
		}

		// Close the connection prematurely to cause a read error
		conn, _, _ := w.(http.Hijacker).Hijack()
		err = conn.Close()
		if err != nil {
			t.Fatalf("Error closing connection: %v", err)
			return
		}
		// Close the connection immediately to cause a read error
		//w.(http.Flusher).Flush()
		//w.(http.CloseNotifier).CloseNotify()
	}))
	defer mockServerReadError.Close()

	type args struct {
		url           string
		bearerToken   string
		caCert        []byte
		allowInsecure bool
		logger        golog.MyLogger
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should return an error when the url is not reachable",
			args: args{
				url:           "http://remotehostthatwillnotexist:9999",
				bearerToken:   authToken,
				caCert:        nil,
				allowInsecure: false,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return status 200 ok when the url is reachable",
			args: args{
				url:           mockServer.URL,
				bearerToken:   authToken,
				caCert:        mockServer.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    expectedBody,
			wantErr: assert.NoError,
		},
		{
			name: "should return an error when the url is reachable but the token is invalid",
			args: args{
				url:           mockServer.URL,
				bearerToken:   "",
				caCert:        mockServer.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return an error when the url is reachable but the connection is refused",
			args: args{
				url:           mockServerClosed.URL,
				bearerToken:   authToken,
				caCert:        mockServerClosed.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return an error when the url is reachable but response cannot be read",
			args: args{
				url:           mockServerReadError.URL,
				bearerToken:   authToken,
				caCert:        nil,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := info.GetJsonFromUrl(tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.allowInsecure, 10*time.Second, tt.args.logger)
			if !tt.wantErr(t, err, fmt.Sprintf("GetJsonFromUrl(%v, %v, %v, %v)", tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.logger)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetJsonFromUrl(%v, %v, %v, %v)", tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.logger)
		})
	}
}

func setPortEnv(t *testing.T, port int) {
	err := os.Setenv("PORT", fmt.Sprintf("%d", port))
	if err != nil {
		t.Errorf("💥💥 ERROR: Unable to set env variable PORT")
		t.FailNow()
	}
}

func startMainServer(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
}

func newRequest(t *testing.T, method, url, body string, useFormUrlencodedContentType bool) *http.Request {
	fmt.Printf("INFO: 🚀🚀'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
	r, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf(fmtErrNewRequest, method, url, err)
	}
	if method == http.MethodPost && useFormUrlencodedContentType {
		r.Header.Set(gohttp.HeaderContentType, "Application/x-www-form-urlencoded")
	} else {
		r.Header.Set(gohttp.HeaderContentType, gohttp.MIMEAppJSON)
	}
	return r
}

func executeTest(t *testing.T, tt TestMainStruct, listenAddr string, l golog.MyLogger) {
	t.Run(tt.name, func(t *testing.T) {
		// prepare the request for this test case
		r := newRequest(t, tt.httpMethod, listenAddr+tt.url, tt.body, tt.useFormUrlencodedContentType)
		if len(tt.paramKeyValues) > 0 {
			parameters := r.URL.Query()
			for paramName, paramValue := range tt.paramKeyValues {
				parameters.Add(paramName, paramValue)
			}
			r.URL.RawQuery = parameters.Encode()
		}
		l.Debug(fmtTraceInfo, tt.name, tt.httpMethod, tt.url)
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			fmt.Printf(fmtErr, err, resp.Body)
			t.Fatal(err)
		}
		defer tools.CloseBody(resp.Body, tt.name, l)
		assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
		receivedJson, _ := io.ReadAll(resp.Body)
		rInfo := &info.RuntimeInfo{}
		tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
		if tt.wantStatusCode == http.StatusOK {
			if tt.contentType == gohttp.MIMEAppJSON {
				err = json.Unmarshal(receivedJson, rInfo)
				assert.Nil(t, err, "the output should be a valid json")
			}
		}
		// check that receivedJson contains the specified tt.wantBody substring
		assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
	})
}

func TestMainExecution(t *testing.T) {
	defaultPort := 9898
	listenAddr := fmt.Sprintf("%s://%s:%d", "http", "127.0.0.1", defaultPort)
	setPortEnv(t, defaultPort)
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)
	// starting main in his own go routine
	var wg sync.WaitGroup
	startMainServer(&wg)
	gohttp.WaitForHttpServer(listenAddr, 1*time.Second, 10)

	tests := []TestMainStruct{
		{
			name:                         "Get on info get handler should contain the Appname field",
			wantStatusCode:               http.StatusOK,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     fmt.Sprintf("\"appname\":\"%s\"", APP),
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          "/info",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "Post on default get handler should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     "Method Not Allowed",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodPost,
			url:                          "/",
			useFormUrlencodedContentType: true,
			body:                         `{"junk":"test with junk text"}`,
		},
		{
			name:                         "Get on nonexistent route should return an http error not found ",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     "page not found",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          "/aroutethatwillneverexisthere",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "Get on info Path should return a valid json containing param value",
			wantStatusCode:               http.StatusOK,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     `"param_name":"╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"`,
			paramKeyValues:               map[string]string{"name": "╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"},
			httpMethod:                   http.MethodGet,
			url:                          "/info",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "/health Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodPost,
			url:                          "/health",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/readiness Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodPost,
			url:                          "/readiness",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/time Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  gohttp.MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodPost,
			url:                          "/time",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/hello Get should return a welcome message",
			wantStatusCode:               http.StatusOK,
			contentType:                  gohttp.MIMEHtml,
			wantBody:                     "Hello World!",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          "/hello",
			useFormUrlencodedContentType: false,
			body:                         ``,
		},
	}

	for _, tt := range tests {
		executeTest(t, tt, listenAddr, l)
	}
}

func init() {
	var err error
	if DEBUG {
		l, err = golog.NewLogger("zap", golog.DebugLevel, fmt.Sprintf("testing_%s ", APP))
		if err != nil {
			log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
		}
	} else {
		l, err = golog.NewLogger("zap", golog.ErrorLevel, fmt.Sprintf("testing_%s ", APP))
		if err != nil {
			log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
		}
	}
}
