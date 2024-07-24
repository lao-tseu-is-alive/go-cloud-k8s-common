package go_http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func GetReadinessHandler(l *log.Logger) http.HandlerFunc {
	handlerName := "GetReadinessHandler"
	l.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.WriteHeader(http.StatusOK)
	}
}
func GetHealthHandler(l *log.Logger) http.HandlerFunc {
	handlerName := "GetHealthHandler"
	l.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.WriteHeader(http.StatusOK)
	}
}

func GetHandlerNotFound(l *log.Logger) http.HandlerFunc {
	handlerName := "GetHandlerNotFound"
	l.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		l.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
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
			l.Printf("💥💥 ERROR: [%s] Not Found was unable to Fprintf. path:'%s', from IP: [%s]\n", handlerName, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Internal server error. myDefaultHandler was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}

func GetHandlerStaticPage(title string, description string, l *log.Logger) http.HandlerFunc {
	handlerName := fmt.Sprintf("GetHandlerStaticPage[%s]", title)
	l.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		w.Header().Set(HeaderContentType, MIMEHtml)
		w.WriteHeader(http.StatusOK)
		n, err := fmt.Fprintf(w, getHtmlPage(title, description))
		if err != nil {
			l.Printf("💥💥 ERROR: [%s]  was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, r.URL.Path, r.RemoteAddr, n)
			http.Error(w, "Internal server error. GetHandlerStaticPage was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}
func GetTimeHandler(l *log.Logger) http.HandlerFunc {
	handlerName := "GetTimeHandler"
	l.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		now := time.Now()
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "{\"time\":\"%s\"}", now.Format(time.RFC3339))
		if err != nil {
			l.Printf("Error doing fmt.Fprintf err: %s", err)
		}
	}
}
func GetWaitHandler(secondsToSleep int, l *log.Logger) http.HandlerFunc {
	handlerName := "GetWaitHandler"
	l.Printf(initCallMsg, handlerName)
	durationOfSleep := time.Duration(secondsToSleep) * time.Second
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, l)
		if r.Method == http.MethodGet {
			w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			time.Sleep(durationOfSleep) // simulate a delay to be ready
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintf(w, "{\"waited\":\"%v seconds\"}", secondsToSleep)
			if err != nil {
				l.Printf("Error doing fmt.Fprintf err: %s", err)
			}
		} else {
			l.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, httpErrMethodNotAllow, http.StatusMethodNotAllowed)
		}
	}
}
