package response

import (
	"net/http"

	"github.com/resueman/merch-store/internal/usecase/apperrors"
)

const (
	ErrInvalidAmountMessage    = "amount must be positive"
	ErrSelfTransferMessage     = "you can't send coins to yourself"
	ErrNotEnoughBalanceMessage = "not enough balance to perform this operation"

	ErrUserNotFoundMessage    = "user not found" // ?
	ErrProductNotFoundMessage = "product not found"

	ErrInvalidPasswordMessage = "invalid password"
	ErrInvalidTokenMessage    = "invalid token"
	ErrTokenExpiredMessage    = "token expired, please re-authenticate"
	ErrGenerateTokenMessage   = "failed to generate token, please try again"

	ErrUnknownMessage = "internal server error"

	ErrBindingMessage = "failed to bind request body"
)

//nolint:errorlint
func getReturnHTTPCodeAndMessage(err error) (int, string) {
	switch err {
	case apperrors.ErrInvalidAmount:
		return http.StatusBadRequest, ErrInvalidAmountMessage
	case apperrors.ErrSelfTransfer:
		return http.StatusBadRequest, ErrSelfTransferMessage
	case apperrors.ErrNotEnoughBalance:
		return http.StatusBadRequest, ErrNotEnoughBalanceMessage
	case apperrors.ErrUserNotFound:
		return http.StatusBadRequest, ErrUserNotFoundMessage
	case apperrors.ErrProductNotFound:
		return http.StatusBadRequest, ErrProductNotFoundMessage
	case apperrors.ErrInvalidPassword:
		return http.StatusBadRequest, ErrInvalidPasswordMessage
	case apperrors.ErrInvalidToken:
		return http.StatusBadRequest, ErrInvalidTokenMessage
	case apperrors.ErrTokenExpired:
		return http.StatusBadRequest, ErrTokenExpiredMessage
	case apperrors.ErrGenerateToken:
		return http.StatusInternalServerError, ErrGenerateTokenMessage
	case apperrors.ErrInvalidClaims:
		return http.StatusBadRequest, ErrUnknownMessage
	}

	return http.StatusInternalServerError, ErrUnknownMessage
}
