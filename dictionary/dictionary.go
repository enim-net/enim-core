package dictionary

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type Locale string

const (
	LocaleID Locale = "id_ID"
	LocaleEN Locale = "en_US"
	LocaleDE Locale = "de_DE"
	LocaleFR Locale = "fr_FR"
	LocaleJA Locale = "ja_JP"

	DefaultLocale = LocaleEN
)

type Params map[string]any

var placeholders = regexp.MustCompile(`{{\s*([\w.]+)\s*}}`)

type Dictionary struct {
	mu       sync.RWMutex
	fallback Locale
	localeOf func(fileName string) Locale
	messages map[Locale]map[string]string
}

type Option func(*Dictionary)

func WithFallback(l Locale) Option {
	return func(d *Dictionary) { d.fallback = l }
}

func WithLocaleFromFile(fn func(fileName string) Locale) Option {
	return func(d *Dictionary) { d.localeOf = fn }
}

func New(opts ...Option) *Dictionary {
	d := &Dictionary{
		fallback: DefaultLocale,
		localeOf: languageFromFile,
		messages: map[Locale]map[string]string{},
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

func (d *Dictionary) Load(fsys fs.FS, dir string) error {
	files, err := fs.Glob(fsys, path.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("dictionary: glob %s: %w", dir, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("dictionary: no *.json files in %q", dir)
	}
	for _, file := range files {
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("dictionary: read %s: %w", file, err)
		}
		var tree map[string]any
		if err := json.Unmarshal(raw, &tree); err != nil {
			return fmt.Errorf("dictionary: parse %s: %w", file, err)
		}

		table := map[string]string{}
		if err := flatten("", tree, table); err != nil {
			return fmt.Errorf("dictionary: %s: %w", file, err)
		}
		d.Add(d.localeOf(path.Base(file)), table)
	}
	return nil
}

func (d *Dictionary) MustLoad(fsys fs.FS, dir string) {
	if err := d.Load(fsys, dir); err != nil {
		panic(err)
	}
}

func (d *Dictionary) Add(locale Locale, msgs map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	table, ok := d.messages[locale]
	if !ok {
		table = make(map[string]string, len(msgs))
		d.messages[locale] = table
	}
	for k, v := range msgs {
		table[k] = v
	}
}

func (d *Dictionary) Has(locale Locale) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, ok := d.messages[locale]
	return ok
}

func (d *Dictionary) Locales() []Locale {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]Locale, 0, len(d.messages))
	for l := range d.messages {
		out = append(out, l)
	}
	return out
}

func (d *Dictionary) T(locale Locale, key string, params Params) string {
	d.mu.RLock()
	text, ok := d.messages[locale][key]
	if !ok {
		text, ok = d.messages[d.fallback][key]
	}
	d.mu.RUnlock()
	if !ok {
		return key
	}

	if len(params) == 0 {
		return text
	}

	return placeholders.ReplaceAllStringFunc(text, func(token string) string {
		name := placeholders.FindStringSubmatch(token)[1]
		if v, ok := params[name]; ok {
			return fmt.Sprint(v)
		}

		return token
	})
}

// languageFromFile maps a file name to its locale: "id_ID.json" -> "id_ID".
func languageFromFile(name string) Locale {
	return Locale(strings.TrimSuffix(name, path.Ext(name)))
}

// Match picks the best loaded locale for an Accept-Language header such as
// "id-ID,id;q=0.9,en;q=0.8". Tags are tried in order (q-values are assumed
// to be descending, as browsers send them); "id-ID" matches "id_ID", and a
// bare language ("id") matches the first loaded locale with that language.
func (d *Dictionary) Match(acceptLanguage string) (Locale, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for part := range strings.SplitSeq(acceptLanguage, ",") {
		tag, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		tag = strings.ReplaceAll(strings.TrimSpace(tag), "-", "_")
		if tag == "" || tag == "*" {
			continue
		}
		for l := range d.messages {
			if strings.EqualFold(string(l), tag) {
				return l, true
			}
		}
		lang, _, _ := strings.Cut(tag, "_")
		var candidates []Locale
		for l := range d.messages {
			if ll, _, _ := strings.Cut(string(l), "_"); strings.EqualFold(ll, lang) {
				candidates = append(candidates, l)
			}
		}
		if len(candidates) > 0 {
			slices.Sort(candidates) // deterministic
			return candidates[0], true
		}
	}
	return "", false
}

func flatten(prefix string, in map[string]any, out map[string]string) error {
	for k, v := range in {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch t := v.(type) {
		case string:
			out[key] = t
		case map[string]any:
			if err := flatten(key, t, out); err != nil {
				return err
			}
		default:
			return fmt.Errorf("key %q: value must be a string or object, got %T", key, v)
		}
	}

	return nil
}
