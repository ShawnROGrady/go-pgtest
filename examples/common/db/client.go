package db

import (
	"context"
	"errors"

	"github.com/ShawnROGrady/go-pgtest/examples/common"
	"github.com/ShawnROGrady/go-pgtest/examples/common/dbqueries"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Client, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &Client{
		pool: pool,
	}, nil
}

func (cl *Client) queries() *pgxQueries {
	return &pgxQueries{
		inner: dbqueries.New(cl.pool),
	}
}

func (cl *Client) Migrate() error {
	m, err := common.Migrations().ForPgxPool(cl.pool)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (cl *Client) Ping(ctx context.Context) error {
	return cl.pool.Ping(ctx)
}

func (cl *Client) Close() {
	cl.pool.Close()
}
