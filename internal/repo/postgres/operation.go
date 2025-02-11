package postgres

import "github.com/resueman/merch-store/pkg/db"

type OperationRepo struct {
	db db.Client
}

func NewOperationRepo(db db.Client) *OperationRepo {
	return &OperationRepo{db: db}
}
