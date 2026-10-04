package errs

import "net/http"

type Category struct {
	Code       string
	Name       string
	HTTPStatus int
}

var (
	CategoryGeneral    = Category{"00", "general", http.StatusInternalServerError}
	CategoryValidation = Category{"01", "validation", http.StatusBadRequest}
	CategoryAuth       = Category{"02", "authentication", http.StatusUnauthorized}
	CategoryForbidden  = Category{"03", "authorization", http.StatusForbidden}
	CategoryNotFound   = Category{"04", "not_found", http.StatusNotFound}
	CategoryConflict   = Category{"05", "conflict", http.StatusConflict}
	CategoryBusiness   = Category{"06", "business_rule", http.StatusUnprocessableEntity}
	CategoryDatabase   = Category{"07", "database", http.StatusInternalServerError}
	CategoryExternal   = Category{"08", "external_service", http.StatusBadGateway}
	CategoryRateLimit  = Category{"09", "rate_limit", http.StatusTooManyRequests}
)
