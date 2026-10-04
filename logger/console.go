package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

type ConsoleFormat string

const (
	FormatJSON ConsoleFormat = "json"
	FormatText ConsoleFormat = "text"
)

type ConsoleOptions struct {
	Format ConsoleFormat
	Writer io.Writer
}

func NewConsole(opts ConsoleOptions) Adapter {
	w := opts.Writer
	if w == nil {
		w = os.Stdout
	}
	
	ho := &slog.HandlerOptions{Level: slog.LevelDebug}
	var h slog.Handler
	
	if opts.Format == FormatJSON {
		h = slog.NewJSONHandler(w, ho)
	} else {
		h = slog.NewTextHandler(w, ho)
	}
	return &consoleAdapter{h: h}
}

type consoleAdapter struct {
	mu sync.Mutex
	h  slog.Handler
}

func (c *consoleAdapter) Write(ctx context.Context, e Entry) error {
	r := slog.NewRecord(e.Time, toSlogLevel(e.Level), e.Message, 0)
	for _, f := range e.Fields {
		r.AddAttrs(slog.Any(f.Key, f.Value))
	}
	if e.Error != nil {
		r.AddAttrs(slog.String(ErrorKey, e.Error.Error()))
	}
	
	if e.Caller != "" {
		r.AddAttrs(slog.String(ErrorKey, e.Error.Error()))
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.h.Handle(ctx, r)
}

func (c *consoleAdapter) Close(context.Context) error { return nil }

func toSlogLevel(l Level) slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
