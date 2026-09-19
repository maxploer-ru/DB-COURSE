package middleware

import (
	"ZVideo/internal/domain"
	"io"
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

			logger := baseLogger
			ctx := domain.WithRequestID(r.Context(), requestID)
			ctx = domain.WithLogger(ctx, logger)
			r = r.WithContext(ctx)

			logger.DebugContext(ctx, "HTTP request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.Int64("content_length", r.ContentLength),
			)

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			attrs := []any{
				slog.Int("status", wrapped.statusCode),
				slog.Duration("duration", duration),
				slog.Int("bytes_written", wrapped.bytesWritten),
			}
			switch {
			case wrapped.statusCode >= http.StatusInternalServerError:
				logger.ErrorContext(ctx, "HTTP request completed", attrs...)
			case wrapped.statusCode >= http.StatusBadRequest:
				logger.WarnContext(ctx, "HTTP request completed", attrs...)
			default:
				logger.InfoContext(ctx, "HTTP request completed", attrs...)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	wroteHeader  bool
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(body)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) Flush() {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *responseWriter) ReadFrom(src io.Reader) (int64, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	if readerFrom, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		n, err := readerFrom.ReadFrom(src)
		rw.bytesWritten += int(n)
		return n, err
	}
	n, err := io.Copy(rw.ResponseWriter, src)
	rw.bytesWritten += int(n)
	return n, err
}
