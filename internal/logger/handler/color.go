// Package handler provides a custom slog.Handler with ANSI color support.
package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// ColorHandler is an slog.Handler that writes colored, human-readable output.
type ColorHandler struct {
	mu      sync.Mutex
	out     io.Writer
	level   *slog.LevelVar
	noColor bool
	attrs   []slog.Attr
	groups  []string
}

// NewColorHandler creates a new ColorHandler.
func NewColorHandler(out io.Writer, level *slog.LevelVar, noColor bool) *ColorHandler {
	return &ColorHandler{
		out:     out,
		level:   level,
		noColor: noColor,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle formats and writes the log record.
func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// Level badge
	levelStr, levelColor := levelLabel(r.Level)
	if h.noColor {
		fmt.Fprintf(&buf, "%-5s ", levelStr)
	} else {
		fmt.Fprintf(&buf, "%s%s%-5s%s ", colorBold, levelColor, levelStr, colorReset)
	}

	// Message
	if h.noColor {
		buf.WriteString(r.Message)
	} else {
		fmt.Fprintf(&buf, "%s%s%s", colorBold, r.Message, colorReset)
	}

	// Persistent attrs set via WithAttrs.
	for _, a := range h.attrs {
		writeAttr(&buf, a, h.noColor)
	}

	// Record attrs.
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(&buf, a, h.noColor)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf.Bytes())
	return err
}

// WithAttrs returns a new handler with the given attributes always included.
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColorHandler{
		out:     h.out,
		level:   h.level,
		noColor: h.noColor,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *ColorHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, len(h.groups)+1)
	copy(groups, h.groups)
	groups[len(h.groups)] = name
	return &ColorHandler{
		out:     h.out,
		level:   h.level,
		noColor: h.noColor,
		attrs:   h.attrs,
		groups:  groups,
	}
}

// levelLabel returns the display label and color for a log level.
func levelLabel(l slog.Level) (string, string) {
	switch {
	case l >= slog.LevelError:
		return "ERROR", colorRed
	case l >= slog.LevelWarn:
		return "WARN", colorYellow
	case l >= slog.LevelInfo:
		return "INFO", colorBlue
	default:
		return "DEBUG", colorGray
	}
}

// writeAttr formats a single attribute into the buffer.
func writeAttr(buf *bytes.Buffer, a slog.Attr, noColor bool) {
	if a.Equal(slog.Attr{}) {
		return
	}
	key := a.Key
	val := strings.TrimSpace(a.Value.String())

	if noColor {
		fmt.Fprintf(buf, " %s=%s", key, val)
	} else {
		fmt.Fprintf(buf, " %s%s%s=%s%s%s", colorCyan, key, colorReset, colorGray, val, colorReset)
	}
}
