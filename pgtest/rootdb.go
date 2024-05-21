package pgtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgPool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Close()
	Config() *pgxpool.Config
}

type rootDB struct {
	db pgPool
}

func (db *rootDB) close() {
	db.db.Close()
}

func (db *rootDB) createDatabase(ctx context.Context, name string) error {
	query := fmt.Sprintf("CREATE DATABASE %q;", name)
	if _, err := db.db.Exec(ctx, query); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// It seems like either code 42P04 or 23505 can be
			// returned if a database already exists with the
			// specified name, so we will treat them the same.
			if pgErr.Code == pgerrcode.DuplicateDatabase || pgErr.Code == pgerrcode.UniqueViolation {
				return &databaseAlreadyExistsWithName{name: name, cause: err}
			}
		}

		return err
	}

	return nil
}

func (db *rootDB) dropDatabase(ctx context.Context, name string) error {
	return dropDatabase(ctx, db.db, name)
}

func (db *rootDB) getAllDatabases(ctx context.Context) ([]string, error) {
	return getAllDatabases(ctx, db.db)
}
