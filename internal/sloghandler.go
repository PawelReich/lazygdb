package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/rivo/tview"
)

type SlogHandler struct {
	w     io.Writer
	level slog.Level
}

func NewSimpleSlogHandler(w io.Writer, level slog.Level) *SlogHandler {
	return &SlogHandler{w: w, level: level}
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	var colorTag string

	switch r.Level {
	case slog.LevelDebug:
		colorTag = "grey"
	case slog.LevelInfo:
		colorTag = "silver"
	case slog.LevelWarn:
		colorTag = "yellow"
	case slog.LevelError:
		colorTag = "red"
	default:
		colorTag = "silver"
	}
	_, err := fmt.Fprintf(h.w, "[-][%s]%s\n", colorTag, tview.Escape(r.Message))

	return err
}

// No-op, satisfy the interface
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *SlogHandler) WithGroup(name string) slog.Handler       { return h }
