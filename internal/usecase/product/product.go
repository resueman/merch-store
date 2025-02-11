package product

import "github.com/resueman/merch-store/internal/repo"

type productUsecase struct {
	repo repo.Product
}

func NewProductUsecase(repo repo.Product) *productUsecase {
	return &productUsecase{repo: repo}
}
