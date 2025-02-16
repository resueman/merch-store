//nolint:wrapcheck
package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	dto "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/usecase"
)

type authHandler struct {
	authService usecase.Auth
}

func NewAuthHandler(e *echo.Echo, authService usecase.Auth) *authHandler {
	h := &authHandler{authService: authService}

	e.POST("/api/auth", h.Auth)

	return h
}

func (h *authHandler) validateAuthRequest(input *dto.AuthRequest) string {
	var errMsg strings.Builder
	if input.Username == "" {
		errMsg.WriteString("username is required;")
	}

	if input.Password == "" {
		errMsg.WriteString("password is required;")
	}

	return errMsg.String()
}

// (POST /api/auth): аутентификация и получение JWT-токена.
// При первой аутентификации пользователь создается автоматически.
func (h *authHandler) Auth(ctx echo.Context) error {
	var input dto.AuthRequest
	if err := ctx.Bind(&input); err != nil {
		return response.SendHandlerError(ctx, http.StatusBadRequest, response.ErrBindingMessage)
	}

	if errMsg := h.validateAuthRequest(&input); errMsg != "" {
		return response.SendHandlerError(ctx, http.StatusBadRequest, errMsg)
	}

	authInput := model.AuthRequestInput{Username: input.Username, Password: input.Password}

	token, err := h.authService.GenerateToken(ctx.Request().Context(), authInput)
	if err != nil {
		return response.SendUsecaseError(ctx, err)
	}

	dto := dto.AuthResponse{Token: &token}

	return response.SendOk(ctx, dto)
}
