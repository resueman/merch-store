package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/resueman/merch-store/pkg/db"
)

type TxManager struct {
	client     db.Client
	timeout    time.Duration
	maxRetries int
}

func NewTxManager(client db.Client, timeout time.Duration, maxRetries int) *TxManager {
	return &TxManager{
		client:     client,
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

func (m *TxManager) ReadCommitted(ctx context.Context, mode db.Mode, f func(ctx context.Context) error) func() error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}

	return func() error {
		return m.transaction(ctx, txOpts, mode, f)
	}
}

func (m *TxManager) RepeatableRead(ctx context.Context, mode db.Mode, f func(ctx context.Context) error) func() error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.RepeatableRead}

	return func() error {
		return m.transaction(ctx, txOpts, mode, f)
	}
}

func (m *TxManager) Serializable(ctx context.Context, mode db.Mode, f func(ctx context.Context) error) func() error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.Serializable}

	return func() error {
		return m.transaction(ctx, txOpts, mode, f)
	}
}

func (m *TxManager) WithRetry(f func() error) error {
	var err error
	for range m.maxRetries {
		if err = f(); err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "40001" ||
			pgErr.Code == "40P01" || pgErr.Code == "55P03") {
			continue
		} else {
			return err
		}
	}

	return errors.New("error executing transaction: retries exceeded")
}

func (m *TxManager) transaction(
	ctx context.Context,
	opts pgx.TxOptions,
	mode db.Mode,
	f func(ctx context.Context) error,
) (err error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return f(ctx)
	}

	var database db.DB
	if mode == db.Read {
		database = m.client.Replica()
	} else {
		database = m.client.Primary()
	}

	tx, err = database.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}

	ctx = ContextWithTx(ctx, tx, database)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover from panic: %v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = fmt.Errorf("rollback error: %w", errRollback)
			}

			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			err = fmt.Errorf("error committing transaction: %w", err)
		}
	}()

	if err = f(ctx); err != nil {
		err = fmt.Errorf("error executing code inside transaction: %w", err)
	}

	return err
}
