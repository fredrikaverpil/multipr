package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type CustomTextHandler struct {
	out   io.Writer
	level slog.Level
	mu    sync.Mutex
}

func NewCustomTextHandler(w io.Writer, level slog.Level) *CustomTextHandler {
	return &CustomTextHandler{
		out:   w,
		level: level,
	}
}

func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format message based on level
	var msg string
	switch r.Level {
	case slog.LevelDebug:
		msg = fmt.Sprintf("[multipr:DEBUG] %s\n", r.Message)
	case slog.LevelWarn:
		msg = fmt.Sprintf("[multipr:WARN] %s\n", r.Message)
	case slog.LevelError:
		msg = fmt.Sprintf("[multipr:ERROR] %s\n", r.Message)
	default: // INFO or other levels
		msg = fmt.Sprintf("[multipr] %s\n", r.Message)
	}

	_, err := h.out.Write([]byte(msg))
	return err
}

func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Ignore attributes in the custom format
	return h
}

func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	// Ignore groups in the custom format
	return h
}
