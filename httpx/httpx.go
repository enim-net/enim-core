package httpx

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/enim-net/enim-core/errs"
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PageMeta struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func ParsePagination(c *fiber.Ctx) Pagination {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

func NewPageMeta(p Pagination, total int64) PageMeta {
	maxPage := 0
	if p.Limit > 0 {
		maxPage = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}

	return PageMeta{
		Page:    p.Page,
		Limit:   p.Limit,
		MaxPage: maxPage,
		Total:   total,
	}
}

// BodyParser strictly decodes a JSON body into dst: the body must be
// non-empty, contain exactly one JSON value and no unknown fields. Failures
// are *errs.Error with code ErrMalformedBody (HTTP 400); the decoder error is
// kept as the cause for logs.
func BodyParser(c *fiber.Ctx, dst interface{}) error {
	body := c.Body()
	if len(body) == 0 {
		return errs.Wrap(errors.New("empty request body"), errs.ErrMalformedBody, nil)
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errs.Wrap(err, errs.ErrMalformedBody, nil)
	}
	if dec.More() {
		return errs.Wrap(errors.New("request body must contain a single JSON object"), errs.ErrMalformedBody, nil)
	}
	return nil
}
