package errors

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternalServer     = errors.New("internal server error")
	ErrDuplicateEntry     = errors.New("duplicate entry")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AppError struct {
	Err     error
	Message string
	Code    int
}

func NewAppError(err error, message string, code int) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}