package usecase

import (
	"context"

	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase/account"
	"github.com/resueman/merch-store/internal/usecase/operation"
	"github.com/resueman/merch-store/internal/usecase/product"
	"github.com/resueman/merch-store/internal/usecase/user"
	"github.com/resueman/merch-store/pkg/db"
)

type User interface {
}

type Account interface {
	GetInfo(ctx context.Context) (*entity.AccountInfo, error)
}

type Operation interface {
	BuyItem(ctx context.Context, itemID string) error
	SendCoin(ctx context.Context, receiverUsername string, amount int) error
}

type Product interface {
	GetProductByName(ctx context.Context, name string) (*entity.Product, error)
}

type Usecase struct {
	User
	Account
	Operation
	Product
	db.TxManager
}

func NewUsecase(repo *repo.Repositories, txManager db.TxManager) *Usecase {
	return &Usecase{
		User:      user.NewUserUsecase(repo.User),
		Account:   account.NewAccountUsecase(repo.Account, repo.Operation, repo.Product, txManager),
		Operation: operation.NewOperationUsecase(repo.Account, repo.Operation, repo.Product, txManager),
		Product:   product.NewProductUsecase(repo.Product),
		TxManager: txManager,
	}
}
