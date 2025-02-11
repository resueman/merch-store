package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/usecase"
)

type accountHandler struct {
	usecase usecase.Account
}

func newAccountHandler(e *echo.Echo, usecase usecase.Account, m ...echo.MiddlewareFunc) *accountHandler {
	h := &accountHandler{usecase: usecase}

	e.GET("/info", h.getInfo, m...)

	return h
}

// Получить информацию о монетах, инвентаре и истории транзакций.
// (GET /api/info).
func (h *accountHandler) getInfo(c echo.Context) error {
	// из контекста достаем информацию о пользователе (какую?)

	// нужно вернуть: кол-во доступных монет,
	// купленный инвентарь (тип предмета, кол-во) -- из таблицы purchase operations,
	// инфо о полученных монетах (кол-во, от кого) -- из таблицы transfer operations,
	// инфо об отправленных монетах (кол-во, кому) -- из таблицы transfer operations.

	// 200
	// 400 -- неверный запрос, не понимаю пока, когда возвращается этот ответ
	// 401 -- нет авторизации
	// 500 -- внутренняя ошибка сервера

	return nil
}
