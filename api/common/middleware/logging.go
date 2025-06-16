package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.WithFields(log.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     rw.statusCode,
			"duration":   duration,
			"user_agent": r.UserAgent(),
		}).Info("Request completed")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
