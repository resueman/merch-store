package repoerrors

import "errors"

var (
	ErrNotEnoughBalance = errors.New("not enough balance")
	ErrNotFound         = errors.New("not found")
)
