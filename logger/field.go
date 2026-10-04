package logger

import "time"

type Field struct {
	Key   string
	Value any
}

const ErrorKey = "error"

func String(k, v string) Field                 { return Field{k, v} }
func Int(k string, v int) Field                { return Field{k, v} }
func Int64(k string, v int64) Field            { return Field{k, v} }
func Uint(k string, v uint) Field              { return Field{k, v} }
func Float64(k string, v float64) Field        { return Field{k, v} }
func Bool(k string, v bool) Field              { return Field{k, v} }
func Duration(k string, v time.Duration) Field { return Field{k, v} }
func Time(k string, v time.Time) Field         { return Field{k, v} }
func Any(k string, v interface{}) Field        { return Field{k, v} }

func Err(err error) Field { return Field{ErrorKey, err} }
