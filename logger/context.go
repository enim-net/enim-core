package logger

import "context"

type ctxKey struct{}

func WithFields(ctx context.Context, fields ...Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	existing := FieldsFromContext(ctx)
	merged := make([]Field, 0, len(existing)+len(fields))
	merged = append(merged, existing...)
	merged = append(merged, fields...)
	
	return context.WithValue(ctx, ctxKey{}, merged)
}

func FieldsFromContext(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	
	f, _ := ctx.Value(ctxKey{}).([]Field)
	return f
}
