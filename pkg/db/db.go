package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Tx Manager.

type Mode int

const (
	Write Mode = iota
	Read
)

type RetryAdatapter interface {
	WithRetry(f func() error) error
}

type TxManager interface {
	RetryAdatapter
	ReadCommitted(ctx context.Context, mode Mode, f func(ctx context.Context) error) func() error
	RepeatableRead(ctx context.Context, mode Mode, f func(ctx context.Context) error) func() error
	Serializable(ctx context.Context, mode Mode, f func(ctx context.Context) error) func() error
}

// DB.

type Pinger interface {
	Ping(ctx context.Context) error
}

type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Query struct {
	Name     string
	QueryRaw string
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

type CtxKey string

const (
	DBKey CtxKey = "db"
)

// Client.

type Client interface {
	Primary() DB
	Replica() DB
	Close() error
}
