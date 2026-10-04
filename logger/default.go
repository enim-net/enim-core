package logger

import (
	"context"
	"sync/atomic"
)

var std atomic.Pointer[Logger]

func init() {
	std.Store(New(Options{Level: LevelInfo}, NewConsole(ConsoleOptions{})))
}

func SetDefault(l *Logger) {
	if l != nil {
		std.Store(l)
	}
}

func L() *Logger { return std.Load() }

func Debug(ctx context.Context, msg string, f ...Field) { L().log(ctx, LevelDebug, msg, f) }
func Info(ctx context.Context, msg string, f ...Field)  { L().log(ctx, LevelInfo, msg, f) }
func Warn(ctx context.Context, msg string, f ...Field)  { L().log(ctx, LevelWarn, msg, f) }
func Error(ctx context.Context, msg string, f ...Field) { L().log(ctx, LevelError, msg, f) }
