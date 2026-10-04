# enim-core

Library Go bersama (shared core) untuk seluruh service di ISP-Project
(`service-billing`, `service-customer`, `service-payment`, dan lainnya).
Tujuannya menyatukan hal-hal lintas service — logging, error, response, dan
konstanta — supaya setiap service tidak menulis ulang hal yang sama.

- **Module path:** `github.com/enim-net/enim-core`
- **Go:** 1.26+

## Isi Paket

| Paket | Status | Keterangan |
|---|---|---|
| [`logger`](./logger) | Tersedia | Structured logger berbasis adapter, field dari `context`, logger default global |
| [`logger/sentry`](./logger/sentry) | Tersedia | Adapter untuk mengirim log ke Sentry |
| `error` | Rencana | Tipe error standar antar service |
| `responses` | Rencana | Format response HTTP standar |
| `constants` | Rencana | Konstanta bersama |

## Instalasi

```bash
go get github.com/enim-net/enim-core
```

Untuk pengembangan lokal (repo belum dipublikasikan), arahkan module ke folder
lokal lewat `replace` di `go.mod` service:

```go
require github.com/enim-net/enim-core v0.0.0

replace github.com/enim-net/enim-core => ../enim-core
```

## Logger

### Konsep

- **`Logger`** — menerima log, menggabungkan field, lalu meneruskan `Entry` ke
  satu atau lebih **`Adapter`**.
- **`Adapter`** — tujuan output (console, Sentry, atau custom). Interface-nya:

  ```go
  type Adapter interface {
      Write(ctx context.Context, e Entry) error
      Close(ctx context.Context) error
  }
  ```

- **Level** — `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`.
  Gunakan `logger.ParseLevel("warn")` untuk membaca level dari env/config.

Urutan field pada setiap entry:

1. Field dasar dari `Options` (`service`, `env`, `Fields`)
2. Field dari `logger.With(...)`
3. Field dari `context` (`logger.WithFields(ctx, ...)`)
4. Field yang dikirim saat memanggil log

Field `logger.Err(err)` diangkat ke `Entry.Error`; error `nil` diabaikan
sehingga aman dipanggil tanpa pengecekan.

### Penggunaan Cepat

Tanpa konfigurasi, logger default menulis format teks ke stdout dengan level
`info`:

```go
import "github.com/enim-net/enim-core/logger"

logger.Info(ctx, "server started", logger.Int("port", 8080))
logger.Error(ctx, "gagal memproses pembayaran", logger.Err(err))
```

### Konfigurasi di `main`

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/enim-net/enim-core/logger"
    "github.com/enim-net/enim-core/logger/sentry"
)

func main() {
    level, _ := logger.ParseLevel(os.Getenv("LOG_LEVEL"))

    adapters := []logger.Adapter{
        logger.NewConsole(logger.ConsoleOptions{Format: logger.FormatJSON}),
    }

    if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
        s, err := sentry.New(sentry.Config{
            DSN:         dsn,
            Environment: os.Getenv("APP_ENV"),
            Release:     os.Getenv("APP_VERSION"),
            MinLevel:    logger.LevelError,
        })
        if err == nil {
            adapters = append(adapters, s)
        }
    }

    log := logger.New(logger.Options{
        Level:     level,
        Service:   "service-billing",
        Env:       os.Getenv("APP_ENV"),
        AddCaller: true,
    }, adapters...)
    logger.SetDefault(log)

    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = log.Close(ctx) // flush adapter (mis. Sentry) sebelum keluar
    }()

    // ...
}
```

### Field dari Context

Cocok untuk middleware HTTP: pasang `request_id` sekali, otomatis ikut di semua
log selama request.

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := logger.WithFields(r.Context(),
            logger.String("request_id", r.Header.Get("X-Request-ID")),
        )
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Child Logger

```go
invoiceLog := logger.L().With(logger.String("module", "invoice"))
invoiceLog.Info(ctx, "invoice dibuat", logger.Int64("invoice_id", id))
```

### Helper Field

`String`, `Int`, `Int64`, `Uint`, `Float64`, `Bool`, `Duration`, `Time`, `Any`,
dan `Err`.

### Adapter

**Console** — `logger.NewConsole(logger.ConsoleOptions{...})`

| Opsi | Default | Keterangan |
|---|---|---|
| `Format` | `FormatText` | `FormatText` atau `FormatJSON` (berbasis `log/slog`) |
| `Writer` | `os.Stdout` | `io.Writer` tujuan |

**Sentry** — `sentry.New(sentry.Config{...})`

| Opsi | Default | Keterangan |
|---|---|---|
| `DSN` | — (wajib) | DSN project Sentry |
| `Environment` | `""` | Environment Sentry |
| `Release` | `""` | Versi aplikasi |
| `MinLevel` | `LevelDebug` | Level minimum yang dikirim; disarankan `LevelError` |
| `SampleRate` | `1.0` | Rasio event yang dikirim |
| `FlushTimeout` | `3s` | Batas waktu flush saat `Close` (deadline `ctx` diutamakan) |
| `Debug` | `false` | Mode debug SDK Sentry |

Entry dengan error dikirim sebagai `CaptureException` (pesan log disimpan di
context `log`); tanpa error dikirim sebagai `CaptureMessage`. Field dikirim
sebagai context `fields`, caller sebagai tag `caller`.

**Filter level per adapter** — `logger.MinLevel(level, adapter)`:

```go
logger.MinLevel(logger.LevelWarn, logger.NewConsole(logger.ConsoleOptions{}))
```

**Adapter custom** — implementasikan `Adapter`, atau pakai `AdapterFunc` untuk
kasus sederhana:

```go
hook := logger.AdapterFunc(func(ctx context.Context, e logger.Entry) error {
    metrics.Inc("log_" + e.Level.String())
    return nil
})
```

`Entry.FieldMap()` tersedia untuk mengubah field menjadi `map[string]any`
(error, `time.Duration`, dan `fmt.Stringer` dinormalisasi ke string).

### Testing

Gunakan `logger.Nop()` untuk logger yang tidak menulis apa pun.

## Pengembangan

```bash
go build ./...
go vet ./...
go test ./...
```

Catat perubahan di [CHANGELOG.md](./CHANGELOG.md).
