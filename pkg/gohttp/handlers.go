package gohttp

import (
	"encoding/json"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/info"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/xid"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ReadinessOKMsg  = "(%s) is ready"
	ReadinessErrMsg = "(%s) is not ready"
	HealthOKMsg     = "(%s) is healthy"
	HealthErrMsg    = "(%s) is not healthy"
)

type FuncAreWeReady func(msg string) bool

type FuncAreWeHealthy func(msg string) bool

type StandardResponse struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	IsOk   bool        `json:"isOk"`
	Data   interface{} `json:"data,omitempty"`   // Optional: for including response data
	Errors []string    `json:"errors,omitempty"` // Optional: for detailed error messages
}

func GetStandardResponse(statusMsg, msg string, state bool, data interface{}, errors ...string) StandardResponse {
	return StandardResponse{
		Status: statusMsg,
		Msg:    msg,
		IsOk:   state,
		Data:   data,
		Errors: errors,
	}
}

func (s *Server) SendJSONResponse(w http.ResponseWriter, statusCode int, status, msg string, isOk bool, data interface{}, errors ...string) {
	response := GetStandardResponse(status, msg, isOk, data, errors...)
	if statusCode >= 400 {
		s.logger.Warn("http status code : %d, %s", statusCode, msg)
	}
	err := s.Json(w, response, statusCode, "")
	if err != nil {
		s.logger.Warn("error while sending json http status code : %d, %s", statusCode, msg)
	}
}

func (s *Server) GetReadinessHandler(readyFunc FuncAreWeReady, msg string) http.HandlerFunc {
	handlerName := "GetReadinessHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		ready := readyFunc(msg)
		if ready {
			msgOK := fmt.Sprintf(ReadinessOKMsg, msg)
			s.SendJSONResponse(w, http.StatusOK, "ready", msgOK, ready, nil)
		} else {
			msgErr := fmt.Sprintf(ReadinessErrMsg, msg)
			s.SendJSONResponse(w, http.StatusServiceUnavailable, "error", msgErr, ready, nil)
		}
	}
}

func (s *Server) GetHealthHandler(healthyFunc FuncAreWeHealthy, msg string) http.HandlerFunc {
	handlerName := "GetHealthHandler"
	s.logger.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		healthy := healthyFunc(msg)
		if healthy {
			msgOK := fmt.Sprintf(HealthOKMsg, msg)
			s.SendJSONResponse(w, http.StatusOK, "healthy", msgOK, healthy, nil)
		} else {
			msgErr := fmt.Sprintf(HealthErrMsg, msg)
			s.SendJSONResponse(w, http.StatusServiceUnavailable, "error", msgErr, healthy, nil)
		}
	}
}

func GetHandlerNotFound(l golog.MyLogger, rootPathNotFoundCounter prometheus.Counter) http.HandlerFunc {
	handlerName := "GetHandlerNotFound"
	l.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.WriteHeader(http.StatusNotFound)
		rootPathNotFoundCounter.Inc()
		type NotFound struct {
			Status  int    `json:"status"`
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		data := &NotFound{
			Status:  http.StatusNotFound,
			Error:   defaultNotFound,
			Message: defaultNotFoundDescription,
		}
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			l.Error("💥💥 ERROR: [%s] Not Found was unable to Fprintf. path:'%s', from IP: [%s]\n", handlerName, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Internal server error. myDefaultHandler was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}

func GetStaticPageHandler(title string, description string, l golog.MyLogger) http.HandlerFunc {
	handlerName := fmt.Sprintf("GetStaticPageHandler[%s]", title)
	l.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.Header().Set(HeaderContentType, MIMEHtml)
		w.WriteHeader(http.StatusOK)
		n, err := fmt.Fprintf(w, "%s", getHtmlPage(title, description))
		if err != nil {
			l.Error("💥💥 ERROR: [%s]  was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, r.URL.Path, r.RemoteAddr, n)
			http.Error(w, "Internal server error. GetStaticPageHandler was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}

func GetTimeHandler(l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetTimeHandler"
	l.Debug(initCallMsg, handlerName)

	type TimeResponse struct {
		Time string `json:"time"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)

		// Create response with current time
		response := TimeResponse{
			Time: time.Now().Format(time.RFC3339),
		}

		// Set response headers
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)

		// Encode and send response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			l.Error("Error encoding time response: %v", err)
			// Can't write an error response at this point as headers are already sent
		}
	}
}

func GetInfoHandler(s *Server) http.HandlerFunc {
	handlerName := "GetInfoHandler"
	logger := s.GetLog()
	logger.Debug("Initial call to %s", handlerName)
	v := s.VersionReader.GetVersionInfo()
	appName := v.App
	ver := v.Version
	data := info.CollectRuntimeInfo(appName, ver, logger)

	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, logger)
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
			logger.Error("GetOsUptime() returned an error : %+#v", err)
		}
		data.UptimeOs = uptimeOS
		guid := xid.New()
		data.RequestId = guid.String()
		err = s.Json(w, data, http.StatusOK, "")
		if err != nil {
			logger.Error("ERROR:  %v doing JsonResponse in %s, from IP: [%s]\n", err, handlerName, r.RemoteAddr)
			return
		}
		logger.Info("SUCCESS: [%s] from IP: [%s]\n", handlerName, r.RemoteAddr)
	}
}

func GetLoginPostHandler(s *Server) http.HandlerFunc {
	handlerName := "GetLoginPostHandler"
	logger := s.GetLog()
	logger.Debug("Initial call to %s", handlerName)

	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, logger)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get login and password hash from the form
		login := strings.TrimSpace(r.FormValue("login"))
		passwordHash := r.FormValue("hashed")

		// Validate login
		if login == "" {
			logger.Warn("Login attempt with empty username from IP: %s", r.RemoteAddr)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Log login attempt without exposing sensitive data
		logger.Debug("Login attempt for user: %s from IP: %s", login, r.RemoteAddr)

		// Authenticate user
		if s.Authenticator.AuthenticateUser(login, passwordHash) {
			userInfo, err := s.Authenticator.GetUserInfoFromLogin(login)
			if err != nil {
				logger.Error("Error getting user info for login '%s': %v", login, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Generate JWT token
			token, err := s.JwtCheck.GetTokenFromUserInfo(userInfo)
			if err != nil {
				logger.Error("Error generating token for user '%s': %v", login, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Log successful login
			logger.Info("Successful login for user '%s' (ID: %d)", userInfo.UserLogin, userInfo.UserId)

			// Prepare and send response
			response := map[string]string{
				"token": token.String(),
			}
			w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Error("Error encoding JSON response: %v", err)
				// Can't write error response at this point as headers are already sent
			}
		} else {
			// Log failed login attempt
			logger.Warn("Failed login attempt for user '%s' from IP: %s", login, r.RemoteAddr)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	}
}

func GetAppInfoHandler(s *Server) http.HandlerFunc {
	handlerName := "GetAppInfoHandler"
	s.logger.Debug(initCallMsg, handlerName)

	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, s.logger)
		appInfo := s.VersionReader.GetVersionInfo()
		err := s.Json(w, appInfo, http.StatusOK, "")
		if err != nil {
			s.logger.Error("Error doing JsonResponse'%s': %v", handlerName, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
