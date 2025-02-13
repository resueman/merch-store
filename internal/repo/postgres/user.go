package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
	"github.com/resueman/merch-store/pkg/db"
)

type UserRepo struct {
	client db.Client
}

func NewUserRepo(client db.Client) *UserRepo {
	return &UserRepo{client: client}
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	primary := r.client.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("id", "username", "password").
		From("users").
		Where(sq.Eq{"username": username}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{QueryRaw: queryRaw, Name: "GetUserByUsername"}
	row := primary.QueryRow(ctx, query, args...)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Username, &user.Hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *entity.CreateUserInput) (int, error) {
	primary := r.client.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Insert("users").
		Columns("username", "password").
		Values(user.Username, user.Hash).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, err
	}

	query := db.Query{QueryRaw: queryRaw, Name: "CreateUser"}
	row := primary.QueryRow(ctx, query, args...)

	var id int
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
