package pgtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rootDB struct {
	db *pgxpool.Pool
}

func (db *rootDB) close() {
	db.db.Close()
}

func (db *rootDB) createDatabase(ctx context.Context, name string) error {
	query := fmt.Sprintf("CREATE DATABASE %q;", name)
	if _, err := db.db.Exec(ctx, query); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return &databaseAlreadyExistsWithName{name: name, cause: err}
		}

		return err
	}

	return nil
}

func (db *rootDB) dropDatabase(ctx context.Context, name string) error {
	query := fmt.Sprintf("DROP DATABASE %q;", name)
	if _, err := db.db.Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

func (db *rootDB) getAllDatabases(ctx context.Context) ([]string, error) {
	return getAllDatabases(ctx, db.db)
}
