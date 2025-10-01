package error_utils

import "fmt"

type ErrorType string

const (
	ErrorCodeNotFound      ErrorType = "ErrorCodeNotFound"
	ErrorCodeBadRequest    ErrorType = "ErrorCodeBadRequest"
	ErrorCodeBusinessLogic ErrorType = "ErrorCodeBusinessLogic"
	ErrorCodeUnexpected    ErrorType = "ErrorCodeUnexpected"
)

type CustomError struct {
	ErrorType ErrorType
	Code      string
	Message   string
	Err       error
}

func (c *CustomError) Error() string {
	if c.Message == "" {
		return c.Code
	}

	return fmt.Sprintf("%s: %s", c.Code, c.Message)
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

func ErrValidationError(msg string) error {
	return &CustomError{
		ErrorType: ErrorCodeBadRequest,
		Code:      "ValidationError",
		Message:   msg,
	}
}

func ErrInternalServerError(msg string) error {
	return &CustomError{
		ErrorType: ErrorCodeUnexpected,
		Code:      "InternalServerError",
		Message:   msg,
	}
}

func ErrBusinessLogic(code string) error {
	return &CustomError{
		ErrorType: ErrorCodeBusinessLogic,
		Code:      code,
	}
}
