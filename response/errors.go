package response

import "github.com/enim-net/enim-core/dictionary"

type APIError struct {
	Status  int
	Code    string
	Message string
	Params  dictionary.Params
	Fields  map[string]string
}

func (e *APIError) Error() string {
	return e.Message
}

func NewError(status int, code string, message string) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func (e *APIError) WithFields(fields map[string]string) *APIError {
	e.Fields = fields
	return e
}
