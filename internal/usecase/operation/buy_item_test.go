package operation

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/pkg/db/postgres"
	"github.com/resueman/merch-store/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestBuyItem_BadInputError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var errUnknownGetIDByUserID = errors.New("unknown error getting account id by user id")
	var errUnknownGetProductByName = errors.New("unknown error getting product by name")

	tests := []struct {
		name     string
		claims   model.Claims
		itemName string
		mock     func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct)
		want     error
	}{
		{
			name:     "unknown error getting account id by user id",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount, _ *mocks.MockProduct) {
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(0, errUnknownGetIDByUserID)
			},
			want: errUnknownGetIDByUserID,
		},
		{
			name:     "unknown error getting product by name",
			itemName: "qwerty",
			claims:   model.Claims{UserID: 111},
			mock: func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct) {
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(1, nil)
				productRepo.EXPECT().
					GetProductByName(gomock.Any(), "qwerty").
					Return(nil, errUnknownGetProductByName)
			},
			want: errUnknownGetProductByName,
		},
		{
			name:     "product with given name doesn't exist",
			itemName: "pen",
			claims:   model.Claims{UserID: 111},
			mock: func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct) {
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(1, nil)
				productRepo.EXPECT().
					GetProductByName(gomock.Any(), "pen").
					Return(nil, repoerrors.ErrNotFound)
			},
			want: apperrors.ErrProductNotFound,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			accountRepo, productRepo := mocks.NewMockAccount(ctrl), mocks.NewMockProduct(ctrl)
			testCase.mock(accountRepo, productRepo)

			uc := NewOperationUsecase(accountRepo, nil, productRepo, nil)
			err := uc.BuyItem(context.Background(), testCase.claims, testCase.itemName)

			require.ErrorIs(t, err, testCase.want)
		})
	}
}

func purchaseWithdrawErrorMock(
	accountRepo *mocks.MockAccount,
	productRepo *mocks.MockProduct,
	txManager *mocks.MockTxManager,
	userID int,
	withdrawErr error,
) {
	accountID := 123

	accountRepo.EXPECT().
		GetIDByUserID(gomock.Any(), userID).
		Return(accountID, nil)

	product := entity.Product{ID: 120, Name: "pen", Price: 100}
	productRepo.EXPECT().
		GetProductByName(gomock.Any(), product.Name).
		Return(&product, nil)

	accountRepo.EXPECT().
		Withdraw(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(withdrawErr)

	txManager.EXPECT().
		Serializable(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
			return func() error { return f(ctx) }
		})

	txManager.EXPECT().
		WithRetry(gomock.Any()).
		DoAndReturn(func(f func() error) error {
			return f()
		})
}

func TestBuyItems_WithdrawErrorInOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var errWithdraw = errors.New("withdraw error")

	tests := []struct {
		name     string
		claims   model.Claims
		itemName string
		mock     func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct, txManager *mocks.MockTxManager)
		want     error
	}{
		{
			name:     "withdraw error: not enough balance",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct, txManager *mocks.MockTxManager) {
				purchaseWithdrawErrorMock(accountRepo, productRepo, txManager, 111, repoerrors.ErrNotEnoughBalance)
			},
			want: apperrors.ErrNotEnoughBalance,
		},
		{
			name:     "withdraw error: unknown",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount, productRepo *mocks.MockProduct, txManager *mocks.MockTxManager) {
				purchaseWithdrawErrorMock(accountRepo, productRepo, txManager, 111, errWithdraw)
			},
			want: errWithdraw,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			testCase.mock(accountRepo, productRepo, txManager)

			uc := NewOperationUsecase(accountRepo, nil, productRepo, txManager)
			err := uc.BuyItem(context.Background(), testCase.claims, testCase.itemName)

			require.ErrorIs(t, err, testCase.want)
		})
	}
}

