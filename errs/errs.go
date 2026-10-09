// Package errs provides application errors with stable codes in the form
// PREFIX-CC-NNN (e.g. ISP-04-101), categories mapped to HTTP statuses,
// localized messages and hidden causes for logging.
package errs

import (
	"errors"
	"fmt"

	"github.com/enim-net/enim-core/dictionary"
)

// P is shorthand for message params.
type P = map[string]any

// Error is an application error with a code, message params, optional
// per-field details (validation) and an optional hidden cause.
type Error struct {
	Code   Code
	Params P
	Fields []FieldError
	cause  error
}

// FieldError is one invalid field inside a validation error.
type FieldError struct {
	Field  string
	Code   Code
	Params P
}

func New(code Code, params P) *Error { return &Error{Code: code, Params: params} }

func Wrap(cause error, code Code, params P) *Error {
	return &Error{Code: code, Params: params, cause: cause}
}

func Validation() *Error { return New(ErrValidation, nil) }

func (e *Error) Add(field string, code Code, params P) *Error {
	e.Fields = append(e.Fields, FieldError{Field: field, Code: code, Params: params})
	return e
}

func (e *Error) HasFields() bool { return len(e.Fields) > 0 }

func (e *Error) Error() string {
	s := e.Code.String() + " " + e.Code.key
	if e.cause != nil {
		s += ": " + e.cause.Error()
	}
	return s
}

func (e *Error) Unwrap() error   { return e.cause }
func (e *Error) HTTPStatus() int { return e.Code.HTTPStatus() }

func (e *Error) Is(target error) bool {
	switch t := target.(type) {
	case Code:
		return e.Code == t
	case *Error:
		return e.Code == t.Code
	}
	return false
}

// Message returns the localized message for this error.
func (e *Error) Message(locale dictionary.Locale) string {
	return dictionary.T(locale, e.Code.key, e.Params)
}

// From converts any error into *Error:
//   - *Error is returned as is (also when wrapped with fmt.Errorf %w)
//   - a bare Code becomes New(code, nil)
//   - anything else becomes ErrInternal wrapping the original
//
// nil returns nil.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	var c Code
	if errors.As(err, &c) {
		return New(c, nil)
	}
	return Wrap(err, ErrInternal, nil)
}

// Response is the JSON body sent to clients.
type Response struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Errors  []FieldResponse `json:"errors,omitempty"`
}

type FieldResponse struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response builds the localized client payload. The cause is never included.
func (e *Error) Response(locale dictionary.Locale) Response {
	r := Response{Code: e.Code.String(), Message: e.Message(locale)}
	for _, f := range e.Fields {
		r.Errors = append(r.Errors, FieldResponse{
			Field:   f.Field,
			Code:    f.Code.String(),
			Message: dictionary.T(locale, f.Code.key, f.Params),
		})
	}
	return r
}

// Format supports %+v to print fields too (handy in logs/tests).
func (e *Error) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		fmt.Fprintf(s, "%s", e.Error())
		for _, f := range e.Fields {
			fmt.Fprintf(s, "\n  %s: %s %v", f.Field, f.Code, f.Params)
		}
		return
	}
	fmt.Fprint(s, e.Error())
}
