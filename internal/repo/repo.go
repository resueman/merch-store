package repo

import (
	"context"

	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo/postgres"
	"github.com/resueman/merch-store/pkg/db"
)

type User interface {
}

type Account interface {
	GetIDByUserID(ctx context.Context, userID int) (int, error)                             // +
	GetIDByUsername(ctx context.Context, username string) (int, error)                      // +
	GetBalanceByAccountID(ctx context.Context, accountID int) (int, error)                  // +
	GetPurchasesByAccountID(ctx context.Context, accountID int) ([]entity.Inventory, error) // +
	Withdraw(ctx context.Context, accountID int, amount int) error                          // +
	Deposit(ctx context.Context, accountID int, amount int) error                           // +
}

type Operation interface {
	ExecPurchaseOperation(ctx context.Context, input entity.PurchaseOperation) error            // +
	ExecTransferOperation(ctx context.Context, input entity.TransferOperation) error            // +
	GetOutgoingTransfers(ctx context.Context, accountID int) ([]entity.OutgoingTransfer, error) // +
	GetIncomingTransfers(ctx context.Context, accountID int) ([]entity.IncomingTransfer, error) // +
}

type Product interface {
	GetProductByName(ctx context.Context, name string) (*entity.Product, error) // +
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
