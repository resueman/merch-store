package middleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		log.Info(ctx.Request().URL.String())
		return next(ctx)
	}
}
