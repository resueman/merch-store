//nolint:wrapcheck
package v1

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	dto "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/model"
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
func (h *operationHandler) buyItem(c echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := ctx.Value(ctxkey.ClaimsKey).(model.Claims)
	if !ok {
		return response.SendHandlerError(c, http.StatusUnauthorized, response.ErrInvalidClaimsMessage)
	}

	item := c.Param("item")
	if item == "" {
		return response.SendHandlerError(c, http.StatusBadRequest, "item name is required")
	}

	if err := h.operationUsecase.BuyItem(ctx, claims, item); err != nil {
		return response.SendUsecaseError(c, err)
	}

	return response.SendNoContent(c)
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
func (h *operationHandler) sendCoin(c echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := ctx.Value(ctxkey.ClaimsKey).(model.Claims)
	if !ok {
		return response.SendHandlerError(c, http.StatusUnauthorized, response.ErrInvalidClaimsMessage)
	}

	var input dto.SendCoinRequest
	if err := c.Bind(&input); err != nil {
		return response.SendHandlerError(c, http.StatusBadRequest, response.ErrBindingMessage)
	}

	if errMsg := h.validateSendCoinRequest(&input); errMsg != "" {
		return response.SendHandlerError(c, http.StatusBadRequest, errMsg)
	}

	err := h.operationUsecase.SendCoin(ctx, claims, input.ToUser, input.Amount)
	if err != nil {
		return response.SendUsecaseError(c, err)
	}

	return response.SendNoContent(c)
}
