package txmanager

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/pkg/db/postgres"
)

type TxManager struct {
	client     db.Client
	timeout    int
	maxRetries int
}

func NewTxManager(client db.Client, timeout int, maxRetries int) *TxManager {
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
	errChan := make(chan error, 1)

	for range m.maxRetries {
		go func() {
			errChan <- f()
		}()

		select {
		case err := <-errChan:
			if err == nil {
				return nil
			}

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && (pgErr.Code == "40001" ||
				pgErr.Code == "40P01" || pgErr.Code == "55P03") {
				continue
			}

			return err
		case <-time.After(time.Duration(m.timeout) * time.Second):
			continue
		}
	}

	return <-errChan
}

func (m *TxManager) transaction(ctx context.Context, opts pgx.TxOptions, mode db.Mode, f func(ctx context.Context) error) (err error) {
	tx, ok := ctx.Value(postgres.TxCtxKey).(pgx.Tx)
	if ok {
		return f(ctx)
	}

	var database db.DB
	if mode == db.Read {
		database = m.client.Replica()
	} else {
		database = m.client.Primary()
	}

	ctx = context.WithValue(ctx, db.DBKey, database)

	tx, err = database.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}

	ctx = postgres.ContextWithTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("recover from panic: %v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "rollback error: %v", errRollback)
			}

			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			err = errors.Wrap(err, "error committing transaction")
		}
	}()

	if err = f(ctx); err != nil {
		err = errors.Wrap(err, "error executing code inside transaction")
	}

	return err
}
