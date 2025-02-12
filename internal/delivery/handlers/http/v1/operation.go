package v1

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/usecase"
)

// 1. сделать валидацию с возвратом кастомных сообщений.

type operationHandler struct {
	operationUsecase usecase.Operation
}

func newOperationHandler(e *echo.Echo, usecase usecase.Operation, m ...echo.MiddlewareFunc) *operationHandler {
	h := &operationHandler{operationUsecase: usecase}

	e.GET("/buy/:item", h.buyItem, m...)
	e.POST("/sendCoin", h.sendCoin, m...)

	return h
}

type buyItemRequest struct {
	Item string `param:"item" validate:"required"`
}

// (GET /api/buy/{item}): купить предмет за монеты.
func (h *operationHandler) buyItem(ctx echo.Context) error { // GetApiBuyItem(ctx echo.Context, item string) error
	item := ctx.Param("item")
	if item == "" {
		return sendHandlerErrorResponse(ctx, http.StatusBadRequest, "item name is required") // можно через bind + validate
	}

	/* либо
	var input buyItemRequest
	if err := ctx.Bind(&input); err != nil {
		return h.sendHandlerErrorResponse(ctx, http.StatusBadRequest, err.Error())
	}
	if err := h.validate.Struct(input); err != nil {
		return h.sendHandlerErrorResponse(ctx, http.StatusBadRequest, err.Error())
	}*/

	if err := h.operationUsecase.BuyItem(ctx.Request().Context(), item); err != nil {
		return sendUsecaseErrorResponse(ctx, err)
	}

	return sendNoContentResponse(ctx)
}

type sendCoinRequest struct {
	Amount int    `json:"amount" validate:"required"`
	ToUser string `json:"toUser" validate:"required"`
}

// (POST /api/sendCoin): отправить монеты другому пользователю.
func (h *operationHandler) sendCoin(ctx echo.Context) error {
	var input sendCoinRequest
	if err := ctx.Bind(&input); err != nil {
		return sendHandlerErrorResponse(ctx, http.StatusBadRequest, err.Error())
	}

	/*if err := h.validate.Struct(input); err != nil {
		return h.sendHandlerErrorResponse(ctx, http.StatusBadRequest, err.Error())
	}*/

	err := h.operationUsecase.SendCoin(ctx.Request().Context(), input.ToUser, input.Amount)
	if err != nil {
		return sendUsecaseErrorResponse(ctx, err)
	}

	return sendNoContentResponse(ctx)
}
