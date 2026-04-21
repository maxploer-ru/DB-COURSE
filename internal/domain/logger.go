package domain

import "context"

type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
}
type contextKey string

const (
	LoggerKey    contextKey = "logger"
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
)

func GetLogger(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerKey).(Logger); ok {
		return logger
	}
	return &nopLogger{}
}

func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithLoggerAttributes(ctx context.Context, attrs ...any) context.Context {
	logger := GetLogger(ctx).With(attrs...)
	return WithLogger(ctx, logger)
}
