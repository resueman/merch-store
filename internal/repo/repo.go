package repo

import (
	"github.com/resueman/merch-store/internal/repo/postgres"
	"github.com/resueman/merch-store/pkg/db"
)

type User interface {
}

type Account interface {
}

type Operation interface {
}

type Product interface {
}

type Repositories struct {
	User
	Account
	Operation
	Product
}

func NewRepositories(pg db.Client) *Repositories {
	return &Repositories{
		User:      postgres.NewUserRepo(pg),
		Account:   postgres.NewAccountRepo(pg),
		Operation: postgres.NewOperationRepo(pg),
		Product:   postgres.NewProductRepo(pg),
	}
}
