package postgres

import "github.com/resueman/merch-store/pkg/db"

type UserRepo struct {
	db db.Client
}

func NewUserRepo(db db.Client) *UserRepo {
	return &UserRepo{db: db}
}
