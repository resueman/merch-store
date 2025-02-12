package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/usecase"
)

type accountHandler struct {
	accountUsecase usecase.Account
}

func newAccountHandler(e *echo.Echo, usecase usecase.Account, m ...echo.MiddlewareFunc) *accountHandler {
	h := &accountHandler{accountUsecase: usecase}

	e.GET("/info", h.getInfo, m...)

	return h
}

// (GET /api/info): получить информацию о монетах, инвентаре и истории транзакций.
func (h *accountHandler) getInfo(ctx echo.Context) error {
	info, err := h.accountUsecase.GetInfo(ctx.Request().Context())
	if err != nil {
		return sendUsecaseErrorResponse(ctx, err)
	}

	return sendOkResponse(ctx, info)
}
