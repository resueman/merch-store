package usecase

import "errors"

var (
	ErrProductNotFound = errors.New("product with given name not found")
)
