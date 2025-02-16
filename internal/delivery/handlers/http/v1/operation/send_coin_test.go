package operation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendCoin_Success(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	mockUsecase.On("SendCoin", mock.Anything, claims, "B", 100).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(`{"toUser":"B","amount":100}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSendCoin_ErrorUnauthorized(t *testing.T) {
	e := echo.New()
	handler := NewOperationHandler(e, new(MockOperationUsecase))

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSendCoin_ErrorBadRequestNoToUser(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(`{"toUser":"","amount":100}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendCoin_ErrorBadRequestIncorrectAmount(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(`{"toUser":"B","amount":0}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendCoin_BadClaims(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := struct {
		UserID int
	}{
		UserID: 123,
	}

	mockUsecase.On("SendCoin", mock.Anything, claims, "B", 100).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSendCoin_ErrorBinding(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(`{"invalidJson}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendCoin_ErrorUsecase(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	mockUsecase.On("SendCoin", mock.Anything, claims, "user2", 100).Return(errors.New("error"))

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(`{"toUser":"user2","amount":100}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.sendCoin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
