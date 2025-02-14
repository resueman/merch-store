package postgres

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resueman/merch-store/pkg/db"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

var _ db.DB = (*pg)(nil)

type key string

const (
	TxKey key = "tx"
)

func ContextWithTx(ctx context.Context, tx pgx.Tx, database db.DB) context.Context {
	newCtx := context.WithValue(ctx, TxKey, tx)
	newCtx = context.WithValue(newCtx, db.DBKey, database)

	return newCtx
}

type pg struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType

	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
}

func NewDB(dbc *pgxpool.Pool) *pg { // db.DB {
	postgresDB := &pg{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,

		pool:    dbc,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// задание недефолтных значейний, попытки подключиться

	return postgresDB
}

func (p *pg) QueryBuilder() squirrel.StatementBuilderType {
	return p.builder
}

func (p *pg) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *pg) Exec(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.QueryRaw, args...)
	}

	return p.pool.Exec(ctx, q.QueryRaw, args...)
}

func (p *pg) Query(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.QueryRaw, args...)
	}

	return p.pool.Query(ctx, q.QueryRaw, args...)
}

func (p *pg) QueryRow(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}

	return p.pool.QueryRow(ctx, q.QueryRaw, args...)
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx, nil
	}

	return p.pool.BeginTx(ctx, txOptions)
}

func (p *pg) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}

	return nil
}
