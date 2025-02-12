package product

import (
	"context"
	"errors"

	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
)

type productUsecase struct {
	productRepo repo.Product
}

func NewProductUsecase(productRepo repo.Product) *productUsecase {
	return &productUsecase{productRepo: productRepo}
}

// +
func (u *productUsecase) GetProductByName(ctx context.Context, name string) (*entity.Product, error) {
	product, err := u.productRepo.GetProductByName(ctx, name)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return nil, apperrors.ErrProductNotFound
		}

		return nil, err
	}

	return product, nil
}

// +
/*func (u *productUsecase) GetProductByID(ctx context.Context, id string) (*entity.Product, error) {
	product, err := u.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return product, nil
}*/
