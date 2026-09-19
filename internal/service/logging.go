package service

import (
	"ZVideo/internal/domain"
	"context"
	"log/slog"
)

func serviceLogger(ctx context.Context, serviceName, operation string, attrs ...any) domain.Logger {
	base := []any{
		slog.String("service", serviceName),
		slog.String("operation", operation),
	}
	return domain.GetLogger(ctx).With(append(base, attrs...)...)
}
