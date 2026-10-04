package logger

import "context"

type Adapter interface {
	Write(ctx context.Context, e Entry) error
	Close(ctx context.Context) error
}

func MinLevel(level Level, a Adapter) Adapter {
	return &levelFilter{min: level, next: a}
}

type levelFilter struct {
	min  Level
	next Adapter
}

func (f *levelFilter) Write(ctx context.Context, e Entry) error {
	if e.Level < f.min {
		return nil
	}
	return f.next.Write(ctx, e)
}

func (f *levelFilter) Close(ctx context.Context) error { return f.next.Close(ctx) }

type AdapterFunc func(ctx context.Context, e Entry) error

func (fn AdapterFunc) Write(ctx context.Context, e Entry) error { return fn(ctx, e) }
func (fn AdapterFunc) Close(context.Context) error              { return nil }
