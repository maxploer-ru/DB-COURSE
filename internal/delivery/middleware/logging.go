package middleware

import (
	"ZVideo/internal/domain"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func Logging(baseLogger domain.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()

			logger := baseLogger.With(slog.String("requestID", requestID))
			ctx := domain.WithLogger(r.Context(), logger)
			r = r.WithContext(ctx)

			logger.InfoContext(ctx, "HTTP request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.InfoContext(ctx, "HTTP request completed",
				slog.Int("status", wrapped.statusCode),
				slog.Duration("duration", duration),
			)
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
