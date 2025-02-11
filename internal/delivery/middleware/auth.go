package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authHeader := ctx.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header missing")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
		}

		token := parts[1]
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token is empty")
		}

		// вшиваем в контекст данные из токена
		// тут работа с JWT, нужен наш secret key
		// надо будет зашить в контекст инфу о пользователе

		return next(ctx)
	}
}
