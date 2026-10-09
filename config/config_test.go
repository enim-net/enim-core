package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type nested struct {
	App App
	DB  DB `prefix:""`
	JWT JWT
	Ext struct {
		Name    string        `env:"NAME" default:"ext"`
		Timeout time.Duration `env:"TIMEOUT" default:"2s"`
		Tags    []string      `env:"TAGS"`
		Ratio   float64       `env:"RATIO" default:"0.5"`
		On      bool          `env:"ON" default:"true"`
		Count   uint          `env:"COUNT" default:"3"`
	} `prefix:"EXT_"`
}

// Regression for v.Field(1): each field must get its own value.
func TestLoadFillsEachField(t *testing.T) {
	type C struct {
		A string `env:"T_A" default:"a"`
		B string `env:"T_B" default:"b"`
		N int    `env:"T_N" default:"7"`
	}
	var c C
	if err := Load(&c, WithFiles()); err != nil {
		t.Fatal(err)
	}
	if c.A != "a" || c.B != "b" || c.N != 7 {
		t.Fatalf("got %+v", c)
	}
}

func TestLoadNestedEnvAndDefaults(t *testing.T) {
	t.Setenv("DB_NAME", "customer")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	t.Setenv("EXT_TAGS", "a, b,,c")
	t.Setenv("EXT_TIMEOUT", "150ms")
	var c nested
	if err := Load(&c, WithFiles()); err != nil {
		t.Fatal(err)
	}
	if c.App.Port != "8080" || c.App.Env != "development" || !c.App.IsDevelopment() {
		t.Errorf("App = %+v", c.App)
	}
	if c.DB.Name != "customer" || c.DB.Port != 6543 || c.DB.Host != "localhost" || c.DB.ConnMaxLifetime != time.Hour {
		t.Errorf("DB = %+v", c.DB)
	}
	if c.JWT.AccessTTL != 15*time.Minute || c.JWT.RefreshTTL != 720*time.Hour {
		t.Errorf("JWT = %+v", c.JWT)
	}
	e := c.Ext
	if e.Name != "ext" || e.Timeout != 150*time.Millisecond || strings.Join(e.Tags, "|") != "a|b|c" || e.Ratio != 0.5 || !e.On || e.Count != 3 {
		t.Errorf("Ext = %+v", e)
	}
}

func TestPresetValuesSurviveWhenEnvEmpty(t *testing.T) {
	var c nested
	c.DB.Name = "preset"
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	if err := Load(&c, WithFiles()); err != nil {
		t.Fatal(err)
	}
	if c.DB.Name != "preset" {
		t.Fatalf("DB.Name = %q", c.DB.Name)
	}
}

func TestLoadErrors(t *testing.T) {
	var c nested
	t.Setenv("DB_PORT", "not-a-number")
	t.Setenv("JWT_SECRET", "short")
	err := Load(&c, WithFiles())
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"DB_NAME is required", "DB_PORT", "JWT_SECRET must be at least 32"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q lacks %q", err, want)
		}
	}
	if err := Load(c); err == nil {
		t.Error("non-pointer accepted")
	}
	mustPanic := func() {
		defer func() {
			if recover() == nil {
				t.Error("MustLoad did not panic")
			}
		}()
		MustLoad(&c, WithFiles())
	}
	mustPanic()
}

func TestLoadDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.env")
	if err := os.WriteFile(f, []byte("FILE_VAL=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Unsetenv("FILE_VAL") })
	var c struct {
		V string `env:"FILE_VAL" required:"true"`
	}
	if err := Load(&c, WithFiles(f, filepath.Join(dir, "missing.env"))); err != nil {
		t.Fatal(err)
	}
	if c.V != "from-file" {
		t.Fatalf("V = %q", c.V)
	}
}

func TestPrintMasksSecrets(t *testing.T) {
	var c nested
	c.DB.Password = "hunter2"
	c.JWT.Secret = "topsecret"
	var buf bytes.Buffer
	Print(&buf, &c)
	out := buf.String()
	if strings.Contains(out, "hunter2") || strings.Contains(out, "topsecret") || !strings.Contains(out, "DB_PASSWORD=********") {
		t.Fatalf("Print output:\n%s", out)
	}
}

func TestDSNQuoting(t *testing.T) {
	d := DB{Host: "h", Port: 5432, User: "u", Password: "p w'd", Name: "n", SSLMode: "disable", TimeZone: "Asia/Jakarta"}
	want := `host=h port=5432 user=u password='p w\'d' dbname=n sslmode=disable TimeZone=Asia/Jakarta`
	if got := d.DSN(); got != want {
		t.Fatalf("DSN = %s", got)
	}
}

func TestJWTValidate(t *testing.T) {
	ok := JWT{Secret: strings.Repeat("x", 32), AccessTTL: time.Minute, RefreshTTL: time.Hour}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ok
	bad.RefreshTTL = time.Second
	if err := bad.Validate(); err == nil {
		t.Fatal("refresh <= access accepted")
	}
}
