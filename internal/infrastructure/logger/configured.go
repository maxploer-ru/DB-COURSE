package logger

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/config"
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewConfigured creates the application logger and owns the lifecycle of its
// output. Call the returned close function during process shutdown.
func NewConfigured(cfg config.LoggingConfig) (domain.Logger, func()) {
	output, closeOutput := configureOutput(cfg.OutputPath)
	return NewSlogLogger(ParseLevel(cfg.Level), output, cfg.AddSource), closeOutput
}

func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func configureOutput(path string) (io.Writer, func()) {
	if path == "" || strings.EqualFold(path, "stdout") {
		return os.Stdout, func() {}
	}

	file := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	return file, func() { _ = file.Close() }
}
