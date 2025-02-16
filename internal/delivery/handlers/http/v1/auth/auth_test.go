package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/response"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateToken(ctx context.Context, input model.AuthRequestInput) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ParseToken(ctx context.Context, token string) (model.Claims, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(model.Claims), args.Error(1)
}

func TestAuth(t *testing.T) {
	e := echo.New()
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(e, mockAuthService)

	t.Run("Successful authentication", func(t *testing.T) {
		mockAuthService.
			On("GenerateToken", mock.Anything, model.AuthRequestInput{Username: "user", Password: "pass"}).
			Return("token", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(`{"username":"user","password":"pass"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, handler.auth(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "token")
		}
	})

	t.Run("Missing username", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(`{"password":"testpass"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, handler.auth(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "username is required;")
		}
	})

	t.Run("Missing password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(`{"username":"testuser"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, handler.auth(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "password is required;")
		}
	})

	t.Run("Binding error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth", strings.NewReader(`{invalid json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		if assert.NoError(t, handler.auth(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), response.ErrBindingMessage)
		}
	})
}
