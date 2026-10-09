package sentry

import "testing"

func TestNewRequiresDSN(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("empty DSN accepted")
	}
}
