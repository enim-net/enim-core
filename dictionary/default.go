package dictionary

import (
	"embed"
	"io/fs"
	"sync/atomic"
)

var coreFS embed.FS

func Core() fs.FS { return coreFS }

var std atomic.Pointer[Dictionary]

func init() {
	std.Store(New())
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
