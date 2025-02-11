package usecase

import (
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase/account"
	"github.com/resueman/merch-store/internal/usecase/operation"
	"github.com/resueman/merch-store/internal/usecase/product"
	"github.com/resueman/merch-store/internal/usecase/user"
)

type User interface {
}

type Account interface {
}

type Operation interface {
}

type Product interface {
}

type Usecase struct {
	User
	Account
	Operation
	Product
}

func NewUsecase(repo *repo.Repositories) *Usecase {
	return &Usecase{
		User:      user.NewUserUsecase(repo.User),
		Account:   account.NewAccountUsecase(repo.Account),
		Operation: operation.NewOperationUsecase(repo.Operation),
		Product:   product.NewProductUsecase(repo.Product),
	}
}
