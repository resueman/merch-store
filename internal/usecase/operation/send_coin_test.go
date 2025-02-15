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
	"github.com/resueman/merch-store/test/mocks"
	"github.com/stretchr/testify/require"
)

type SendCoinInput struct {
	ReceiverID string
	Amount     int
}

func TestSendCoin_BadInputError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var unknownErrGettingAccountIDByUserID = errors.New("unknown error getting account id by user id")
	var unknownErrGettingReceiverAccountID = errors.New("unknown error getting account id by username")

	tests := []struct {
		name   string
		amount int
		mock   func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string)
		want   error
	}{
		{
			name:   "transfer 0 coins",
			amount: 0,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
			},
			want: apperrors.ErrInvalidAmount,
		},
		{
			name:   "transfer negative amount",
			amount: -100,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
			},
			want: apperrors.ErrInvalidAmount,
		},
		{
			name:   "unknown error getting sender's account id",
			amount: 100,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(0, unknownErrGettingAccountIDByUserID)
			},
			want: unknownErrGettingAccountIDByUserID,
		},
		{
			name:   "unknown error getting receiver's account id",
			amount: 100,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
				senderAccountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(senderAccountID, nil)

				accountRepo.EXPECT().
					GetIDByUsername(gomock.Any(), receiverUsername).
					Return(0, unknownErrGettingReceiverAccountID)
			},
			want: unknownErrGettingReceiverAccountID,
		},
		{
			name:   "sender doesn't exist",
			amount: 100,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
				senderAccountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), claims.UserID).
					Return(senderAccountID, nil)

				accountRepo.EXPECT().
					GetIDByUsername(gomock.Any(), receiverUsername).
					Return(0, repoerrors.ErrNotFound)
			},
			want: apperrors.ErrUserNotFound,
		},
		{
			name:   "self transfer",
			amount: 100,
			mock: func(accountRepo *mocks.MockAccount, claims model.Claims, receiverUsername string) {
				senderAccountID := 123
				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), 111).
					Return(senderAccountID, nil)

				accountRepo.EXPECT().
					GetIDByUsername(gomock.Any(), receiverUsername).
					Return(senderAccountID, nil)
			},
			want: apperrors.ErrSelfTransfer,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			testCase.mock(accountRepo, claims, receiverUsername)

			uc := NewOperationUsecase(accountRepo, nil, nil, nil)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, testCase.amount)

			require.ErrorIs(t, err, testCase.want)
		})
	}
}

