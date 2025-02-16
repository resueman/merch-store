package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) ParseToken(ctx context.Context, tokenString string) (model.Claims, error) {
	args := m.Called(ctx, tokenString)
	return args.Get(0).(model.Claims), args.Error(1)
}

func (m *MockAuthUsecase) GenerateToken(ctx context.Context, claims model.AuthRequestInput) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	e := echo.New()
	mockUsecase := &MockAuthUsecase{}
	authMiddleware := NewAuthMiddleware(mockUsecase)

	t.Run("Missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := authMiddleware.AuthMiddleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Authorization header missing")
	})

	t.Run("Invalid authorization header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "InvalidHeader")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := authMiddleware.AuthMiddleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid authorization header format")
	})

	t.Run("Empty token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := authMiddleware.AuthMiddleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Token is empty")
	})

	t.Run("Valid token", func(t *testing.T) {
		claims := model.Claims{UserID: 123}
		mockUsecase.On("ParseToken", mock.Anything, "valid_token").Return(claims, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := authMiddleware.AuthMiddleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		assert.Equal(t, claims, c.Request().Context().Value(ctxkey.ClaimsKey))
	})

	t.Run("Error parsing token", func(t *testing.T) {
		mockUsecase.On("ParseToken", mock.Anything, "invalid_token").Return(model.Claims{}, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := authMiddleware.AuthMiddleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
