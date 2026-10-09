package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/enim-net/enim-core/dictionary"
	"github.com/enim-net/enim-core/errs"
)

func app() *fiber.App {
	a := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	a.Get("/ok", func(c *fiber.Ctx) error {
		return OK(c, Schema{Message: "general.success"}, fiber.Map{"x": 1})
	})
	a.Post("/created", func(c *fiber.Ctx) error { return Created(c, fiber.Map{"id": "1"}, "general.create.success") })
	a.Get("/list", func(c *fiber.Ctx) error {
		return List(c, []int{1, 2}, fiber.Map{"page": 1}, "general.list.success")
	})
	a.Get("/notfound", func(c *fiber.Ctx) error { return errs.New(errs.ErrNotFound, nil) })
	a.Get("/wrapped", func(c *fiber.Ctx) error {
		return fmt.Errorf("svc: %w", errs.Wrap(errors.New(`pq: password authentication failed for user "app"`), errs.ErrDatabase, nil))
	})
	a.Get("/raw", func(c *fiber.Ctx) error {
		return errors.New(`dial tcp 10.0.0.5:5432: connect: connection refused`)
	})
	a.Get("/invalid", func(c *fiber.Ctx) error {
		return errs.Validation().Add("email", errs.ErrInvalidEmail, errs.P{"field": "email"})
	})
	a.Get("/apierr", func(c *fiber.Ctx) error {
		return NewError(fiber.StatusServiceUnavailable, "07", "Database unreachable")
	})
	return a
}

type env struct {
	Schema struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Errors  any    `json:"errors"`
	} `json:"schema"`
	Data       any `json:"data"`
	Pagination any `json:"pagination"`
}

func do(t *testing.T, method, path, lang string) (int, env, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}
	res, err := app().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	var e env
	if err := json.Unmarshal(raw, &e); err != nil {
		t.Fatalf("%s: not JSON: %s", path, raw)
	}
	return res.StatusCode, e, string(raw)
}

func TestSuccessEnvelopes(t *testing.T) {
	st, e, _ := do(t, "GET", "/ok", "")
	if st != 200 || e.Schema.Code != "00" || e.Schema.Message != "Operation completed successfully." {
		t.Errorf("OK: %d %+v", st, e.Schema)
	}
	st, e, _ = do(t, "GET", "/ok", "id-ID,id;q=0.9")
	if e.Schema.Message != "Proses berhasil dilakukan." {
		t.Errorf("browser Accept-Language not localized: %q", e.Schema.Message)
	}
	st, e, _ = do(t, "POST", "/created", "")
	if st != 201 || e.Schema.Message != "Your data has been created successfully." {
		t.Errorf("Created: %d %+v", st, e.Schema)
	}
	_, e, _ = do(t, "GET", "/list", "")
	if e.Pagination == nil {
		t.Error("List without pagination")
	}
}

func TestErrorsAreMappedAndNeverLeak(t *testing.T) {
	cases := []struct {
		path, code, leak string
		status           int
	}{
		{"/notfound", "APP-04-000", "", 404},
		{"/wrapped", "APP-07-000", "password authentication", 500},
		{"/raw", "APP-00-000", "10.0.0.5", 500},
		{"/invalid", "APP-01-000", "", 400},
		{"/apierr", "07", "", 503},
		{"/no-such-route", "00", "", 404},
	}
	for _, c := range cases {
		st, e, raw := do(t, "GET", c.path, "")
		if st != c.status || e.Schema.Code != c.code {
			t.Errorf("%s: %d %q, want %d %q", c.path, st, e.Schema.Code, c.status, c.code)
		}
		if c.leak != "" && strings.Contains(raw, c.leak) {
			t.Errorf("%s leaked internal detail: %s", c.path, raw)
		}
	}
	_, e, _ := do(t, "GET", "/invalid", "id_ID")
	fields, ok := e.Schema.Errors.([]any)
	if !ok || len(fields) != 1 || fields[0].(map[string]any)["message"] != "email harus berupa email yang valid." {
		t.Fatalf("field errors: %#v", e.Schema.Errors)
	}
	_, e, _ = do(t, "GET", "/raw", "")
	if e.Schema.Message != dictionary.T(dictionary.LocaleEN, "error.internal", nil) {
		t.Errorf("internal message = %q", e.Schema.Message)
	}
}
