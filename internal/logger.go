package internal

import (
	"context"
	"io"
	"log/slog"
)

type SimpleSlogHandler struct {
	w *io.Writer
	level slog.Level
}

func (h *SimpleSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SimpleSlogHandler) Handle() slog.Handler {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func (_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.LevelKey {
				level := attr.Value.Any().(slog.Level)
				var colorTag string
				switch level {
				case slog.LevelDebug:
					colorTag = "gray"
				case slog.LevelInfo:
					colorTag = "green"
				case slog.LevelWarn:
					colorTag = "yellow"
				case slog.LevelError:
					colorTag = "red"
				default:
					colorTag = "white"
				}

				coloredLevel := fmt.Sprintf("[%s::b]%s[-]", colorTag, level.String())
				attr.Value = slog.StringValue(coloredLevel)
			}
			return attr
		},
	}

	return slog.NewTextHandler(view, opts)
}

// No-op, satisfy the interface
func (h *SimpleSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *SimpleSlogHandler) WithGroup(name string) slog.Handler       { return h }
