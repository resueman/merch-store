package middleware

import (
	"github.com/labstack/echo"
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return next(ctx)
	}
}
