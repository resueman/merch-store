//nolint:wrapcheck
package v1

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	dto "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/usecase"
)

type operationHandler struct {
	operationUsecase usecase.Operation
}

func newOperationHandler(e *echo.Echo, usecase usecase.Operation, m ...echo.MiddlewareFunc) *operationHandler {
	h := &operationHandler{operationUsecase: usecase}

	e.GET("api/buy/:item", h.buyItem, m...)
	e.POST("api/sendCoin", h.sendCoin, m...)

	return h
}

// (GET /api/buy/{item}): купить предмет за монеты.
func (h *operationHandler) buyItem(ctx echo.Context) error {
	item := ctx.Param("item")
	if item == "" {
		return response.SendHandlerError(ctx, http.StatusBadRequest, "item name is required")
	}

	if err := h.operationUsecase.BuyItem(ctx.Request().Context(), item); err != nil {
		return response.SendUsecaseError(ctx, err)
	}

	return response.SendNoContent(ctx)
}

func (h *operationHandler) validateSendCoinRequest(input *dto.SendCoinRequest) string {
	var errMsg strings.Builder
	if input.Amount <= 0 {
		errMsg.WriteString("amount must be positive;")
	}

	if input.ToUser == "" {
		errMsg.WriteString("toUser is required;")
	}

	return errMsg.String()
}

// (POST /api/sendCoin): отправить монеты другому пользователю.
func (h *operationHandler) sendCoin(ctx echo.Context) error {
	var input dto.SendCoinRequest
	if err := ctx.Bind(&input); err != nil {
		return response.SendHandlerError(ctx, http.StatusBadRequest, response.ErrBindingMessage)
	}

	if errMsg := h.validateSendCoinRequest(&input); errMsg != "" {
		return response.SendHandlerError(ctx, http.StatusBadRequest, errMsg)
	}

	err := h.operationUsecase.SendCoin(ctx.Request().Context(), input.ToUser, input.Amount)
	if err != nil {
		return response.SendUsecaseError(ctx, err)
	}

	return response.SendNoContent(ctx)
}
