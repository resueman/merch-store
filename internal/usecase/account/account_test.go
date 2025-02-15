package account

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/test/mocks"
	"github.com/stretchr/testify/require"
)

type repoInfo struct {
	accountID         int
	balance           int
	purchases         []entity.Purchase
	incomingTransfers []entity.Transfer
	outgoingTransfers []entity.Transfer
}

type repoInfoError struct {
	gettingAccountErr    error
	balanceErr           error
	purchasesErr         error
	incomingTransfersErr error
	outgoingTransfersErr error
}

func getRepoInfoMock(
	accountRepo *mocks.MockAccount,
	operationRepo *mocks.MockOperation,
	claims model.Claims,
	repoData *repoInfo,
) {
	accountID := repoData.accountID
	accountRepo.EXPECT().
		GetIDByUserID(gomock.Any(), claims.UserID).
		Return(accountID, nil)

	accountRepo.EXPECT().
		GetBalanceByAccountID(gomock.Any(), accountID).
		Return(repoData.balance, nil)

	accountRepo.EXPECT().
		GetPurchasesByAccountID(gomock.Any(), accountID).
		Return(repoData.purchases, nil)

	operationRepo.EXPECT().
		GetIncomingTransfers(gomock.Any(), accountID).
		Return(repoData.incomingTransfers, nil)

	operationRepo.EXPECT().
		GetOutgoingTransfers(gomock.Any(), accountID).
		Return(repoData.outgoingTransfers, nil)
}

func getRepoInfoWithErrorMock(
	accountRepo *mocks.MockAccount,
	operationRepo *mocks.MockOperation,
	claims model.Claims,
	repoData *repoInfoError,
) {
	accountID := 123
	accountRepo.EXPECT().
		GetIDByUserID(gomock.Any(), claims.UserID).
		Return(accountID, repoData.gettingAccountErr)

	if repoData.gettingAccountErr != nil {
		return
	}

	accountRepo.EXPECT().
		GetBalanceByAccountID(gomock.Any(), accountID).
		Return(100, repoData.balanceErr)

	if repoData.balanceErr != nil {
		return
	}

	accountRepo.EXPECT().
		GetPurchasesByAccountID(gomock.Any(), accountID).
		Return([]entity.Purchase{}, repoData.purchasesErr)

	if repoData.purchasesErr != nil {
		return
	}

	operationRepo.EXPECT().
		GetIncomingTransfers(gomock.Any(), accountID).
		Return([]entity.Transfer{}, repoData.incomingTransfersErr)

	if repoData.incomingTransfersErr != nil {
		return
	}

	operationRepo.EXPECT().
		GetOutgoingTransfers(gomock.Any(), accountID).
		Return([]entity.Transfer{}, repoData.outgoingTransfersErr)
}

func txManagerMock(txManager *mocks.MockTxManager) {
	txManager.EXPECT().
		ReadCommitted(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
			return func() error { return f(ctx) }
		})

	txManager.EXPECT().
		WithRetry(gomock.Any()).
		DoAndReturn(func(f func() error) error {
			return f()
		})
}

func TestGetInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			repoData *repoInfo,
		)
		in   *repoInfo
		want *model.AccountInfo
	}{
		{
			name: "success with non-empty inventory, incoming and outgoing transfers",
			mock: func(
				accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				repoData *repoInfo,
			) {
				getRepoInfoMock(accountRepo, operationRepo, claims, repoData)
				txManagerMock(txManager)
			},
			in: &repoInfo{
				balance: 300,
				purchases: []entity.Purchase{
					{Name: "pen", Quantity: 1},
					{Name: "book", Quantity: 2},
				},
				incomingTransfers: []entity.Transfer{
					{Amount: 10, SenderUsername: "B", RecipientUsername: "A"},
					{Amount: 20, SenderUsername: "C", RecipientUsername: "A"},
					{Amount: 70, SenderUsername: "D", RecipientUsername: "A"},
				},
				outgoingTransfers: []entity.Transfer{
					{Amount: 50, SenderUsername: "A", RecipientUsername: "E"},
					{Amount: 200, SenderUsername: "A", RecipientUsername: "F"},
					{Amount: 250, SenderUsername: "A", RecipientUsername: "G"},
				},
			},
			want: &model.AccountInfo{
				Balance: 300,
				Inventory: []model.Inventory{
					{Name: "pen", Quantity: 1},
					{Name: "book", Quantity: 2},
				},
				IncomingTransfers: []model.IncomingTransfer{
					{Amount: 10, SenderUsername: "B"},
					{Amount: 20, SenderUsername: "C"},
					{Amount: 70, SenderUsername: "D"},
				},
				OutgoingTransfers: []model.OutgoingTransfer{
					{Amount: 50, RecipientUsername: "E"},
					{Amount: 200, RecipientUsername: "F"},
					{Amount: 250, RecipientUsername: "G"},
				},
			},
		},
		{
			name: "success with empty inventory, incoming and outgoing transfers",
			mock: func(
				accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				in *repoInfo,
			) {
				getRepoInfoMock(accountRepo, operationRepo, claims, in)
				txManagerMock(txManager)
			},
			in: &repoInfo{
				balance:           100,
				purchases:         []entity.Purchase{},
				incomingTransfers: []entity.Transfer{},
				outgoingTransfers: []entity.Transfer{},
			},
			want: &model.AccountInfo{
				Balance:           100,
				Inventory:         []model.Inventory{},
				IncomingTransfers: []model.IncomingTransfer{},
				OutgoingTransfers: []model.OutgoingTransfer{},
			},
		},
	}

	claims := model.Claims{
		UserID: 111,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			tt.mock(accountRepo, operationRepo, txManager, claims, tt.in)

			accountUsecase := NewAccountUsecase(accountRepo, operationRepo, productRepo, txManager)

			actual, err := accountUsecase.GetInfo(context.Background(), claims)

			require.NoError(t, err)

			if actual == nil {
				t.Errorf("expected non-nil struct, got nil")
			} else {
				require.True(t, reflect.DeepEqual(actual, tt.want))
			}
		})
	}
}

