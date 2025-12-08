package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			backend := r.Header.Get("X-Backend")
			fields := log.Fields{
				"method":     r.Method,
				"backend":    backend,
				"path":       r.URL.Path,
				"status":     rw.statusCode,
				"durationMs": duration.Milliseconds(),
			}
			log.WithFields(fields).Info("request")
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
