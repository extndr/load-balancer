package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Logger retrieves the logger from the context
func Logger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return l
	}
	return zap.NewNop()
}

// Logging middleware adds logger to context and logs HTTP requests
func Logging(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			reqLogger := baseLogger.With(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remoteAddr", r.RemoteAddr),
			)

			ctx := WithLogger(r.Context(), reqLogger)

			start := time.Now()
			next.ServeHTTP(rw, r.WithContext(ctx))
			duration := time.Since(start)

			backend := r.Header.Get("X-Backend")

			reqLogger.Info("request completed",
				zap.String("backend", backend),
				zap.Int("status", rw.statusCode),
				zap.Int64("duration_ms", duration.Milliseconds()),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
