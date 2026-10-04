package dictionary

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
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

func languageFromFile(name string) Locale {
	stem := strings.TrimSuffix(name, path.Ext(name))
	if i := strings.IndexAny(stem, "_"); i > 0 {
		stem = stem[:i]
	}
	return Locale(strings.ToLower(stem))
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
