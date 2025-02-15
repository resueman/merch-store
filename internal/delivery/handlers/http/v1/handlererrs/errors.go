package handlererrs

import "errors"

var (
	ErrInvalidClaims  = errors.New("invalid claims")
	ErrBindingMessage = errors.New("invalid request body")
)
