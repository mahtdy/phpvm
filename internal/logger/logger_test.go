package logger_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitAndLevels(t *testing.T) {
	var buf bytes.Buffer
	logger.Reset()
	logger.Init(logger.Options{
		Level:   logger.LevelDebug,
		Output:  &buf,
		NoColor: true,
	})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	out := buf.String()
	assert.Contains(t, out, "debug message")
	assert.Contains(t, out, "info message")
	assert.Contains(t, out, "warn message")
	assert.Contains(t, out, "error message")
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger.Reset()
	logger.Init(logger.Options{
		Level:   logger.LevelWarn,
		Output:  &buf,
		NoColor: true,
	})

	logger.Debug("should not appear")
	logger.Info("should not appear either")
	logger.Warn("should appear")
	logger.Error("should also appear")

	out := buf.String()
	assert.NotContains(t, out, "should not appear")
	assert.Contains(t, out, "should appear")
	assert.Contains(t, out, "should also appear")
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger.Reset()
	logger.Init(logger.Options{
		Level:  logger.LevelInfo,
		JSON:   true,
		Output: &buf,
	})

	logger.Info("json test", "key", "value")

	out := buf.String()
	assert.True(t, strings.HasPrefix(strings.TrimSpace(out), "{"), "expected JSON output, got: %s", out)
	assert.Contains(t, out, `"msg":"json test"`)
	assert.Contains(t, out, `"key":"value"`)
}

func TestWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger.Reset()
	logger.Init(logger.Options{
		Level:   logger.LevelDebug,
		Output:  &buf,
		NoColor: true,
	})

	child := logger.With("component", "test")
	child.Info("with attrs")

	out := buf.String()
	assert.Contains(t, out, "component")
	assert.Contains(t, out, "test")
}

func TestGetReturnsLogger(t *testing.T) {
	logger.Reset()
	logger.Init(logger.Options{Level: logger.LevelInfo, Output: new(bytes.Buffer)})
	l := logger.Get()
	assert.NotNil(t, l)
	assert.IsType(t, &slog.Logger{}, l)
}
