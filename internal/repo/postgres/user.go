package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/pkg/db"
)

type UserRepo struct {
	client db.Client
}

func NewUserRepo(client db.Client) *UserRepo {
	return &UserRepo{client: client}
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Replica()
	}

	queryRaw, args, err := database.QueryBuilder().
		Select("id", "username", "password").
		From("users").
		Where(sq.Eq{"username": username}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{QueryRaw: queryRaw, Name: "GetUserByUsername"}
	row := database.QueryRow(ctx, query, args...)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Username, &user.Hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repoerrors.ErrNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *entity.CreateUserInput) (int, error) {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Primary()
	}

	tx, err := database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("error beginning transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover from panic: %v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = fmt.Errorf("rollback error: %v", errRollback)
			}

			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			err = fmt.Errorf("error committing transaction: %v", err)
		}
	}()

	createUserRaw, args, err := database.QueryBuilder().
		Insert("users").
		Columns("username", "password").
		Values(user.Username, user.Hash).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, err
	}

	row := tx.QueryRow(ctx, createUserRaw, args...)

	var userID int
	if err := row.Scan(&userID); err != nil {
		return 0, err
	}

	createAccountRaw, args, err := database.QueryBuilder().
		Insert("accounts").
		Columns("user_id").
		Values(userID).
		ToSql()

	if err != nil {
		return 0, err
	}

	if _, err = tx.Exec(ctx, createAccountRaw, args...); err != nil {
		return 0, err
	}

	return userID, nil
}
