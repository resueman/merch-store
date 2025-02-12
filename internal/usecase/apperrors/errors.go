package apperrors

import "errors"

var (
	ErrInvalidAmount    = errors.New("amount must be positive")
	ErrSelfTransfer     = errors.New("self transfer")
	ErrNotEnoughBalance = errors.New("not enough balance")
	ErrProductNotFound  = errors.New("product not found")
	ErrUserNotFound     = errors.New("user with given token not found") // ?
)
