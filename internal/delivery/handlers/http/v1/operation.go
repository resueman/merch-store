package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/usecase"
)

type operationHandler struct {
	usecase usecase.Operation
}

func newOperationHandler(e *echo.Echo, usecase usecase.Operation, m ...echo.MiddlewareFunc) *operationHandler {
	h := &operationHandler{usecase: usecase}

	e.GET("/buy/:item", h.buyItem, m...)
	e.POST("/sendCoin", h.sendCoin, m...)

	return h
}

// Купить предмет за монеты.
// (GET /api/buy/{item})
func (h *operationHandler) buyItem(c echo.Context) error { // GetApiBuyItem(ctx echo.Context, item string) error

	// достаем данные о пользователе из контекста

	// достаем данные о предмете из запроса

	// обращаемся к usecase, чтобы купить предмет user or account id + item id

	// 200
	// 400 -- неверный запрос, например нет такого предмета
	// 401
	// 500

	return nil
}

// Отправить монеты другому пользователю.
// (POST /api/sendCoin)
func (h *operationHandler) sendCoin(c echo.Context) error {
	// достаем данные об отправителе из контекста
	// из тела запроса достаем данные о получателе и кол-ве монет
	// валидируем, что в теле запроса есть toUser и amount

	// обращаемся к usecase, чтобы отправить монеты
	// sender id (user or account id) + receiver id (user or account id) + amount

	// 200
	// 400 -- неверный запрос, например не прошла валидация
	// 401
	// 500

	return nil
}
