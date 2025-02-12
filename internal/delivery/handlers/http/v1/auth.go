package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/usecase"
)

type authHandler struct {
	userService usecase.User
}

func newAuthHandler(e *echo.Echo, userService usecase.User) *authHandler {
	h := &authHandler{userService: userService}

	e.POST("/auth", h.auth)

	return h
}

type authRequest struct {
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required"`
}

// Аутентификация и получение JWT-токена. При первой аутентификации пользователь создается автоматически.
// (POST /api/auth).
func (h *authHandler) auth(c echo.Context) error {
	// получаем username и password из тела запроса
	// валидация, что оба параметра присутствуют, иначе 400

	// проверяем, существует ли пользователь с таким username
	// если нет, то делаем sign up и выдаем токен

	// если пользователь существует, то проверяем password
	// если password неверный, то возвращаем 401
	// если password верный, то выдаем JWT-токен

	// 200 -- AuthResponse с JWT-токеном
	// 400 -- неверные данные, нет username или password
	// 401 -- предполагаю, что неверный пароль
	// 500 -- внутренняя ошибка сервера, например проблемы с БД, этот ответ как раз и будет снижать SLI успешности

	return nil
}
