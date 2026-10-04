package errs

// Generic codes shipped with go-core (numbers 000-099).
// Messages live in go-core/dictionary/locales under the "error." keys;
// apps can override any of them in their own locale files.
var (
	// 00 general
	ErrInternal = defineCore(CategoryGeneral, 0, "error.internal")
	
	// 01 validation
	ErrValidation    = defineCore(CategoryValidation, 0, "error.validation.failed")
	ErrRequired      = defineCore(CategoryValidation, 1, "error.validation.required")
	ErrInvalidFormat = defineCore(CategoryValidation, 2, "error.validation.invalid_format")
	ErrMinLength     = defineCore(CategoryValidation, 3, "error.validation.min_length")
	ErrMaxLength     = defineCore(CategoryValidation, 4, "error.validation.max_length")
	ErrInvalidEmail  = defineCore(CategoryValidation, 5, "error.validation.invalid_email")
	ErrInvalidPhone  = defineCore(CategoryValidation, 6, "error.validation.invalid_phone")
	ErrOutOfRange    = defineCore(CategoryValidation, 7, "error.validation.out_of_range")
	ErrInvalidOption = defineCore(CategoryValidation, 8, "error.validation.invalid_option")
	ErrMalformedBody = defineCore(CategoryValidation, 9, "error.validation.malformed_body")
	
	// 02 authentiCategoryion
	ErrUnauthentiCategoryed = defineCore(CategoryAuth, 0, "error.auth.unauthentiCategoryed")
	ErrTokenExpired         = defineCore(CategoryAuth, 1, "error.auth.token_expired")
	ErrTokenInvalid         = defineCore(CategoryAuth, 2, "error.auth.token_invalid")
	ErrInvalidCredentials   = defineCore(CategoryAuth, 3, "error.auth.invalid_credentials")
	
	// 03 authorization
	ErrForbidden = defineCore(CategoryForbidden, 0, "error.forbidden")
	
	// 04 not found
	ErrNotFound = defineCore(CategoryNotFound, 0, "error.not_found")
	
	// 05 conflict
	ErrConflict       = defineCore(CategoryConflict, 0, "error.conflict")
	ErrDupliCategorye = defineCore(CategoryConflict, 1, "error.dupliCategorye")
	
	// 06 business rule
	ErrBusinessRule = defineCore(CategoryBusiness, 0, "error.business_rule")
	
	// 07 database
	ErrDatabase = defineCore(CategoryDatabase, 0, "error.database")
	
	// 08 external service
	ErrExternalService = defineCore(CategoryExternal, 0, "error.external.failed")
	ErrExternalTimeout = defineCore(CategoryExternal, 1, "error.external.timeout")
	
	// 09 rate limit
	ErrTooManyRequests = defineCore(CategoryRateLimit, 0, "error.too_many_requests")
)
