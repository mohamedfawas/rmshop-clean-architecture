package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var requestCounter int64

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic occurred: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		requestID := atomic.AddInt64(&requestCounter, 1)
		start := time.Now()

		rw := &responseWriter{w, http.StatusOK, []byte{}}

		next.ServeHTTP(rw, r)

		log.Printf(
			"RequestID: %d | Method: %s | Path: %s | RemoteAddr: %s | UserAgent: %s | Status: %d | Latency: %v | ResponseBody: %s",
			requestID,
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			rw.status,
			time.Since(start),
			string(rw.body),
		)
	})
}
