package middleware

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var requestCounter int64

type responseWriter struct {
	http.ResponseWriter
	status int
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := atomic.AddInt64(&requestCounter, 1)
		// Add request ID to the request context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		start := time.Now()

		// Create a custom ResponseWriter to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log the request details
		log.Printf(
			"RequestID: %d | Method: %s | Path: %s | RemoteAddr: %s | UserAgent: %s | Status: %d | Latency: %v",
			requestID,
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			rw.status,
			time.Since(start),
		)
	})
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
