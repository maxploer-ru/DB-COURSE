package logger

import (
	"ZVideo/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlogLoggerPreservesStructuredAttributes(t *testing.T) {
	var output bytes.Buffer
	appLogger := NewSlogLogger(slog.LevelDebug, &output, false)
	ctx := domain.WithUserID(domain.WithRequestID(context.Background(), "request-123"), 42)

	appLogger.With("service", "VideoService", "operation", "GetVideo").InfoContext(ctx, "video loaded", slog.Int("video_id", 7))

	var record map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &record))
	require.Equal(t, "VideoService", record["service"])
	require.Equal(t, "GetVideo", record["operation"])
	require.Equal(t, "request-123", record["request_id"])
	require.Equal(t, float64(42), record["user_id"])
	require.Equal(t, float64(7), record["video_id"])
}
