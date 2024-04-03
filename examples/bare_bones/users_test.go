package example

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/pgtest"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pgtestSupervisor pgtest.Supervisor

func TestMain(m *testing.M) {
	var (
		ctx = context.Background()
		err error
	)

	pgtestSupervisor, err = pgtest.NewSupervisor(
		ctx,
		pgtest.WithResetOp(pgtest.DropAllTables()),
	)
	if err != nil {
		log.Fatalf("initialize pgtest supervisor: %s", err)
	}

	// Run all of the tests, then drop any databases created by the
	// supervisor.
	os.Exit(pgtest.RunMain(ctx, m, pgtestSupervisor))
}

// Helper to get a querier, and perform any set up.
func NewTestQuerier(t testing.TB) Querier {
	t.Helper()
	ctx := context.Background()

	// Acquire a test db from the supervisor's internal pool. The test db
	// will automatically be returned to the pool at the end of the test.
	testDB := pgtestSupervisor.GetTestDB(t)

	// Connect to the test db.
	db, err := sql.Open("pgx", testDB.DataSourceName())
	require.NoError(t, err, "open test db")

	t.Cleanup(func() { db.Close() })

	require.NoError(
		t,
		CreateUsersTable(ctx, db),
		"failed to create users table",
	)

	return db
}

func TestCreateUserSuccess(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		querier = NewTestQuerier(t)

		name  = "foo"
		email = "foo@example.com"
	)

	newUser, err := CreateUser(ctx, querier, UserInfo{Name: name, Email: email})
	require.NoError(t, err, "failed to create user")

	assert.Equal(t, name, newUser.Name, "unexpected newUser.Name")
	assert.Equal(t, email, newUser.Email, "unexpected newUser.Email")
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		querier = NewTestQuerier(t)

		email = "foo@example.com"
	)

	_, err := CreateUser(ctx, querier, UserInfo{Name: "newUser1", Email: email})
	require.NoError(t, err, "failed to create first user")

	newUser2, err := CreateUser(ctx, querier, UserInfo{Name: "newUser2", Email: email})
	require.Errorf(t, err, "unexpectedly no error creating second user; newUser2=%#v", newUser2)

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		// Verify error is caused by UNIQUE constraint on users.email.
		assert.Equal(t, "23505", pgErr.Code, "unexpected pgErr.Code")
		assert.Equal(t, "users_email_key", pgErr.ConstraintName, "unexpected pgErr.ConstraintName")
	}
}

func TestGetUserByEmailSuccess(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		querier = NewTestQuerier(t)

		name  = "foo"
		email = "foo@example.com"
	)

	newUser, err := CreateUser(ctx, querier, UserInfo{Name: name, Email: email})
	require.NoError(t, err, "failed to create user")

	foundUser, err := GetUserByEmail(ctx, querier, email)
	require.NoError(t, err, "failed to get user")

	assert.Equal(t, newUser, foundUser, "unexpected foundUser; should equal newUser")
}

func TestGetUserByEmailNotFound(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		querier = NewTestQuerier(t)

		email = "foo@example.com"
	)

	foundUser, err := GetUserByEmail(ctx, querier, email)
	require.Errorf(t, err, "unexpectedly no error getting user by email; foundUser=%#v", foundUser)

	assert.ErrorIs(t, err, sql.ErrNoRows, "err should be sql.ErrNoRows")
}
