package apperrors

import "errors"

var (
	ErrInvalidAmount    = errors.New("amount must be positive")
	ErrSelfTransfer     = errors.New("self transfer")
	ErrNotEnoughBalance = errors.New("not enough balance")

	ErrUserNotFound    = errors.New("user with given token not found") // ?
	ErrProductNotFound = errors.New("product not found")

	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrGenerateToken   = errors.New("failed to generate token")
)