func TestGetInfo_Error_NoAccountForUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	unknownErrGettingAccountIDByUserID := errors.New("unknown error getting account id by user id")

	tests := []struct {
		name    string
		mock    func(accountRepo *mocks.MockAccount, claims model.Claims)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "unknown error getting sender's account id",
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims) {
				repoInfoError := &repoInfoError{gettingAccountErr: unknownErrGettingAccountIDByUserID}
				getRepoInfoWithErrorMock(accountRepo, nil, claims, repoInfoError)
			},
			want:    nil,
			wantErr: unknownErrGettingAccountIDByUserID,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			tt.mock(accountRepo, claims)

			accountUsecase := NewAccountUsecase(accountRepo, nil, nil, nil)

			info, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, info)
		})
	}
}

func TestGetInfo_Error_ErrorGettingBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	unknownErrGettingBalance := errors.New("error")

	tests := []struct {
		name    string
		mock    func(accountRepo *mocks.MockAccount, txManager *mocks.MockTxManager, claims model.Claims)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "unknown error getting balance",
			mock: func(accountRepo *mocks.MockAccount, txManager *mocks.MockTxManager, claims model.Claims) {
				repoInfoError := &repoInfoError{balanceErr: unknownErrGettingBalance}
				getRepoInfoWithErrorMock(accountRepo, nil, claims, repoInfoError)

				txManagerMock(txManager)
			},
			want:    nil,
			wantErr: unknownErrGettingBalance,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			tt.mock(accountRepo, txManager, claims)

			accountUsecase := NewAccountUsecase(accountRepo, nil, nil, txManager)
			info, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, info)
		})
	}
}

func TestGetInfo_Error_ErrorGettingPurchases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	unknownErrGettingPurchases := errors.New("error")

	tests := []struct {
		name    string
		mock    func(accountRepo *mocks.MockAccount, txManager *mocks.MockTxManager, claims model.Claims)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "unknown error getting purchases",
			mock: func(accountRepo *mocks.MockAccount, txManager *mocks.MockTxManager, claims model.Claims) {
				repoInfoError := &repoInfoError{purchasesErr: unknownErrGettingPurchases}
				getRepoInfoWithErrorMock(accountRepo, nil, claims, repoInfoError)

				txManagerMock(txManager)
			},
			want:    nil,
			wantErr: unknownErrGettingPurchases,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			tt.mock(accountRepo, txManager, claims)

			accountUsecase := NewAccountUsecase(accountRepo, nil, nil, txManager)
			info, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, info)
		})
	}
}

func TestGetInfo_Error_ErrorGettingIncomingTransfers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	unknownErrGettingIncomingTransfers := errors.New("error")

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
		)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "unknown error getting incoming transfers",
			mock: func(
				accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
			) {
				repoInfoError := &repoInfoError{incomingTransfersErr: unknownErrGettingIncomingTransfers}
				getRepoInfoWithErrorMock(accountRepo, operationRepo, claims, repoInfoError)

				txManagerMock(txManager)
			},
			want:    nil,
			wantErr: unknownErrGettingIncomingTransfers,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			tt.mock(accountRepo, operationRepo, txManager, claims)

			accountUsecase := NewAccountUsecase(accountRepo, operationRepo, nil, txManager)
			info, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, info)
		})
	}
}

func TestGetInfo_Error_ErrorGettingOutgoingTransfers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	unknownErrGettingOutgoingTransfers := errors.New("error")

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
		)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "unknown error getting outgoing transfers",
			mock: func(
				accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
			) {
				repoInfoError := &repoInfoError{outgoingTransfersErr: unknownErrGettingOutgoingTransfers}
				getRepoInfoWithErrorMock(accountRepo, operationRepo, claims, repoInfoError)

				txManagerMock(txManager)
			},
			want:    nil,
			wantErr: unknownErrGettingOutgoingTransfers,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)

			tt.mock(accountRepo, operationRepo, txManager, claims)

			accountUsecase := NewAccountUsecase(accountRepo, operationRepo, nil, txManager)
			info, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, info)
		})
	}
}

func TestGetInfo_Error_TxManagerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txManagerError := errors.New("error")
	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
		)
		want    *model.AccountInfo
		wantErr error
	}{
		{
			name: "tx manager error",
			mock: func(
				accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
			) {
				repoInfoError := &repoInfoError{outgoingTransfersErr: txManagerError}
				getRepoInfoWithErrorMock(accountRepo, operationRepo, claims, repoInfoError)

				txManagerMock(txManager)
			},
			want:    nil,
			wantErr: txManagerError,
		},
	}

	claims := model.Claims{UserID: 111}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			productRepo := mocks.NewMockProduct(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, operationRepo, txManager, claims)

			accountUsecase := NewAccountUsecase(accountRepo, operationRepo, productRepo, txManager)

			actual, err := accountUsecase.GetInfo(context.Background(), claims)

			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, actual)
		})
	}
}
