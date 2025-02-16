//nolint:wrapcheck
package account

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/converter"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/usecase"
)

type AccountHandler struct {
	accountUsecase usecase.Account
}

func NewAccountHandler(e *echo.Echo, usecase usecase.Account, m ...echo.MiddlewareFunc) *AccountHandler {
	h := &AccountHandler{accountUsecase: usecase}

	e.GET("api/info", h.getInfo, m...)

	return h
}

// (GET /api/info): получить информацию о монетах, инвентаре и истории транзакций.
func (h *AccountHandler) getInfo(c echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := ctx.Value(ctxkey.ClaimsKey).(model.Claims)
	if !ok {
		return response.SendHandlerError(c, http.StatusUnauthorized, response.ErrInvalidClaimsMessage)
	}

	info, err := h.accountUsecase.GetInfo(ctx, claims)
	if err != nil {
		return response.SendUsecaseError(c, err)
	}

	dto := converter.ConvertAccountInfoToInfoResponse(info)

	return response.SendOk(c, dto)
}
