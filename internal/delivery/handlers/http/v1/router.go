package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/middleware"
	"github.com/resueman/merch-store/internal/usecase"
)

func NewRouter(handler *echo.Echo, services *usecase.Usecase, m *middleware.AuthMiddleware) {
	handler.Use(middleware.LoggerMiddleware)

	newAuthHandler(handler, services.Auth)
	newOperationHandler(handler, services.Operation, m.AuthMiddleware)
	newAccountHandler(handler, services.Account, m.AuthMiddleware)
}
