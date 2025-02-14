package usecase

import (
	"context"
	"time"

	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase/account"
	"github.com/resueman/merch-store/internal/usecase/auth"
	"github.com/resueman/merch-store/internal/usecase/operation"
	"github.com/resueman/merch-store/pkg/db"
)

type Auth interface {
	GenerateToken(ctx context.Context, input model.AuthRequestInput) (string, error)
	ParseToken(ctx context.Context, tokenString string) (model.Claims, error)
}

type Account interface {
	GetInfo(ctx context.Context) (*model.AccountInfo, error)
}

type Operation interface {
	BuyItem(ctx context.Context, itemID string) error
	SendCoin(ctx context.Context, receiverUsername string, amount int) error
}

type Usecase struct {
	Auth
	Account
	Operation
	db.TxManager
}

type PasswordManager interface {
	HashPassword(password string) string
	ComparePassword(password, hash string) bool
}

func NewUsecase(repo *repo.Repositories, txManager db.TxManager,
	passwordManager PasswordManager, secretKey string, tokenTTL time.Duration) *Usecase {
	return &Usecase{
		Auth:      auth.NewAuthUsecase(repo.User, passwordManager, secretKey, tokenTTL),
		Account:   account.NewAccountUsecase(repo.Account, repo.Operation, repo.Product, txManager),
		Operation: operation.NewOperationUsecase(repo.Account, repo.Operation, repo.Product, txManager),
		TxManager: txManager,
	}
}
