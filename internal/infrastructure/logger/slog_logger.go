package logger

import (
	"ZVideo/internal/domain"
	"context"
	"io"
	"log/slog"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger(level slog.Level, output io.Writer, addSource bool) domain.Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	})
	return &SlogLogger{logger: slog.New(handler)}
}

func (l *SlogLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	var attrs []slog.Attr
	if rid, ok := ctx.Value(domain.RequestIDKey).(string); ok && rid != "" {
		attrs = append(attrs, slog.String("requestID", rid))
	}
	if uid, ok := ctx.Value(domain.UserIDKey).(int); ok && uid != 0 {
		attrs = append(attrs, slog.Int("userID", uid))
	}
	for _, arg := range args {
		if attr, ok := arg.(slog.Attr); ok {
			attrs = append(attrs, attr)
		} else {
		}
	}
	l.logger.LogAttrs(ctx, level, msg, attrs...)
}

func argsToAttrs(args []any) []slog.Attr {
	var attrs []slog.Attr
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
	}
	return attrs
}

func (l *SlogLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}
func (l *SlogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}
func (l *SlogLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}
func (l *SlogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

func (l *SlogLogger) With(args ...any) domain.Logger {
	return &SlogLogger{logger: l.logger.With(args...)}
}
func (l *SlogLogger) WithGroup(name string) domain.Logger {
	return &SlogLogger{logger: l.logger.WithGroup(name)}
}
