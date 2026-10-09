// Package locales embeds enim-core's message files (one JSON file per
// locale, named after the locale: en_US.json, id_ID.json).
package locales

import "embed"

// FS holds the core locale files at its root.
//
//go:embed *.json
var FS embed.FS
