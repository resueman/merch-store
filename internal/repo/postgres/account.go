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

type AccountRepo struct {
	db db.Client
}

func NewAccountRepo(db db.Client) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) GetIDByUserID(ctx context.Context, userID int) (int, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("id").
		From("accounts").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return 0, err
	}

	query := db.Query{Name: "GetAccountID", QueryRaw: queryRaw}

	var accountID int
	if err = primary.QueryRow(ctx, query, args...).Scan(&accountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repoerrors.ErrNotFound
		}

		return 0, err
	}

	return accountID, nil
}

func (r *AccountRepo) GetIDByUsername(ctx context.Context, username string) (int, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("id").
		From("accounts").
		Join("users ON accounts.user_id = users.id").
		Where(sq.Eq{"users.username": username}).
		ToSql()

	if err != nil {
		return 0, err
	}

	query := db.Query{Name: "GetAccountIDByUsername", QueryRaw: queryRaw}

	var accountID int
	if err = primary.QueryRow(ctx, query, args...).Scan(&accountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repoerrors.ErrNotFound
		}

		return 0, err
	}

	return accountID, nil
}

func (r *AccountRepo) GetBalanceByAccountID(ctx context.Context, accountID int) (int, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("balance").
		From("accounts").
		Where(sq.Eq{"id": accountID}).
		ToSql()

	if err != nil {
		return 0, err
	}

	query := db.Query{Name: "GetBalance", QueryRaw: queryRaw}

	var balance int
	if err = primary.QueryRow(ctx, query, args...).Scan(&balance); err != nil {
		return 0, err
	}

	return balance, nil
}

func (r *AccountRepo) GetPurchasesByAccountID(ctx context.Context, accountID int) ([]entity.Purchase, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("p.name", "SUM(ops.quantity) AS quantity").
		From("purchase_operations ops").
		Join("products p ON ops.product_id = p.id").
		Where(sq.Eq{"ops.customer_account_id": accountID}).
		GroupBy("p.name").
		OrderBy("quantity DESC").
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetUserPurchases", QueryRaw: queryRaw}
	rows, err := primary.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	inventory := entity.Purchase{}
	inventories := []entity.Purchase{}

	for rows.Next() {
		if err = rows.Scan(&inventory.Name, &inventory.Quantity); err != nil {
			return nil, err
		}

		inventories = append(inventories, inventory)
	}

	return inventories, nil
}

func (r *AccountRepo) Withdraw(ctx context.Context, accountID int, amount int) error {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	selectQuery, args, err := builder.Select("balance").
		From("accounts").
		Where(sq.Eq{"account_id": accountID}).
		Suffix("FOR UPDATE").
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "Withdraw: get balance for update", QueryRaw: selectQuery}

	var balance int
	if err = primary.QueryRow(ctx, query, args...).Scan(&balance); err != nil {
		return err
	}

	if balance < amount {
		return repoerrors.ErrNotEnoughBalance
	}

	updateQuery, args, err := builder.Update("accounts").
		Set("balance", sq.Expr("balance - ?", amount)).
		Where(sq.Eq{"account_id": accountID}).
		ToSql()

	if err != nil {
		return err
	}

	query = db.Query{Name: "Withdraw: update balance", QueryRaw: updateQuery}
	if _, err = primary.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *AccountRepo) Deposit(ctx context.Context, accountID int, amount int) error {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Update("accounts").
		Set("balance", sq.Expr("balance + ?", amount)).
		Where(sq.Eq{"id": accountID}).
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "Deposit", QueryRaw: queryRaw}
	if _, err = primary.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
