package postgres

import "github.com/resueman/merch-store/pkg/db"

type AccountRepo struct {
	db db.Client
}

func NewAccountRepo(db db.Client) *AccountRepo {
	return &AccountRepo{db: db}
}
