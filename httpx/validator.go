package httpx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/enim-net/enim-core/errs"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" {
			return ""
		}
		return name
	})
	return v
}

// Validate validates s and returns field -> failed tag, or nil.
// Prefer ValidateErr, which returns a localized *errs.Error.
func Validate(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var verrs validator.ValidationErrors
	if !asValidationErrors(err, &verrs) {
		return map[string]string{"_": err.Error()}
	}

	fields := make(map[string]string, len(verrs))
	for _, e := range verrs {
		fields[e.Field()] = e.Tag()
	}

	return fields
}

func asValidationErrors(err error, dst *validator.ValidationErrors) bool {
	var v validator.ValidationErrors
	ok := errors.As(err, &v)

	if ok {
		*dst = v
	}
	return ok
}

// ValidateErr validates s and returns nil or an *errs.Error (category
// validation, HTTP 400) with one localized field error per failed rule.
// Field names follow the json tags.
func ValidateErr(s any) *errs.Error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !asValidationErrors(err, &verrs) {
		return errs.Wrap(err, errs.ErrValidation, nil)
	}
	out := errs.Validation()
	for _, fe := range verrs {
		code, params := fieldCode(fe)
		out.Add(fe.Field(), code, params)
	}
	return out
}

// fieldCode maps a validator tag to a core error code and message params.
func fieldCode(e validator.FieldError) (errs.Code, errs.P) {
	p := errs.P{"field": e.Field()}
	switch e.Tag() {
	case "required", "required_if", "required_unless", "required_with", "required_without":
		return errs.ErrRequired, p
	case "email":
		return errs.ErrInvalidEmail, p
	case "e164":
		return errs.ErrInvalidPhone, p
	case "min":
		if isString(e) {
			p["min"] = e.Param()
			return errs.ErrMinLength, p
		}
		return errs.ErrOutOfRange, p
	case "max":
		if isString(e) {
			p["max"] = e.Param()
			return errs.ErrMaxLength, p
		}
		return errs.ErrOutOfRange, p
	case "gt", "gte", "lt", "lte":
		return errs.ErrOutOfRange, p
	case "oneof":
		p["options"] = strings.Join(strings.Fields(e.Param()), ", ")
		return errs.ErrInvalidOption, p
	default:
		return errs.ErrInvalidFormat, p
	}
}

func isString(e validator.FieldError) bool { return e.Kind() == reflect.String }
