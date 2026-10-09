package dictionary

import (
	"testing"
	"testing/fstest"
)

func TestCoreMessagesLoadedByDefault(t *testing.T) {
	// Regression: embed directives were missing, so T returned the key.
	if got := T(LocaleEN, "general.success", nil); got != "Operation completed successfully." {
		t.Fatalf("en general.success = %q", got)
	}
	if got := T(LocaleID, "general.success", nil); got != "Proses berhasil dilakukan." {
		t.Fatalf("id general.success = %q", got)
	}
	// Regression: id_ID.json was registered as locale "id".
	if !Has(LocaleID) || !Has(LocaleEN) {
		t.Fatalf("locales = %v", Default().Locales())
	}
}

func TestLocaleFilesHaveSameKeys(t *testing.T) {
	d := Default()
	en, id := d.messages[LocaleEN], d.messages[LocaleID]
	for k := range en {
		if _, ok := id[k]; !ok {
			t.Errorf("id_ID missing %s", k)
		}
	}
	for k := range id {
		if _, ok := en[k]; !ok {
			t.Errorf("en_US missing %s", k)
		}
	}
}

func TestTParamsFallbackAndMissing(t *testing.T) {
	d := New(WithFallback(LocaleEN))
	d.Add(LocaleEN, map[string]string{"hello": "Hello {{name}} from {{ place }}", "only.en": "en"})
	d.Add(LocaleID, map[string]string{"hello": "Halo {{name}}"})
	if got := d.T(LocaleEN, "hello", Params{"name": "Ana", "place": "Bandung"}); got != "Hello Ana from Bandung" {
		t.Errorf("params: %q", got)
	}
	if got := d.T(LocaleEN, "hello", Params{"name": "Ana"}); got != "Hello Ana from {{ place }}" {
		t.Errorf("missing param must stay visible: %q", got)
	}
	if got := d.T(LocaleID, "only.en", nil); got != "en" {
		t.Errorf("fallback: %q", got)
	}
	if got := d.T(LocaleID, "nope", nil); got != "nope" {
		t.Errorf("missing key: %q", got)
	}
}

func TestLoadNestedAndOverride(t *testing.T) {
	fsys := fstest.MapFS{
		"msg/en_US.json": {Data: []byte(`{"general":{"success":"Overridden"},"svc.x":"X"}`)},
		"bad/en_US.json": {Data: []byte(`{"n": 1}`)},
	}
	d := New()
	d.MustLoad(Core(), ".")
	if err := d.Load(fsys, "msg"); err != nil {
		t.Fatal(err)
	}
	if d.T(LocaleEN, "general.success", nil) != "Overridden" || d.T(LocaleEN, "svc.x", nil) != "X" {
		t.Fatal("service files must add and override keys")
	}
	if err := d.Load(fsys, "bad"); err == nil {
		t.Fatal("non-string value accepted")
	}
	if err := d.Load(fsys, "empty"); err == nil {
		t.Fatal("empty dir accepted")
	}
}

func TestMatchAcceptLanguage(t *testing.T) {
	d := New()
	d.MustLoad(Core(), ".")
	cases := map[string]Locale{
		"id_ID":                   LocaleID,
		"id-ID,id;q=0.9,en;q=0.8": LocaleID,
		"id":                      LocaleID,
		"en-GB,en;q=0.9":          LocaleEN,
		"fr-FR, en-US;q=0.5":      LocaleEN,
		"EN-us":                   LocaleEN,
	}
	for header, want := range cases {
		if got, ok := d.Match(header); !ok || got != want {
			t.Errorf("Match(%q) = %q, %v; want %q", header, got, ok, want)
		}
	}
	for _, header := range []string{"", "*", "ja-JP"} {
		if got, ok := d.Match(header); ok {
			t.Errorf("Match(%q) = %q, want no match", header, got)
		}
	}
}