func TestBuyItems_PurchaseErrorInOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var errPurchaseOperation = errors.New("purchase operation error")

	tests := []struct {
		name     string
		claims   model.Claims
		itemName string
		mock     func(accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			productRepo *mocks.MockProduct,
			txManager *mocks.MockTxManager,
		)
		want error
	}{
		{
			name:     "unknown error performing purchase operation",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				productRepo *mocks.MockProduct,
				txManager *mocks.MockTxManager,
			) {
				accountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(accountID, nil)

				product := entity.Product{ID: 120, Name: "pen", Price: 100}
				productRepo.EXPECT().
					GetProductByName(gomock.Any(), product.Name).
					Return(&product, nil)

				accountRepo.EXPECT().
					Withdraw(gomock.Any(), accountID, product.Price).
					Return(nil)

				operationRepo.EXPECT().
					ExecPurchaseOperation(gomock.Any(), gomock.Any()).
					Return(errPurchaseOperation)

				txManager.EXPECT().
					Serializable(gomock.Any(), db.Write, gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
						return func() error { return f(ctx) }
					})

				txManager.EXPECT().
					WithRetry(gomock.Any()).
					DoAndReturn(func(f func() error) error {
						return f()
					})
			},
			want: errPurchaseOperation,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			testCase.mock(accountRepo, operationRepo, productRepo, txManager)

			uc := NewOperationUsecase(accountRepo, operationRepo, productRepo, txManager)
			err := uc.BuyItem(context.Background(), testCase.claims, testCase.itemName)

			require.ErrorIs(t, err, testCase.want)
		})
	}
}

func TestBuyItems_TxManagerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		claims   model.Claims
		itemName string
		mock     func(accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			productRepo *mocks.MockProduct,
			txManager *mocks.MockTxManager,
		)
		want error
	}{
		{
			name:     "transaction retries exceeded",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				productRepo *mocks.MockProduct,
				txManager *mocks.MockTxManager,
			) {
				accountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(accountID, nil)

				product := entity.Product{ID: 120, Name: "pen", Price: 100}
				productRepo.EXPECT().
					GetProductByName(gomock.Any(), "pen").
					Return(&product, nil)

				txManager.EXPECT().
					Serializable(gomock.Any(), db.Write, gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
						return func() error { return f(ctx) }
					})

				txManager.EXPECT().
					WithRetry(gomock.Any()).
					Return(postgres.ErrTxRetriesExceeded)
			},
			want: postgres.ErrTxRetriesExceeded,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			testCase.mock(accountRepo, operationRepo, productRepo, txManager)

			uc := NewOperationUsecase(accountRepo, operationRepo, productRepo, txManager)
			err := uc.BuyItem(context.Background(), testCase.claims, testCase.itemName)

			require.ErrorIs(t, err, testCase.want)
		})
	}
}

func TestBuyItems_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		claims   model.Claims
		itemName string
		mock     func(accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			productRepo *mocks.MockProduct,
			txManager *mocks.MockTxManager,
		)
	}{
		{
			name:     "successful purchase",
			claims:   model.Claims{UserID: 111},
			itemName: "pen",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				productRepo *mocks.MockProduct,
				txManager *mocks.MockTxManager,
			) {
				accountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(accountID, nil)

				product := entity.Product{ID: 120, Name: "pen", Price: 100}
				productRepo.EXPECT().
					GetProductByName(gomock.Any(), "pen").
					Return(&product, nil)

				accountRepo.EXPECT().
					Withdraw(gomock.Any(), accountID, product.Price).
					Return(nil)

				operationRepo.EXPECT().
					ExecPurchaseOperation(gomock.Any(), gomock.Any()).
					Return(nil)

				txManager.EXPECT().
					Serializable(gomock.Any(), db.Write, gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
						return func() error { return f(ctx) }
					})

				txManager.EXPECT().
					WithRetry(gomock.Any()).
					DoAndReturn(func(f func() error) error {
						return f()
					})
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			testCase.mock(accountRepo, operationRepo, productRepo, txManager)

			uc := NewOperationUsecase(accountRepo, operationRepo, productRepo, txManager)
			err := uc.BuyItem(context.Background(), testCase.claims, testCase.itemName)

			require.NoError(t, err)
		})
	}
}
