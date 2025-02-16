package operation

import (
	"context"
	"testing"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOperationUsecase struct {
	mock.Mock
}

func (m *MockOperationUsecase) BuyItem(ctx context.Context, claims model.Claims, item string) error {
	args := m.Called(ctx, claims, item)
	return args.Error(0)
}

func (m *MockOperationUsecase) SendCoin(ctx context.Context, claims model.Claims, toUser string, amount int) error {
	args := m.Called(ctx, claims, toUser, amount)
	return args.Error(0)
}

func TestNewOperationHandler(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockOperationUsecase)
	handler := NewOperationHandler(e, mockUsecase)

	assert.NotNil(t, handler)
}
