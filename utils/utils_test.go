package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("U_S", "v")
	t.Setenv("U_I", "42")
	t.Setenv("U_B", "true")
	t.Setenv("U_D", "90s")
	t.Setenv("U_DS", "30")
	t.Setenv("U_BAD", "x")
	if GetEnv("U_S", "d") != "v" || GetEnv("U_MISSING", "d") != "d" {
		t.Error("GetEnv")
	}
	if GetEnvInt("U_I", 0) != 42 || GetEnvInt("U_BAD", 7) != 7 {
		t.Error("GetEnvInt")
	}
	if !GetEnvBool("U_B", false) || GetEnvBool("U_BAD", false) {
		t.Error("GetEnvBool")
	}
	if GetEnvDuration("U_D", 0) != 90*time.Second || GetEnvDuration("U_DS", 0) != 30*time.Second || GetEnvDuration("U_BAD", time.Second) != time.Second {
		t.Error("GetEnvDuration")
	}
}

func TestLoadDotEnvSkipsMissing(t *testing.T) {
	f := filepath.Join(t.TempDir(), "x.env")
	if err := os.WriteFile(f, []byte("U_FROM_FILE=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Unsetenv("U_FROM_FILE") })
	if err := LoadDotEnv("/nonexistent/.env", f); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("U_FROM_FILE") != "1" {
		t.Fatal("file not loaded")
	}
}
