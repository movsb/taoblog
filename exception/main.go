package exception

type HTTPError struct {
	Code    int         `json:"-"`
	Message string      `json:"message,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
}

type IHTTPError interface {
	ToHTTPError() *HTTPError
}

type ValidationError struct {
	Message string
}

func NewValidationError(message string) ValidationError {
	return ValidationError{
		Message: message,
	}
}

func (e ValidationError) ToHTTPError() *HTTPError {
	return &HTTPError{
		Code:    400,
		Message: e.Message,
	}
}
