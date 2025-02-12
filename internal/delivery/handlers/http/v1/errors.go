package v1

import (
	"net/http"

	"github.com/resueman/merch-store/internal/usecase/apperrors"
)

const (
	ErrInvalidAmountMessage    = "amount must be positive"
	ErrSelfTransferMessage     = "you can't send coins to yourself"
	ErrNotEnoughBalanceMessage = "not enough balance to perform this operation"
	ErrUserNotFoundMessage     = "user not found" // ?
	ErrProductNotFoundMessage  = "product not found"
	ErrUnknownMessage          = "internal server error"
	ErrBindingMessage          = "failed to bind request body"
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
		return http.StatusNotFound, ErrUserNotFoundMessage
	case apperrors.ErrProductNotFound:
		return http.StatusNotFound, ErrProductNotFoundMessage
	}

	return http.StatusInternalServerError, ErrUnknownMessage
}
