package httpx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
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
func messageFor(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", e.Param())
	case "alphanum":
		return "must contain only letters and numbers"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", e.Param())
	default:
		return fmt.Sprintf("failed %q validation", e.Tag())
	}
}
