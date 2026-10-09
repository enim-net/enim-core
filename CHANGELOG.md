# Changelog

## [Unreleased] — planned v0.0.5

### Fixed
- **config**: `Load` wrote every env value/default into the second struct
  field (`v.Field(1)`); each field now gets its own value.
- **logger**: `Logger.Debug/Info/Warn/Error` were empty and `log()` never
  called adapters, so configured loggers produced no output. Caller is now
  reported as `file.go:line` at the real call site.
- **logger**: console adapter panicked (nil error dereference) when
  `AddCaller` was enabled; caller is written as the `caller` attribute.
- **errs**: `SetPrefix` regex contained a stray `%`, so every valid prefix
  (e.g. `ISP`) panicked.
- **errs**: `ErrInternal` (category general) mapped to HTTP 200; general now
  maps to 500. Success envelopes still use code `"00"`.
- **errs**: damaged names `ErrUnauthentiCategoryed` / `ErrDuplicateCategory`
  and keys `error.auth.unauthentiCategoryed` / `error.dupliCategorye`.
  New names: `ErrUnauthenticated`, `ErrDuplicate`, keys
  `error.auth.unauthenticated`, `error.duplicate`. Old names kept as
  deprecated aliases.
- **dictionary**: locale files were never embedded (missing `//go:embed`),
  so `T` returned keys. Core messages are now loaded into the default
  dictionary at init.
- **dictionary**: `id_ID.json` was registered as locale `id`; files now map
  to their full name (`id_ID`).
- **response**: `ErrorHandler` returned `err.Error()` to clients for unknown
  errors (could leak SQL/driver/host details) and ignored `*errs.Error`.
  It now renders `*errs.Error` (status, code, localized message, field
  errors), turns anything else into `ErrInternal`, and logs the cause.
- **response**: removed debug `fmt.Println` from `OK`.
- **httpx**: `NewPageMeta` divided by zero when `Limit` was 0.

### Added
- `dictionary.Match` / `(*Dictionary).Match`: pick a loaded locale from an
  `Accept-Language` header (`id-ID,id;q=0.9` → `id_ID`); used by
  `response.Locale`.
- `httpx.ValidateErr`: validation failures as a localized `*errs.Error` with
  per-field codes (required, email, min/max length, oneof, range...).
- `error.*` messages for every core code in `en_US` and `id_ID`.
- Tests for every package (previously none).

### Changed
- `httpx.BodyParser` errors are `*errs.Error` with `ErrMalformedBody`
  (HTTP 400) instead of plain errors; the decoder error is the cause.
- `en_US` strings fixed: `general.create.conflict` (was Indonesian),
  `general.create.failure`, `validate.format.uuid`.
- logger/sentry: removed a no-op `MinLevel` default assignment.
