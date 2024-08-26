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
		n, err := fmt.Fprintf(w, getHtmlPage(title, description))
		if err != nil {
			l.Error("💥💥 ERROR: [%s]  was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, r.URL.Path, r.RemoteAddr, n)
			http.Error(w, "Internal server error. GetStaticPageHandler was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}

func GetTimeHandler(l golog.MyLogger) http.HandlerFunc {
	handlerName := "GetTimeHandler"
	l.Debug(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		now := time.Now()
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "{\"time\":\"%s\"}", now.Format(time.RFC3339))
		if err != nil {
			l.Error("Error doing fmt.Fprintf err: %s", err)
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
		login := r.FormValue("login")
		// password := r.FormValue("pass")
		passwordHash := r.FormValue("hashed")
		s.logger.Debug("login: %s , password: %s, hash: %s ", login, passwordHash)
		// maybe it was not a form but a fetch data post
		if len(strings.Trim(login, " ")) < 1 {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}

		if s.Authenticator.AuthenticateUser(login, passwordHash) {
			userInfo, err := s.Authenticator.GetUserInfoFromLogin(login)
			if err != nil {
				s.logger.Error("Error getting user info from login: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			s.logger.Info(fmt.Sprintf("LoginUser(%s) succesfull login for User id (%d)", userInfo.UserLogin, userInfo.UserId))
			token, err := s.JwtCheck.GetTokenFromUserInfo(userInfo)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Prepare the response
			response := map[string]string{
				"token": token.String(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	}
}
