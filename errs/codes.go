package errs

// Generic codes shipped with go-core (numbers 000-099).
// Messages live in go-core/dictionary/locales under the "error." keys;
// apps can override any of them in their own locale files.
var (
	// 00 general
	ErrInternal = defineCore(CatGeneral, 0, "error.internal")
	
	// 01 validation
	ErrValidation    = defineCore(CatValidation, 0, "error.validation.failed")
	ErrRequired      = defineCore(CatValidation, 1, "error.validation.required")
	ErrInvalidFormat = defineCore(CatValidation, 2, "error.validation.invalid_format")
	ErrMinLength     = defineCore(CatValidation, 3, "error.validation.min_length")
	ErrMaxLength     = defineCore(CatValidation, 4, "error.validation.max_length")
	ErrInvalidEmail  = defineCore(CatValidation, 5, "error.validation.invalid_email")
	ErrInvalidPhone  = defineCore(CatValidation, 6, "error.validation.invalid_phone")
	ErrOutOfRange    = defineCore(CatValidation, 7, "error.validation.out_of_range")
	ErrInvalidOption = defineCore(CatValidation, 8, "error.validation.invalid_option")
	ErrMalformedBody = defineCore(CatValidation, 9, "error.validation.malformed_body")
	
	// 02 authentication
	ErrUnauthenticated    = defineCore(CatAuth, 0, "error.auth.unauthenticated")
	ErrTokenExpired       = defineCore(CatAuth, 1, "error.auth.token_expired")
	ErrTokenInvalid       = defineCore(CatAuth, 2, "error.auth.token_invalid")
	ErrInvalidCredentials = defineCore(CatAuth, 3, "error.auth.invalid_credentials")
	
	// 03 authorization
	ErrForbidden = defineCore(CatForbidden, 0, "error.forbidden")
	
	// 04 not found
	ErrNotFound = defineCore(CatNotFound, 0, "error.not_found")
	
	// 05 conflict
	ErrConflict  = defineCore(CatConflict, 0, "error.conflict")
	ErrDuplicate = defineCore(CatConflict, 1, "error.duplicate")
	
	// 06 business rule
	ErrBusinessRule = defineCore(CatBusiness, 0, "error.business_rule")
	
	// 07 database
	ErrDatabase = defineCore(CatDatabase, 0, "error.database")
	
	// 08 external service
	ErrExternalService = defineCore(CatExternal, 0, "error.external.failed")
	ErrExternalTimeout = defineCore(CatExternal, 1, "error.external.timeout")
	
	// 09 rate limit
	ErrTooManyRequests = defineCore(CatRateLimit, 0, "error.too_many_requests")
)
