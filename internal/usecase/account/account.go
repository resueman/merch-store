package account

import "github.com/resueman/merch-store/internal/repo"

type accountUsecase struct {
	repo repo.Account
}

func NewAccountUsecase(repo repo.Account) *accountUsecase {
	return &accountUsecase{repo: repo}
}