func transferWithdrawErrorMock(
	accountRepo *mocks.MockAccount,
	txManager *mocks.MockTxManager,
	claims model.Claims,
	receiverUsername string,
	amount int,
	withdrawErr error,
) {
	senderAccountID, receiverAccountID := 123, 456

	accountRepo.EXPECT().
		GetIDByUserID(gomock.Any(), claims.UserID).
		Return(senderAccountID, nil)

	accountRepo.EXPECT().
		GetIDByUsername(gomock.Any(), receiverUsername).
		Return(receiverAccountID, nil)

	accountRepo.EXPECT().
		Withdraw(gomock.Any(), senderAccountID, amount).
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

func TestSendCoin_WithdrawError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var accountRepoUnknownWithdrawError = errors.New("unknown withdraw error from account repo")

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			receiverUsername string,
			amount int,
			repoWithdrawErr error,
		)
		returnedError error
		want          error
	}{
		{
			name: "not enough balance",
			mock: func(accountRepo *mocks.MockAccount,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
				repoWithdrawErr error,
			) {
				transferWithdrawErrorMock(accountRepo,
					txManager,
					claims,
					receiverUsername,
					amount,
					repoWithdrawErr)
			},
			returnedError: repoerrors.ErrNotEnoughBalance,
			want:          apperrors.ErrNotEnoughBalance,
		},
		{
			name: "unknown withdraw error",
			mock: func(accountRepo *mocks.MockAccount,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
				repoWithdrawErr error,
			) {
				transferWithdrawErrorMock(accountRepo,
					txManager,
					claims,
					receiverUsername,
					amount,
					repoWithdrawErr,
				)
			},
			returnedError: accountRepoUnknownWithdrawError,
			want:          accountRepoUnknownWithdrawError,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	amount := 100
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, txManager, claims, receiverUsername, amount, tt.returnedError)

			uc := NewOperationUsecase(accountRepo, nil, nil, txManager)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, amount)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestSendCoin_DepositError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var accountRepoUnknownDepositError = errors.New("unknown deposit error from account repo")
	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			receiverUsername string,
			amount int,
			repoDepositErr error,
		)
		returnedError error
		want          error
	}{
		{
			name: "unknown deposit error",
			mock: func(accountRepo *mocks.MockAccount,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
				repoDepositErr error,
			) {
				senderAccountID, receiverAccountID := 123, 456

				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), claims.UserID).
					Return(senderAccountID, nil)

				accountRepo.EXPECT().
					GetIDByUsername(gomock.Any(), receiverUsername).
					Return(receiverAccountID, nil)

				accountRepo.EXPECT().
					Withdraw(gomock.Any(), senderAccountID, amount).
					Return(nil)

				accountRepo.EXPECT().
					Deposit(gomock.Any(), receiverAccountID, amount).
					Return(repoDepositErr)

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
			},
			returnedError: accountRepoUnknownDepositError,
			want:          accountRepoUnknownDepositError,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	amount := 100
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, txManager, claims, receiverUsername, amount, tt.returnedError)

			uc := NewOperationUsecase(accountRepo, nil, nil, txManager)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, amount)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestSendCoin_TxManagerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var txManagerError = errors.New("transaction manager error")

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			receiverUsername string,
			amount int,
		)
		want error
	}{
		{
			name: "transaction manager error",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
			) {
				senderAccountID, receiverAccountID := 123, 456

				accountRepo.EXPECT().
					GetIDByUserID(gomock.Any(), claims.UserID).
					Return(senderAccountID, nil)

				accountRepo.EXPECT().
					GetIDByUsername(gomock.Any(), receiverUsername).
					Return(receiverAccountID, nil)

				accountRepo.EXPECT().
					Withdraw(gomock.Any(), senderAccountID, amount).
					Return(nil)

				accountRepo.EXPECT().
					Deposit(gomock.Any(), receiverAccountID, amount).
					Return(nil)

				transferOperation := entity.TransferOperation{
					SenderAccountID:    senderAccountID,
					RecipientAccountID: receiverAccountID,
					Amount:             amount,
				}

				operationRepo.EXPECT().
					ExecTransferOperation(gomock.Any(), transferOperation).
					Return(nil)

				txManager.EXPECT().
					Serializable(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, _ db.Mode, f func(context.Context) error) func() error {
						return func() error { return f(ctx) }
					})

				txManager.EXPECT().
					WithRetry(gomock.Any()).
					Return(txManagerError)
			},
			want: txManagerError,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	amount := 100
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, operationRepo, txManager, claims, receiverUsername, amount)

			uc := NewOperationUsecase(accountRepo, nil, nil, txManager)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, amount)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func transferOperationMock(
	accountRepo *mocks.MockAccount,
	operationRepo *mocks.MockOperation,
	txManager *mocks.MockTxManager,
	claims model.Claims,
	receiverUsername string,
	amount int,
	transferErr error,
) {
	senderAccountID, receiverAccountID := 123, 456

	accountRepo.EXPECT().
		GetIDByUserID(gomock.Any(), claims.UserID).
		Return(senderAccountID, nil)

	accountRepo.EXPECT().
		GetIDByUsername(gomock.Any(), receiverUsername).
		Return(receiverAccountID, nil)

	accountRepo.EXPECT().
		Withdraw(gomock.Any(), senderAccountID, amount).
		Return(nil)

	accountRepo.EXPECT().
		Deposit(gomock.Any(), receiverAccountID, amount).
		Return(nil)

	transferOperation := entity.TransferOperation{
		SenderAccountID:    senderAccountID,
		RecipientAccountID: receiverAccountID,
		Amount:             amount,
	}

	operationRepo.EXPECT().
		ExecTransferOperation(gomock.Any(), transferOperation).
		Return(transferErr)

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

func TestSendCoin_TransferOperationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferErr := errors.New("transfer operation error")

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			receiverUsername string,
			amount int,
		)
		want error
	}{
		{
			name: "successful coin transfer",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
			) {
				transferOperationMock(accountRepo, operationRepo, txManager, claims, receiverUsername, amount, transferErr)
			},
			want: transferErr,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	amount := 100
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, operationRepo, txManager, claims, receiverUsername, amount)

			uc := NewOperationUsecase(accountRepo, operationRepo, nil, txManager)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, amount)

			require.ErrorIs(t, err, tt.want)
		})
	}
}
func TestSendCoin_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		mock func(
			accountRepo *mocks.MockAccount,
			operationRepo *mocks.MockOperation,
			txManager *mocks.MockTxManager,
			claims model.Claims,
			receiverUsername string,
			amount int,
		)
		want error
	}{
		{
			name: "successful coin transfer",
			mock: func(accountRepo *mocks.MockAccount,
				operationRepo *mocks.MockOperation,
				txManager *mocks.MockTxManager,
				claims model.Claims,
				receiverUsername string,
				amount int,
			) {
				transferOperationMock(accountRepo, operationRepo, txManager, claims, receiverUsername, amount, nil)
			},
			want: nil,
		},
	}

	claims := model.Claims{UserID: 111}
	receiverUsername := "receiver"
	amount := 100
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accountRepo := mocks.NewMockAccount(ctrl)
			operationRepo := mocks.NewMockOperation(ctrl)
			txManager := mocks.NewMockTxManager(ctrl)
			tt.mock(accountRepo, operationRepo, txManager, claims, receiverUsername, amount)

			uc := NewOperationUsecase(accountRepo, operationRepo, nil, txManager)
			err := uc.SendCoin(context.Background(), claims, receiverUsername, amount)

			require.NoError(t, err)
		})
	}
}
