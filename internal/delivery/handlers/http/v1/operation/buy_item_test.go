package operation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuyItem_Success(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	mockUsecase.On("BuyItem", mock.Anything, claims, "pen").Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/buy/pen", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("item")
	c.SetParamValues("pen")

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.BuyItem(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBuyItem_ErrorUnauthorized(t *testing.T) {
	e := echo.New()
	handler := NewOperationHandler(e, new(MockOperationUsecase))

	req := httptest.NewRequest(http.MethodGet, "/api/buy/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.BuyItem(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestBuyItem_BadClaims(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := struct {
		UserID int
	}{
		UserID: 123,
	}

	mockUsecase.On("BuyItem", mock.Anything, claims, "pen").Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/buy/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.BuyItem(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestBuyItem_ErrorBadRequest(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	req := httptest.NewRequest(http.MethodGet, "/api/buy/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.BuyItem(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBuyItem_ErrorUsecase(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	claims := model.Claims{UserID: 123}
	mockUsecase.On("BuyItem", mock.Anything, claims, "pen").Return(errors.New("some error"))

	req := httptest.NewRequest(http.MethodGet, "/api/buy/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("item")
	c.SetParamValues("pen")

	ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
	c.SetRequest(c.Request().WithContext(ctx))

	err := handler.BuyItem(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
