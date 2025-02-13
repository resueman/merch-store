package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/usecase"
)

type AuthMiddleware struct {
	authUsecase usecase.Auth
}

func NewAuthMiddleware(authUsecase usecase.Auth) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: authUsecase}
}

func (m *AuthMiddleware) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authHeader := ctx.Request().Header.Get("Authorization")
		if authHeader == "" {
			return response.SendHandlerError(ctx, http.StatusUnauthorized, "Authorization header missing")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.SendHandlerError(ctx, http.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		if tokenString == "" {
			return response.SendHandlerError(ctx, http.StatusUnauthorized, "Token is empty")
		}

		claims, err := m.authUsecase.ParseToken(ctx.Request().Context(), tokenString)
		if err != nil {
			return response.SendHandlerError(ctx, http.StatusUnauthorized, "Invalid token")
		}

		ctx.Set(string(ctxkey.ClaimsKey), claims)

		return next(ctx)
	}
}
