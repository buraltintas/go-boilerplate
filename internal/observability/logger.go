package observability

import (
	"log/slog"
	"os"
)

// Logger is the application-wide structured logger.
// Single responsibility: provide structured logging interface.
var Logger *slog.Logger

func init() {
	// JSON structured logging for production
	// Can be made configurable via env vars if needed
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Convenience wrappers for common logging patterns

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// WithContext returns a logger with request context fields.
// This allows correlation across request lifecycle.
func WithContext(requestID, method, path string) *slog.Logger {
	return Logger.With(
		"request_id", requestID,
		"method", method,
		"path", path,
	)
}
