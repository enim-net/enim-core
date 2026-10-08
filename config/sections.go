package config

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type App struct {
	Name string `env:"APP_NAME" default:"app"`
	Env  string `env:"APP_ENV" default:"development"`
	Port string `env:"PORT" default:"8080"`
}

func (a App) IsProduction() bool {
	return strings.EqualFold(a.Env, "production")
}
func (a App) IsDevelopment() bool {
	return strings.EqualFold(a.Env, "development")
}
func (a App) IsRegression() bool {
	return strings.EqualFold(a.Env, "regression")
}
func (a App) IsSandbox() bool { return strings.EqualFold(a.Env, "sandbox") }

// DB is a PostgreSQL connection. Use with `prefix:"DB_"` for DB_HOST, DB_PORT...
type DB struct {
	Host     string `env:"DB_HOST" default:"localhost"`
	Port     int    `env:"DB_PORT" default:"5432"`
	User     string `env:"DB_USER" default:"postgres"`
	Password string `env:"DB_PASSWORD" secret:"true"`
	Name     string `env:"DB_NAME" required:"true"`
	SSLMode  string `env:"DB_SSLMODE" default:"disable"`
	TimeZone string `env:"DB_TIMEZONE" default:"Asia/Jakarta"`
	
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" default:"1h"`
}

func (d DB) DSN() string {
	parts := []string{
		"host=" + quote(d.Host),
		"port=" + strconv.Itoa(d.Port),
		"user=" + quote(d.User),
	}
	if d.Password != "" {
		parts = append(parts, "password="+quote(d.Password))
	}
	parts = append(parts, "dbname="+quote(d.Name), "sslmode="+quote(d.SSLMode))
	if d.TimeZone != "" {
		parts = append(parts, "TimeZone="+quote(d.TimeZone))
	}
	return strings.Join(parts, " ")
}

type JWT struct {
	Secret     string        `env:"JWT_SECRET" required:"true" secret:"true"`
	Issuer     string        `env:"JWT_ISSUER"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" default:"720h"`
}

func (j JWT) Validate() error {
	if j.Secret != "" && len(j.Secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters (generate one with: openssl rand -base64 58)")
	}
	if j.AccessTTL > 0 && j.RefreshTTL > 0 && j.RefreshTTL <= j.AccessTTL {
		return errors.New("JWT_REFRESH_TTL must be greater than ACCESS_TTL")
	}
	return nil
}

func quote(s string) string {
	if s != "" && !strings.ContainsAny(s, " '\\\t\n") {
		return s
	}
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + r.Replace(s) + "'"
}
