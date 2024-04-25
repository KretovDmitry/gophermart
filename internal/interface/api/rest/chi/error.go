package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
)

func checkJSONDecodeError(err error) error {
	var e *json.UnmarshalTypeError
	if errors.As(err, &e) {
		return fmt.Errorf("%w: %s must be of type %s, got %s",
			errs.ErrInvalidRequest, e.Field, e.Type, e.Value)
	}
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: empty body", errs.ErrInvalidRequest)
	}

	return err
}
