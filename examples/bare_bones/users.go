package example

import (
	"context"
	"database/sql"
	"time"
)

// Querier defines the operations for querying the database.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Creates the users table.
func CreateUsersTable(ctx context.Context, q Querier) error {
	_, err := q.ExecContext(ctx, `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,

		CONSTRAINT users_email_not_empty CHECK (email <> ''),
		CONSTRAINT users_name_not_empty CHECK (name <> '')
	);`)
	return err
}

type User struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Name      string
}

type UserInfo struct {
	Email string
	Name  string
}

func CreateUser(ctx context.Context, q Querier, info UserInfo) (*User, error) {
	row := q.QueryRowContext(ctx, `INSERT INTO USERS (
		email, name
	) VALUES (
		$1, $2
	)
	RETURNING
		id, created_at, updated_at, email, name
	;`, info.Email, info.Name)

	var newUser User
	if err := row.Scan(
		&newUser.ID,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
		&newUser.Email,
		&newUser.Name,
	); err != nil {
		return nil, err
	}

	return &newUser, nil
}

func GetUserByEmail(ctx context.Context, q Querier, email string) (*User, error) {
	row := q.QueryRowContext(ctx, `SELECT
		id, created_at, updated_at, email, name
	FROM users
	WHERE email=$1;`, email)

	var found User
	if err := row.Scan(
		&found.ID,
		&found.CreatedAt,
		&found.UpdatedAt,
		&found.Email,
		&found.Name,
	); err != nil {
		return nil, err
	}

	return &found, nil
}
