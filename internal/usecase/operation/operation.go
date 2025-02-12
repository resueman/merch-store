package operation

import (
	"context"
	"errors"

	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
	"github.com/resueman/merch-store/pkg/db"
)

type operationUsecase struct {
	accountRepo   repo.Account
	operationRepo repo.Operation
	productRepo   repo.Product
	txManager     db.TxManager
}

func NewOperationUsecase(account repo.Account, operation repo.Operation, product repo.Product, txManager db.TxManager) *operationUsecase {
	return &operationUsecase{
		accountRepo:   account,
		operationRepo: operation,
		productRepo:   product,
		txManager:     txManager,
	}
}

// Проверить:
// 1. Товар с заданным именем существует
// 2. Покупатель существует (уже проверено в middleware?)
// 3. Кол-во монет достаточно для покупки товара (проверяется в бд, надо вернуть соответствующую ошибку)
func (u *operationUsecase) BuyItem(ctx context.Context, itemName string) error {
	userID := ctx.Value("userID").(int)

	customerAccountID, err := u.accountRepo.GetIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return apperrors.ErrUserNotFound // ?
		}

		return err
	}

	product, err := u.productRepo.GetProductByName(ctx, itemName)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return apperrors.ErrProductNotFound
		}

		return err
	}

	// В самом начале транзакции можно было бы зарезервировать продукт,
	// вернуть ошибку, если его нет в нужном количестве.
	// Но по условию мерч бесконечен, поэтому пропускаем этот шаг.
	transaction := func(ctx context.Context) error {
		if err := u.accountRepo.Withdraw(ctx, customerAccountID, product.Price); err != nil {
			if errors.Is(err, repoerrors.ErrNotEnoughBalance) {
				return apperrors.ErrNotEnoughBalance
			}

			return err
		}

		operation := entity.PurchaseOperation{
			ItemID:            product.ID,
			CustomerAccountID: customerAccountID,
			Quantity:          1,
			TotalPrice:        product.Price,
		}

		if err := u.operationRepo.ExecPurchaseOperation(ctx, operation); err != nil {
			return err
		}

		return nil
	}

	shouldRetry := func(err error) bool {
		return !errors.Is(err, repoerrors.ErrNotEnoughBalance)
	}

	serializable := u.txManager.Serializable(ctx, transaction)
	if err = u.txManager.WithRetry(serializable, shouldRetry); err != nil {
		return err
	}

	return nil
}

// Проверить:
// 1. Пользователь отправляет монеты не себе
// 2. Пользователь отправляет положительное кол-во монет
// 3. Получатель существует
// 4. Отправитель существует (уже проверено в middleware?)
// 5. Кол-во монет достаточно для перевода (проверяется в бд, надо вернуть соответствующую ошибку)
func (u *operationUsecase) SendCoin(ctx context.Context, receiverUsername string, amount int) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}

	userID := ctx.Value("userID").(int)

	senderAccountID, err := u.accountRepo.GetIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return apperrors.ErrUserNotFound // ?
		}

		return err
	}

	receiverAccountID, err := u.accountRepo.GetIDByUsername(ctx, receiverUsername)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return apperrors.ErrUserNotFound
		}

		return err
	}

	if senderAccountID == receiverAccountID {
		return apperrors.ErrSelfTransfer
	}

	transaction := func(ctx context.Context) error {
		if err := u.accountRepo.Withdraw(ctx, senderAccountID, amount); err != nil {
			if errors.Is(err, repoerrors.ErrNotEnoughBalance) {
				return apperrors.ErrNotEnoughBalance
			}

			return err
		}

		if err := u.accountRepo.Deposit(ctx, receiverAccountID, amount); err != nil {
			return err
		}

		operation := entity.TransferOperation{
			SenderAccountID:    senderAccountID,
			RecipientAccountID: receiverAccountID,
			Amount:             amount,
		}

		if err := u.operationRepo.ExecTransferOperation(ctx, operation); err != nil {
			return err
		}

		return nil
	}

	shouldRetry := func(err error) bool {
		return !errors.Is(err, repoerrors.ErrNotEnoughBalance)
	}

	serializable := u.txManager.Serializable(ctx, transaction)
	if err = u.txManager.WithRetry(serializable, shouldRetry); err != nil { // если не ошибка ErrNotEnoughBalance, то повторить транзакцию
		return err
	}

	return nil
}
