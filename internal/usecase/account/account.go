package account

import (
	"context"

	"github.com/resueman/merch-store/internal/delivery/ctxkey"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase/converter"
	"github.com/resueman/merch-store/pkg/db"
)

type accountUsecase struct {
	accountRepo   repo.Account
	operationRepo repo.Operation
	productRepo   repo.Product
	txManager     db.TxManager
}

func NewAccountUsecase(account repo.Account, operation repo.Operation,
	product repo.Product, txManager db.TxManager) *accountUsecase {
	return &accountUsecase{
		accountRepo:   account,
		operationRepo: operation,
		productRepo:   product,
		txManager:     txManager,
	}
}

// Проверить:
func (u *accountUsecase) GetInfo(ctx context.Context) (*model.AccountInfo, error) {
	userID := ctx.Value(ctxkey.ClaimsKey).(model.Claims).UserID

	accountID, err := u.accountRepo.GetIDByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance := 0
	purchases := []entity.Purchase{}
	incomingTransfers := []entity.Transfer{}
	outgoingTransfers := []entity.Transfer{}
	transaction := func(ctx context.Context) error {
		var err error
		// в этот момент кто-то может прислать монет
		// пользователь может на одной странице купить товар, а на второй смотреть баланс и информацию о покупках
		balance, err = u.accountRepo.GetBalanceByAccountID(ctx, accountID)
		if err != nil {
			return err
		}

		// пользователь может на одной странице купить товар, а на второй смотреть баланс и информацию о покупках
		purchases, err = u.accountRepo.GetPurchasesByAccountID(ctx, accountID)
		if err != nil {
			return err
		}

		// в этот момент кто-то может прислать монет
		incomingTransfers, err = u.operationRepo.GetIncomingTransfers(ctx, accountID)
		if err != nil {
			return err
		}

		// на одной странице пользователь отправляет монеты, а на другой подгружает баланс
		outgoingTransfers, err = u.operationRepo.GetOutgoingTransfers(ctx, accountID)
		if err != nil {
			return err
		}

		return nil
	}

	shouldRetry := func(err error) bool {
		return true
	}

	serializable := u.txManager.Serializable(ctx, transaction)
	if err = u.txManager.WithRetry(serializable, shouldRetry); err != nil {
		return nil, err
	}

	info := &model.AccountInfo{
		Balance:           balance,
		Inventory:         converter.ConvertPurchasesToInventory(purchases),
		IncomingTransfers: converter.ConvertTransfersToIncomingTransfers(incomingTransfers),
		OutgoingTransfers: converter.ConvertTransfersToOutgoingTransfers(outgoingTransfers),
	}

	return info, nil
}

// Если сделаем Repeatable read, то будет проблема, что если между чтением баланса и
// чтением количества полученных монет кто-то прислал монеты, то мы увидим баланс
// на момент начала транзакции.
// (Аномалия phantom read, вставка в таблицу транзакций)

// Или если мы послали монеты и тут же начали смотреть информацию об аккаунте,
// то может быть проблема при Repeatable read, что изменения об отправке монет
// зафиксируются между чтением баланса и чтением отправленных монет.
// (Аномалия phantom read, вставка в таблицу транзакций)

// Для борьбы с фантомными чтениями можно использовать Serializable.
// Но тогда придется обрабатывать 40001 ошибку. Возникает проблема, что если
// пользователи будут отправлять мне монеты, а я буду получать информацию о них, то
// чья-то транзакция может завершиться с ошибкой и надо будет ее повторять.

// Можно использовать repeatable read, но корректировать баланс вручную.
// actualBalance := totalReceived - totalSent - totalPurchase + initialBalance -- ужасное решение

// Еще есть repeatable read + SELECT ... FOR SHARE, но надо посмотреть подробнее.
