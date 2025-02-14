//nolint:wrapcheck
package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/converter"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/usecase"
)

type accountHandler struct {
	accountUsecase usecase.Account
}

func newAccountHandler(e *echo.Echo, usecase usecase.Account, m ...echo.MiddlewareFunc) *accountHandler {
	h := &accountHandler{accountUsecase: usecase}

	e.GET("api/info", h.getInfo, m...)

	return h
}

// (GET /api/info): получить информацию о монетах, инвентаре и истории транзакций.
func (h *accountHandler) getInfo(ctx echo.Context) error {
	info, err := h.accountUsecase.GetInfo(ctx.Request().Context())
	if err != nil {
		return response.SendUsecaseError(ctx, err)
	}

	dto := converter.ConvertAccountInfoToInfoResponse(info)

	return response.SendOk(ctx, dto)
}
