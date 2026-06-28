// Package logger provides a structured, leveled logger for phpvm.
// It wraps the standard library slog package with:
//   - Colored human-readable output for terminal
//   - JSON output for file logging
//   - A global singleton accessible across the application
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/mahtdy/phpvm/internal/logger/handler"
)

// Level constants mirror slog levels for external callers.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

var (
	global *slog.Logger
	once   sync.Once
	mu     sync.RWMutex
)

// Options configures the global logger.
type Options struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level slog.Level
	// JSON outputs structured JSON instead of colored text.
	JSON bool
	// NoColor disables ANSI color codes in terminal output.
	NoColor bool
	// Output is the writer to write logs to. Defaults to os.Stderr.
	Output io.Writer
}

// Init initialises the global logger with the given options.
// It is idempotent when called multiple times with the same writer; call
// Reset() first to reconfigure.
func Init(opts Options) {
	mu.Lock()
	defer mu.Unlock()

	out := opts.Output
	if out == nil {
		out = os.Stderr
	}

	lvl := &slog.LevelVar{}
	lvl.Set(opts.Level)

	var h slog.Handler
	if opts.JSON {
		h = slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level:     lvl,
			AddSource: false,
		})
	} else {
		h = handler.NewColorHandler(out, lvl, opts.NoColor)
	}

	global = slog.New(h)
	slog.SetDefault(global)
}

// Reset clears the singleton so Init can be called again (useful in tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	global = nil
	once = sync.Once{}
}

// ensure the global logger is initialised with safe defaults.
func ensure() *slog.Logger {
	mu.RLock()
	l := global
	mu.RUnlock()
	if l != nil {
		return l
	}
	Init(Options{Level: LevelInfo})
	mu.RLock()
	defer mu.RUnlock()
	return global
}

// Debug logs a message at DEBUG level.
func Debug(msg string, args ...any) {
	ensure().Debug(msg, args...)
}

// Info logs a message at INFO level.
func Info(msg string, args ...any) {
	ensure().Info(msg, args...)
}

// Warn logs a message at WARN level.
func Warn(msg string, args ...any) {
	ensure().Warn(msg, args...)
}

// Error logs a message at ERROR level.
func Error(msg string, args ...any) {
	ensure().Error(msg, args...)
}

// DebugCtx logs with context.
func DebugCtx(ctx context.Context, msg string, args ...any) {
	ensure().DebugContext(ctx, msg, args...)
}

// InfoCtx logs with context.
func InfoCtx(ctx context.Context, msg string, args ...any) {
	ensure().InfoContext(ctx, msg, args...)
}

// WarnCtx logs with context.
func WarnCtx(ctx context.Context, msg string, args ...any) {
	ensure().WarnContext(ctx, msg, args...)
}

// ErrorCtx logs with context.
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	ensure().ErrorContext(ctx, msg, args...)
}

// With returns a child logger with the given attributes always attached.
func With(args ...any) *slog.Logger {
	return ensure().With(args...)
}

// Get returns the underlying *slog.Logger for advanced usage.
func Get() *slog.Logger {
	return ensure()
}
