package errs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/enim-net/enim-core/dictionary"
)

func TestSetPrefix(t *testing.T) {
	t.Cleanup(func() { SetPrefix("APP") })
	for _, p := range []string{"ISP", "AB", "BILLING1", "A1"} {
		SetPrefix(p) // must not panic (regression: stray % in the regex)
		if got := ErrNotFound.String(); got != p+"-04-000" {
			t.Errorf("prefix %s: code = %s", p, got)
		}
	}
	for _, p := range []string{"", "A", "isp", "TOOLONGPREFIX", "1SP", "IS-P"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("SetPrefix(%q) did not panic", p)
				}
			}()
			SetPrefix(p)
		}()
	}
}

func TestCategoryHTTPStatus(t *testing.T) {
	cases := map[Code]int{
		ErrInternal:        http.StatusInternalServerError, // regression: was 200
		ErrValidation:      http.StatusBadRequest,
		ErrUnauthenticated: http.StatusUnauthorized,
		ErrForbidden:       http.StatusForbidden,
		ErrNotFound:        http.StatusNotFound,
		ErrDuplicate:       http.StatusConflict,
		ErrBusinessRule:    http.StatusUnprocessableEntity,
		ErrDatabase:        http.StatusInternalServerError,
		ErrExternalTimeout: http.StatusBadGateway,
		ErrTooManyRequests: http.StatusTooManyRequests,
	}
	for code, want := range cases {
		if got := code.HTTPStatus(); got != want {
			t.Errorf("%s (%s) HTTP = %d, want %d", code, code.Key(), got, want)
		}
	}
}

func TestDeprecatedAliases(t *testing.T) {
	if ErrUnauthentiCategoryed != ErrUnauthenticated || ErrDuplicateCategory != ErrDuplicate {
		t.Fatal("deprecated aliases must be the same codes")
	}
	if ErrUnauthenticated.Key() != "error.auth.unauthenticated" || ErrDuplicate.Key() != "error.duplicate" {
		t.Fatal("keys still damaged")
	}
}

func TestDefineRules(t *testing.T) {
	c := Define(CategoryNotFound, 901, "test.thing.not_found")
	if c.ID() != "04-901" || c.Category() != CategoryNotFound {
		t.Fatalf("Define: %s %v", c.ID(), c.Category())
	}
	mustPanic(t, "core range", func() { Define(CategoryNotFound, 99, "x") })
	mustPanic(t, "above range", func() { Define(CategoryNotFound, 1000, "x") })
	mustPanic(t, "duplicate", func() { Define(CategoryNotFound, 901, "test.other") })
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected panic", name)
		}
	}()
	fn()
}

func TestWrapIsAsFrom(t *testing.T) {
	cause := sql.ErrConnDone
	e := ErrDatabase.Wrap(cause, nil)
	wrapped := fmt.Errorf("repo: %w", e)

	if !errors.Is(wrapped, ErrDatabase) || errors.Is(wrapped, ErrNotFound) {
		t.Fatal("errors.Is by code")
	}
	if !errors.Is(wrapped, cause) {
		t.Fatal("cause must stay reachable for errors.Is")
	}
	var got *Error
	if !errors.As(wrapped, &got) || got.Code != ErrDatabase {
		t.Fatal("errors.As")
	}
	if From(wrapped) != e {
		t.Fatal("From must unwrap to the same *Error")
	}
	if f := From(ErrNotFound); f.Code != ErrNotFound {
		t.Fatal("From(bare Code)")
	}
	if f := From(errors.New("boom")); f.Code != ErrInternal || !errors.Is(f, ErrInternal) || f.Unwrap() == nil {
		t.Fatal("From(unknown) must wrap as ErrInternal")
	}
	if From(nil) != nil {
		t.Fatal("From(nil)")
	}
}

func TestResponseIsLocalizedAndHidesCause(t *testing.T) {
	e := Wrap(errors.New(`pq: relation "customers" does not exist`), ErrDatabase, nil)
	for _, l := range []dictionary.Locale{dictionary.LocaleEN, dictionary.LocaleID} {
		r := e.Response(l)
		if r.Code != "APP-07-000" || r.Message == "" || r.Message == "error.database" {
			t.Errorf("%s: %+v", l, r)
		}
		if strings.Contains(r.Message, "relation") {
			t.Errorf("%s: cause leaked into message", l)
		}
	}
	if en, id := e.Response(dictionary.LocaleEN).Message, e.Response(dictionary.LocaleID).Message; en == id {
		t.Errorf("messages not localized: %q", en)
	}
}

func TestValidationFields(t *testing.T) {
	v := Validation().Add("email", ErrInvalidEmail, P{"field": "email"}).Add("name", ErrRequired, P{"field": "name"})
	if !v.HasFields() || v.HTTPStatus() != http.StatusBadRequest {
		t.Fatal("validation error")
	}
	r := v.Response(dictionary.LocaleEN)
	if len(r.Errors) != 2 || r.Errors[0].Field != "email" || r.Errors[0].Code != "APP-01-005" {
		t.Fatalf("%+v", r.Errors)
	}
	if r.Errors[1].Message != "name is required." {
		t.Fatalf("message = %q", r.Errors[1].Message)
	}
	if s := fmt.Sprintf("%+v", v); !strings.Contains(s, "email: APP-01-005") {
		t.Fatalf("%%+v = %q", s)
	}
}

func TestEveryCoreCodeHasMessages(t *testing.T) {
	for _, c := range []Code{ErrInternal, ErrValidation, ErrRequired, ErrInvalidFormat, ErrMinLength,
		ErrMaxLength, ErrInvalidEmail, ErrInvalidPhone, ErrOutOfRange, ErrInvalidOption, ErrMalformedBody,
		ErrUnauthenticated, ErrTokenExpired, ErrTokenInvalid, ErrInvalidCredentials, ErrForbidden,
		ErrNotFound, ErrConflict, ErrDuplicate, ErrBusinessRule, ErrDatabase, ErrExternalService,
		ErrExternalTimeout, ErrTooManyRequests} {
		for _, l := range []dictionary.Locale{dictionary.LocaleEN, dictionary.LocaleID} {
			if msg := dictionary.T(l, c.Key(), nil); msg == c.Key() {
				t.Errorf("%s has no %s message for key %s", c, l, c.Key())
			}
		}
	}
}
