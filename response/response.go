package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/enim-net/enim-core/dictionary"
	"github.com/enim-net/enim-core/errs"
	"github.com/enim-net/enim-core/logger"
)

func Locale(c *fiber.Ctx) dictionary.Locale {
	if v, ok := c.Locals("locale").(dictionary.Locale); ok && dictionary.Has(v) {
		return v
	}
	if l, ok := dictionary.Match(c.Get(fiber.HeaderAcceptLanguage)); ok {
		return l
	}
	return dictionary.DefaultLocale
}

func Msg(c *fiber.Ctx, keyOrText string, params dictionary.Params) string {
	return dictionary.T(Locale(c), keyOrText, params)
}

type Schema struct {
	Errors  any    `json:"errors"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Envelope struct {
	Schema     Schema `json:"schema"`
	Data       any    `json:"data"`
	Pagination any    `json:"pagination,omitempty"`
}

func JSON(c *fiber.Ctx, status int, env Envelope) error {
	return c.Status(status).JSON(env)
}

func OK(c *fiber.Ctx, s Schema, data any) error {
	code := s.Code
	msg := ""
	if code == "" {
		code = errs.CategoryGeneral.Code
	}
	convMsg := Msg(c, s.Message, nil)
	if len(convMsg) > 0 {
		msg = convMsg
	} else {
		msg = s.Message
	}

	return JSON(c, fiber.StatusOK, Envelope{
		Schema: Schema{
			Errors:  s.Errors,
			Message: msg,
			Code:    code,
		},
		Data: data,
	})
}
func Created(c *fiber.Ctx, data any, message string, params ...dictionary.Params) error {
	return JSON(c, fiber.StatusCreated, Envelope{
		Schema: Schema{
			Message: Msg(c, message, firstParams(params)),
			Code:    errs.CategoryGeneral.Code,
		},
		Data: data,
	})
}

func List(c *fiber.Ctx, data any, pagination any, message string, params ...dictionary.Params) error {
	return JSON(c, fiber.StatusOK, Envelope{
		Schema: Schema{
			Code:    errs.CategoryGeneral.Code,
			Message: Msg(c, message, firstParams(params)),
		},
		Data:       data,
		Pagination: pagination,
	})
}

// ErrorHandler is a fiber.Config.ErrorHandler. It renders *APIError,
// *fiber.Error and *errs.Error; any other error becomes errs.ErrInternal.
// Causes are logged, never sent to the client.
func ErrorHandler(c *fiber.Ctx, err error) error {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		var errBody any
		if apiErr.Fields != nil {
			errBody = apiErr.Fields
		}
		return JSON(c, apiErr.Status, Envelope{
			Schema: Schema{
				Code:    apiErr.Code,
				Errors:  errBody,
				Message: Msg(c, apiErr.Message, apiErr.Params),
			},
			Data: nil,
		})
	}

	if fe, ok := errors.AsType[*fiber.Error](err); ok {
		return JSON(c, fe.Code, Envelope{
			Schema: Schema{
				Message: fe.Message,
				Code:    errs.CategoryGeneral.Code,
			},
			Data: nil,
		})
	}

	e := errs.From(err)
	status := e.HTTPStatus()
	fields := []logger.Field{
		logger.String("code", e.Code.String()),
		logger.String("method", c.Method()),
		logger.String("path", c.Path()),
		logger.Int("status", status),
		logger.Err(err),
	}
	if status >= fiber.StatusInternalServerError {
		logger.Error(c.UserContext(), "request failed", fields...)
	} else {
		logger.Warn(c.UserContext(), "request rejected", fields...)
	}

	body := e.Response(Locale(c))
	var fieldErrs any
	if len(body.Errors) > 0 {
		fieldErrs = body.Errors
	}
	return JSON(c, status, Envelope{
		Schema: Schema{Code: body.Code, Message: body.Message, Errors: fieldErrs},
		Data:   nil,
	})
}

// firstParams returns the first params map from a variadic argument, or nil.
func firstParams(params []dictionary.Params) dictionary.Params {
	if len(params) > 0 {
		return params[0]
	}
	return nil
}
