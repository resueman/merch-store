package user

import "github.com/resueman/merch-store/internal/repo"

type userUsecase struct {
	repo repo.User
}

func NewUserUsecase(repo repo.User) *userUsecase {
	return &userUsecase{repo: repo}
}

// попытка создать пользователя с существующим именем
