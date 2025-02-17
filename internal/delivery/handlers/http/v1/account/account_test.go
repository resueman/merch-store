package account

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	v1 "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/converter"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountUsecase struct {
	mock.Mock
}

func (m *MockAccountUsecase) GetInfo(ctx context.Context, claims model.Claims) (*model.AccountInfo, error) {
	args := m.Called(ctx, claims)
	info, err := args.Get(0).(model.AccountInfo), args.Error(1)
	return &info, err
}

func TestGetInfo(t *testing.T) {
	e := echo.New()
	mockUsecase := &MockAccountUsecase{}
	handler := NewAccountHandler(e, mockUsecase)

	t.Run("successful get info", func(t *testing.T) {
		claims := model.Claims{UserID: 123}
		info := model.AccountInfo{
			Balance: 100,
			Inventory: []model.Inventory{
				{Name: "pen", Quantity: 10},
				{Name: "book", Quantity: 5},
			},
			OutgoingTransfers: []model.OutgoingTransfer{
				{Amount: 100, RecipientUsername: "B"},
				{Amount: 200, RecipientUsername: "C"},
			},
			IncomingTransfers: []model.IncomingTransfer{
				{Amount: 100, SenderUsername: "D"},
				{Amount: 200, SenderUsername: "E"},
			},
		}

		mockUsecase.On("GetInfo", mock.Anything, claims).Return(info, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ctx := context.WithValue(c.Request().Context(), ctxkey.ClaimsKey, claims)
		c.SetRequest(c.Request().WithContext(ctx))

		err := handler.GetInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response v1.InfoResponse
		err = json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, converter.ConvertAccountInfoToInfoResponse(&info), response)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.GetInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
