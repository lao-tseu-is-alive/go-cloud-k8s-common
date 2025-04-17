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

func GetReadinessHandler(l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetReadinessHandler"
	l.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.WriteHeader(http.StatusOK)
	}
}

func GetHealthHandler(l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetHealthHandler"
	l.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.WriteHeader(http.StatusOK)
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
		err = s.JsonResponse(w, data)
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
