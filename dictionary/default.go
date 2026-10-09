package dictionary

import (
	"io/fs"
	"sync/atomic"

	"github.com/enim-net/enim-core/dictionary/locales"
)

// Core returns enim-core's embedded locale files (load with dir ".").
func Core() fs.FS { return locales.FS }

var std atomic.Pointer[Dictionary]

// The default dictionary starts with core messages loaded so responses are
// localized out of the box. Services add or override keys with Load.
func init() {
	d := New()
	d.MustLoad(locales.FS, ".")
	std.Store(d)
}

func Default() *Dictionary {
	return std.Load()
}

func SetDefault(d *Dictionary) {
	if d != nil {
		std.Store(d)
	}
}

func Load(fsys fs.FS, dir string) error                 { return Default().Load(fsys, dir) }
func MustLoad(fsys fs.FS, dir string)                   { Default().MustLoad(fsys, dir) }
func Has(locale Locale) bool                            { return Default().Has(locale) }
func T(locale Locale, key string, params Params) string { return Default().T(locale, key, params) }
func Match(acceptLanguage string) (Locale, bool)        { return Default().Match(acceptLanguage) }
