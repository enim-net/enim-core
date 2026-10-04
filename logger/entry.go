package logger

import (
	"fmt"
	"time"
)

type Entry struct {
	Time    time.Time
	Level   Level
	Message string
	Fields  []Field
	Error   error
	Caller  string
}

func (e Entry) FieldMap() map[string]any {
	m := make(map[string]any, len(e.Fields)+2)
	for _, f := range e.Fields {
		m[f.Key] = normalize(f.Value)
	}
	
	if e.Error != nil {
		m[ErrorKey] = e.Error.Error()
	}
	if e.Caller != "" {
		m["caller"] = e.Caller
	}
	return m
}

func normalize(v any) any {
	switch t := v.(type) {
	case error:
		return t.Error()
	case time.Duration:
		return t.String()
	case fmt.Stringer:
		return t.String()
	default:
		return v
	}
}
