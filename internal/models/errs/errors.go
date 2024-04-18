package errs

import (
	"errors"
	"fmt"
)

// Common sentinel errors.
var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("data conflict")
	ErrRateLimit       = errors.New("rate limit")
	ErrInvalidPassword = errors.New("invalid password")
)

// Type just for murshallig purpose.
// Should only be used immediately before marshalling.
type JSON struct {
	Error string `json:"error"`
}

// Let users know which required request parameter is not provided.
type RequiredJSONBodyParamError struct {
	ParamName string
}

func (e *RequiredJSONBodyParamError) Error() string {
	return fmt.Sprintf("JSON body argument %q is required, but not found", e.ParamName)
}

// Authorization errors wrapper.
type InvalidAuthorizationError struct {
	Message string
}

func (e *InvalidAuthorizationError) Error() string {
	return fmt.Sprintf("invalid authorization data: %s", e.Message)
}

// Provides details at which field unique violation has occurred.
type AlreadyExistsError struct {
	FieldName string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("record with field %q already exists", e.FieldName)
}

type InvalidPasswordError struct {
	Message string
}

func (e *InvalidPasswordError) Error() string {
	return fmt.Sprintf("invalid password: %s", e.Message)
}
