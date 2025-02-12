package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type QueryExecutor interface {
	QueryRow(ctx context.Context, q Query, args ...interface{}) pgx.Row
	Query(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
}

type Pooler interface {
	Pinger
	Transactor
	QueryExecutor
	Close() error
}

type QueryBuilder interface {
	QueryBuilder() squirrel.StatementBuilderType
}

type DB interface {
	Pooler
	QueryBuilder
	Transactor
}

type Client interface {
	Primary() DB
	Replica() DB
	Close() error
}

type RetryAdatapter interface {
	WithRetry(f func() error, shouldRetry func(error) bool) error
}

type TxManager interface {
	RetryAdatapter
	ReadCommitted(ctx context.Context, f func(ctx context.Context) error) func() error
	RepeatableRead(ctx context.Context, f func(ctx context.Context) error) func() error
	Serializable(ctx context.Context, f func(ctx context.Context) error) func() error
}
