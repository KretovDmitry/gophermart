package errs

import (
	"errors"
)

// Common sentinel errors.
var (
	ErrNotFound           = errors.New("not found")
	ErrRateLimit          = errors.New("rate limit")
	ErrDataConflict       = errors.New("data conflict")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrRequiredBodyParam  = errors.New("body argument is required, bot not found")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

// Type just for murshallig purpose.
// Should only be used immediately before marshalling.
type JSON struct {
	Error string `json:"error"`
}
