package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resueman/merch-store/pkg/db"
)

var _ db.DB = (*pg)(nil) //

type pgClient struct {
	primary *pg
}

func NewPostgresClient(ctx context.Context, dsn string) (*pgClient, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pgc := &pgClient{primary: NewDB(conn)} // nil !!!!!!!!!!!!!!!

	return pgc, nil
}

func (c *pgClient) Primary() db.DB { //
	return c.primary
}

func (c *pgClient) Replica() db.DB { //
	return c.primary
}

func (c *pgClient) Close() error {
	if c.primary != nil {
		return c.primary.Close()
	}
	return nil
}
