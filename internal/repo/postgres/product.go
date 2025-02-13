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
	client db.Client
}

func NewProductRepo(client db.Client) *ProductRepo {
	return &ProductRepo{client: client}
}

func (r *ProductRepo) GetProductByName(ctx context.Context, name string) (*entity.Product, error) {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Replica()
	}

	queryRaw, args, err := database.QueryBuilder().
		Select("id", "name", "price").
		From("products").
		Where(sq.Eq{"name": name}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetProductByName", QueryRaw: queryRaw}
	product := entity.Product{}

	if err = database.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrors.ErrNotFound
		}

		return nil, err
	}

	return &product, nil
}
