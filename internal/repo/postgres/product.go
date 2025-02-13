package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/pkg/db"
)

type ProductRepo struct {
	db db.Client
}

func NewProductRepo(db db.Client) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) GetProductByName(ctx context.Context, name string) (*entity.Product, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("id", "name", "price").
		From("products").
		Where(sq.Eq{"name": name}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetProductByName", QueryRaw: queryRaw}
	product := entity.Product{}

	if err = primary.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrors.ErrNotFound
		}

		return nil, err
	}

	return &product, nil
}
