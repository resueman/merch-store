package postgres

import "github.com/resueman/merch-store/pkg/db"

type ProductRepo struct {
	db db.Client
}

func NewProductRepo(db db.Client) *ProductRepo {
	return &ProductRepo{db: db}
}
