package httpx

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/enim-net/enim-core/dictionary"
	"github.com/enim-net/enim-core/errs"
)

func TestParsePagination(t *testing.T) {
	cases := map[string]Pagination{
		"":                   {1, 10},
		"?page=3&limit=20":   {3, 20},
		"?page=0&limit=0":    {1, 10},
		"?page=-2&limit=500": {1, 100},
		"?page=x&limit=y":    {1, 10},
	}
	for q, want := range cases {
		a := fiber.New()
		var got Pagination
		a.Get("/", func(c *fiber.Ctx) error { got = ParsePagination(c); return nil })
		if _, err := a.Test(httptest.NewRequest("GET", "/"+q, nil)); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%q: %+v, want %+v", q, got, want)
		}
	}
	if (Pagination{Page: 3, Limit: 20}).Offset() != 40 {
		t.Error("Offset")
	}
}

func TestNewPageMeta(t *testing.T) {
	if m := NewPageMeta(Pagination{Page: 1, Limit: 10}, 41); m.MaxPage != 5 || m.Total != 41 {
		t.Errorf("%+v", m)
	}
	if m := NewPageMeta(Pagination{}, 41); m.MaxPage != 0 { // regression: division by zero
		t.Errorf("%+v", m)
	}
}

type body struct {
	Name  string `json:"name" validate:"required,min=3,max=10"`
	Email string `json:"email" validate:"required,email"`
	Plan  string `json:"plan" validate:"oneof=basic pro"`
	Age   int    `json:"age" validate:"gte=18"`
}

func TestBodyParser(t *testing.T) {
	cases := map[string]bool{
		`{"name":"abc","email":"a@b.co","plan":"pro","age":20}`: true,
		``:                         false,
		`{"name":"abc","extra":1}`: false,
		`{"name":"abc"}{"x":1}`:    false,
		`not json`:                 false,
	}
	for in, ok := range cases {
		a := fiber.New()
		var err error
		a.Post("/", func(c *fiber.Ctx) error { var b body; err = BodyParser(c, &b); return nil })
		req := httptest.NewRequest("POST", "/", strings.NewReader(in))
		req.Header.Set("Content-Type", "application/json")
		if _, e := a.Test(req); e != nil {
			t.Fatal(e)
		}
		if (err == nil) != ok {
			t.Errorf("%q: err = %v", in, err)
		}
		if err != nil && (!errors.Is(err, errs.ErrMalformedBody) || errs.From(err).HTTPStatus() != 400) {
			t.Errorf("%q: want ErrMalformedBody/400, got %v", in, err)
		}
	}
}

func TestValidateAndValidateErr(t *testing.T) {
	good := body{Name: "abcd", Email: "a@b.co", Plan: "basic", Age: 20}
	if Validate(good) != nil || ValidateErr(good) != nil {
		t.Fatal("valid body rejected")
	}
	bad := body{Name: "ab", Email: "nope", Plan: "gold", Age: 10}
	if m := Validate(bad); m["name"] != "min" || m["email"] != "email" {
		t.Fatalf("Validate = %v", m)
	}
	e := ValidateErr(bad)
	if e == nil || e.HTTPStatus() != 400 || len(e.Fields) != 4 {
		t.Fatalf("ValidateErr = %+v", e)
	}
	r := e.Response(dictionary.LocaleEN)
	got := map[string]string{}
	for _, f := range r.Errors {
		got[f.Field] = f.Message
	}
	want := map[string]string{
		"name":  "name must be at least 3 characters.",
		"email": "email must be a valid email address.",
		"plan":  "plan must be one of: basic, pro.",
		"age":   "age is out of the allowed range.",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %q, want %q", k, got[k], v)
		}
	}
	if e := ValidateErr(body{}); e == nil || e.Fields[0].Code != errs.ErrRequired {
		t.Errorf("required: %+v", e)
	}
}
