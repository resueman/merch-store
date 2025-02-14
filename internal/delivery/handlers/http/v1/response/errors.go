package response

import (
	"errors"
	"net/http"

	"github.com/resueman/merch-store/internal/usecase/apperrors"
)

const (
	ErrInvalidAmountMessage    = "amount must be positive"
	ErrSelfTransferMessage     = "you can't send coins to yourself"
	ErrNotEnoughBalanceMessage = "not enough balance to perform this operation"
	ErrUserNotFoundMessage     = "user not found"
	ErrProductNotFoundMessage  = "product not found"

	ErrInvalidPasswordMessage = "invalid password"
	ErrInvalidTokenMessage    = "invalid token"
	ErrTokenExpiredMessage    = "token expired, please re-authenticate"
	ErrGenerateTokenMessage   = "failed to generate token, please try again"

	ErrUnknownMessage = "internal server error"

	ErrBindingMessage = "invalid request body"
)

//nolint:errorlint
func getReturnHTTPCodeAndMessage(err error) (int, string) {
	badRequestErrors := []struct {
		err     error
		message string
	}{
		{apperrors.ErrInvalidAmount, ErrInvalidAmountMessage},
		{apperrors.ErrSelfTransfer, ErrSelfTransferMessage},
		{apperrors.ErrNotEnoughBalance, ErrNotEnoughBalanceMessage},
		{apperrors.ErrUserNotFound, ErrUserNotFoundMessage},
		{apperrors.ErrProductNotFound, ErrProductNotFoundMessage},
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e.err) {
			return http.StatusBadRequest, e.message
		}
	}

	unauthorizedErrors := []struct {
		err     error
		message string
	}{
		{apperrors.ErrInvalidPassword, ErrInvalidPasswordMessage},
		{apperrors.ErrInvalidToken, ErrInvalidTokenMessage},
		{apperrors.ErrTokenExpired, ErrTokenExpiredMessage},
		{apperrors.ErrGenerateToken, ErrGenerateTokenMessage},
		{apperrors.ErrInvalidClaims, ErrUnknownMessage},
	}

	for _, e := range unauthorizedErrors {
		if errors.Is(err, e.err) {
			return http.StatusUnauthorized, e.message
		}
	}

	return http.StatusInternalServerError, ErrUnknownMessage
}
