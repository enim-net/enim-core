package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
)

type capture struct {
	mu      sync.Mutex
	entries []Entry
	closed  int
}

func (c *capture) Write(_ context.Context, e Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
	return nil
}
func (c *capture) Close(context.Context) error { c.closed++; return nil }

// Regression: Logger methods were empty and log() never called adapters.
func TestLoggerWritesToAdapters(t *testing.T) {
	rec := &capture{}
	l := New(Options{Level: LevelInfo, Service: "svc", Env: "test", AddCaller: true}, rec)
	ctx := WithFields(context.Background(), String("request_id", "r1"))
	l.With(String("module", "m")).Info(ctx, "hello", Int("n", 1), Err(nil))
	l.Debug(ctx, "filtered")
	l.Error(ctx, "failed", Err(errors.New("boom")))

	if len(rec.entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(rec.entries))
	}
	e := rec.entries[0]
	var keys []string
	for _, f := range e.Fields {
		keys = append(keys, f.Key)
	}
	if strings.Join(keys, ",") != "service,env,module,request_id,n" {
		t.Errorf("field order = %v", keys)
	}
	if e.Message != "hello" || e.Level != LevelInfo || e.Error != nil {
		t.Errorf("entry = %+v", e)
	}
	if !strings.HasPrefix(e.Caller, "logger/logger_test.go:") {
		t.Errorf("caller = %q, want this test file", e.Caller)
	}
	if rec.entries[1].Error == nil || rec.entries[1].Error.Error() != "boom" {
		t.Errorf("error not lifted: %+v", rec.entries[1])
	}
}

func TestPackageLevelFunctionsUseDefault(t *testing.T) {
	prev := L()
	t.Cleanup(func() { SetDefault(prev) })
	rec := &capture{}
	SetDefault(New(Options{Level: LevelDebug, AddCaller: true}, rec))
	Debug(context.Background(), "d")
	Warn(context.Background(), "w")
	if len(rec.entries) != 2 || !strings.HasPrefix(rec.entries[0].Caller, "logger/logger_test.go:") {
		t.Fatalf("%+v", rec.entries)
	}
	SetDefault(nil) // ignored
	if L() == nil {
		t.Fatal("SetDefault(nil) cleared the logger")
	}
}

func TestAdapterErrorsAndClose(t *testing.T) {
	var reported []error
	failing := AdapterFunc(func(context.Context, Entry) error { return errors.New("down") })
	rec := &capture{}
	l := New(Options{OnAdapterError: func(err error) { reported = append(reported, err) }}, failing, MinLevel(LevelWarn, rec))
	l.Info(context.Background(), "x")
	l.Warn(context.Background(), "y")
	if len(reported) != 2 || len(rec.entries) != 1 {
		t.Fatalf("reported=%v entries=%d", reported, len(rec.entries))
	}
	_ = l.Close(context.Background())
	_ = l.Close(context.Background())
	if rec.closed != 1 {
		t.Fatalf("closed %d times", rec.closed)
	}
	Nop().Error(context.Background(), "nothing") // must not panic
	Nop().Info(nil, "nil ctx")                   //nolint:staticcheck // nil ctx tolerated
}

// Regression: console adapter dereferenced a nil error when a caller was set.
func TestConsoleFormats(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{Level: LevelDebug, AddCaller: true}, NewConsole(ConsoleOptions{Format: FormatJSON, Writer: &buf}))
	l.Info(context.Background(), "json line", String("k", "v"))
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %q", buf.String())
	}
	if m["msg"] != "json line" || m["k"] != "v" || !strings.Contains(m["caller"].(string), "logger_test.go") {
		t.Fatalf("%v", m)
	}
	buf.Reset()
	l = New(Options{Level: LevelDebug}, NewConsole(ConsoleOptions{Writer: &buf}))
	l.Error(context.Background(), "text line", Err(errors.New("bad")))
	if out := buf.String(); !strings.Contains(out, "msg=\"text line\"") || !strings.Contains(out, "error=bad") {
		t.Fatalf("text output %q", out)
	}
}

func TestParseLevelAndFieldMap(t *testing.T) {
	for in, want := range map[string]Level{"debug": LevelDebug, " INFO ": LevelInfo, "warn": LevelWarn, "error": LevelError} {
		if got, err := ParseLevel(in); err != nil || got != want {
			t.Errorf("ParseLevel(%q) = %v, %v", in, got, err)
		}
	}
	if _, err := ParseLevel("loud"); err == nil {
		t.Error("invalid level accepted")
	}
	e := Entry{Fields: []Field{Any("e", errors.New("x")), Duration("d", 1500000000)}, Error: errors.New("y"), Caller: "a.go:1"}
	m := e.FieldMap()
	if m["e"] != "x" || m["d"] != "1.5s" || m[ErrorKey] != "y" || m["caller"] != "a.go:1" {
		t.Fatalf("%v", m)
	}
}
