package errs

import (
	"errors"
)

// Common sentinel errors.
var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("data conflict")
	ErrRateLimit             = errors.New("rate limit")
	ErrContentType           = errors.New("invalid content type")
	ErrInvalidPayload        = errors.New("invalid payload")
	ErrRequiredJSONBodyParam = errors.New("JSON body argument is required, bot not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
)

// Type just for murshallig purpose.
// Should only be used immediately before marshalling.
type JSON struct {
	Error string `json:"error"`
}
