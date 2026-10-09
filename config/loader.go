package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/enim-net/enim-core/utils"
)

type Option func(*options)
type options struct {
	files []string
}

func WithFiles(files ...string) Option {
	return func(o *options) {
		o.files = files
	}
}

type Validator interface {
	Validate() error
}

func Load(dst any, opts ...Option) error {
	o := options{files: []string{".env"}}
	for _, fn := range opts {
		fn(&o)
	}
	if len(o.files) > 0 {
		if err := utils.LoadDotEnv(o.files...); err != nil {
			return fmt.Errorf("config: %w", err)
		}
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return errors.New("config: Load needs a pointer to a struct")
	}

	var errs []error
	fill(rv.Elem(), "", &errs)
	validate(rv.Elem(), &errs)
	if len(errs) > 0 {
		return fmt.Errorf("config: invalid configuration:\n - %w", joinLines(errs))
	}
	return nil
}

func MustLoad(dst any, opts ...Option) {
	if err := Load(dst, opts...); err != nil {
		panic(err)
	}
}

var durationType = reflect.TypeOf(time.Duration(0))

func fill(v reflect.Value, prefix string, errs *[]error) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		fv := v.Field(i)
		key, hasEnv := f.Tag.Lookup("env")
		if !hasEnv {
			if fv.Kind() == reflect.Struct {
				fill(fv, prefix+f.Tag.Get("prefix"), errs)
			}
			continue
		}
		key = prefix + key
		raw, ok := os.LookupEnv(key)
		raw = strings.TrimSpace(raw)

		switch {
		case ok && raw != "":
		// environment wins
		case !fv.IsZero():
			continue
		case f.Tag.Get("default") != "":
			raw = f.Tag.Get("default")
		default:
			if f.Tag.Get("required") == "true" {
				*errs = append(*errs, fmt.Errorf("config: field %s is required", key))
			}
			continue
		}

		if err := set(fv, raw); err != nil {
			*errs = append(*errs, fmt.Errorf("%s: %v", key, err))
		}
	}
}

func set(fv reflect.Value, raw string) error {
	if fv.Type() == durationType {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("invalid duration %q (use e.g. 15m, 1h, 720h)", raw)
		}
		fv.SetInt(int64(d))
		return nil
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid bool %q (use true/false)", raw)
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer %q", raw)
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer %q", raw)
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(raw, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid number %q", raw)
		}
		fv.SetFloat(n)
	case reflect.Slice:
		if fv.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("unsupported slice type %s", fv.Type())
		}
		var out []string
		for _, s := range strings.Split(raw, ",") {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
		fv.Set(reflect.ValueOf(out))
	default:
		return fmt.Errorf("unsupported field type %s", fv.Type())
	}
	return nil
}

// validate walks nested sections first, then the section itself.
func validate(v reflect.Value, errs *[]error) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).IsExported() && v.Field(i).Kind() == reflect.Struct {
			if _, isEnv := t.Field(i).Tag.Lookup("env"); !isEnv {
				validate(v.Field(i), errs)
			}
		}
	}
	var target any
	if v.CanAddr() {
		target = v.Addr().Interface() // pointer receivers
	} else {
		target = v.Interface()
	}
	if val, ok := target.(Validator); ok {
		if err := val.Validate(); err != nil {
			*errs = append(*errs, err)
		}
	}
}

// Print writes the loaded config as KEY=value lines, masking secret fields.
// Handy at startup in development:
//
//	if !cfg.App.IsProduction() { config.Print(os.Stdout, &cfg) }
func Print(w io.Writer, cfg any) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		printFields(w, v, "")
	}
}

func printFields(w io.Writer, v reflect.Value, prefix string) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		fv := v.Field(i)
		key, hasEnv := f.Tag.Lookup("env")
		if !hasEnv {
			if fv.Kind() == reflect.Struct {
				printFields(w, fv, prefix+f.Tag.Get("prefix"))
			}
			continue
		}
		val := fmt.Sprint(fv.Interface())
		if f.Tag.Get("secret") == "true" && !fv.IsZero() {
			val = "********"
		}
		fmt.Fprintf(w, "%s=%s\n", prefix+key, val)
	}
}

func joinLines(errs []error) error {
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return errors.New(strings.Join(msgs, "\n  - "))
}
