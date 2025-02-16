package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/account"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/auth"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/operation"
	"github.com/resueman/merch-store/internal/delivery/middleware"
	"github.com/resueman/merch-store/internal/usecase"
)

func NewRouter(handler *echo.Echo, services *usecase.Usecase, m *middleware.AuthMiddleware) {
	handler.Use(middleware.LoggerMiddleware)

	auth.NewAuthHandler(handler, services.Auth)
	operation.NewOperationHandler(handler, services.Operation, m.AuthMiddleware)
	account.NewAccountHandler(handler, services.Account, m.AuthMiddleware)
}
