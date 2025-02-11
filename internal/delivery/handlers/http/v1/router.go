package v1

import (
	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/middleware"
	"github.com/resueman/merch-store/internal/usecase"
)

func NewRouter(handler *echo.Echo, services *usecase.Usecase) {
	handler.Use(middleware.LoggerMiddleware)

	newAuthHandler(handler, services.User)
	newOperationHandler(handler, services.Operation, middleware.AuthMiddleware)
	newAccountHandler(handler, services.Account, middleware.AuthMiddleware)
}
