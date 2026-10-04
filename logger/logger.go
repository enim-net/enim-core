package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Options struct {
	Level          Level
	Service        string
	Env            string
	Fields         []Field
	AddCaller      bool
	OnAdapterError func(err error)
}

type Logger struct {
	core   *core
	fields []Field
}

type core struct {
	level     Level
	adapters  []Adapter
	base      []Field
	addCaller bool
	onError   func(error)
	closeOnce sync.Once
	closeErr  error
}

func New(opts Options, adapters ...Adapter) *Logger {
	base := make([]Field, 0, len(opts.Fields)+2)
	
	if opts.Service != "" {
		base = append(base, String("service", opts.Service))
	}
	
	if opts.Env != "" {
		base = append(base, String("env", opts.Env))
	}
	base = append(base, opts.Fields...)
	
	onError := opts.OnAdapterError
	if onError == nil {
		onError = func(err error) { fmt.Fprintln(os.Stderr, "logger: ", err) }
	}
	
	return &Logger{core: &core{
		level:     opts.Level,
		adapters:  adapters,
		base:      base,
		addCaller: opts.AddCaller,
		onError:   onError,
	}}
}

func Nop() *Logger { return New(Options{}) }

func (l *Logger) With(fields ...Field) *Logger {
	merged := make([]Field, 0, len(l.fields)+len(fields))
	merged = append(merged, l.fields...)
	merged = append(merged, fields...)
	
	return &Logger{core: l.core, fields: merged}
}

// Enabled reports whether entries at level would be logged.
func (l *Logger) Enabled(level Level) bool                               { return level >= l.core.level }
func (l *Logger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (l *Logger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (l *Logger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (l *Logger) Error(ctx context.Context, msg string, fields ...Field) {}

func (l *Logger) Close(ctx context.Context) error {
	c := l.core
	c.closeOnce.Do(func() {
		var errs []error
		for _, a := range c.adapters {
			if err := a.Close(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		c.closeErr = errors.Join(errs...)
	})
	
	return c.closeErr
}

const callerSkip = 3

func (l *Logger) log(ctx context.Context, level Level, msg string, fields []Field) {
	c := l.core
	if level < c.level || len(c.adapters) == 0 {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	
	ctxFields := FieldsFromContext(ctx)
	all := make([]Field, 0, len(c.base)+len(l.fields)+len(ctxFields)+len(fields))
	all = append(all, c.base...)
	all = append(all, l.fields...)
	all = append(all, ctxFields...)
	all = append(all, fields...)
	
	e := Entry{Time: time.Now(), Level: level, Message: msg}
	e.Fields = make([]Field, 0, len(all))
	for _, f := range all {
		if f.Key == ErrorKey {
			err, isErr := f.Value.(error)
			if f.Value == nil || (isErr && err == nil) {
				continue
			}
			if isErr && e.Error == nil {
				e.Error = err
				continue
			}
		}
		e.Fields = append(e.Fields, f)
	}
	
	if c.addCaller {
		if _, file, line, ok := runtime.Caller(callerSkip - 1); ok {
			e.Caller = fmt.Sprintf("%s: %d", shortPath(file), line)
		}
	}
	
}

func shortPath(p string) string {
	dir, file := filepath.Split(p)
	return filepath.Join(filepath.Base(dir), file)
}
