package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	DefaultFilePerms = 0o755
	RegularFilePerms = 0o644
)

type Options struct {
	LevelDebug bool
	LogFile    string
}

type Logger struct {
	logger *slog.Logger
	level  slog.Level
}

func NewLogger(opts Options) (*Logger, error) {
	var level slog.Level
	if opts.LevelDebug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	var handlers []slog.Handler

	// Use custom text handler instead of default text handler for neater output
	handlers = append(handlers, NewCustomTextHandler(os.Stderr, level))

	if opts.LogFile != "" {
		if err := os.MkdirAll(filepath.Dir(opts.LogFile), DefaultFilePerms); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(opts.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, RegularFilePerms)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level}))
	}

	handler := newMultiHandler(handlers...)

	return &Logger{
		logger: slog.New(handler),
		level:  level,
	}, nil
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
