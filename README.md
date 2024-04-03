# go-pgtest

Helpers for running integration tests against postgres databases.

Internally this package provides high level utilities for managing real
postgres databases for integration tests, with support for both creating a new
database per-test and using a pool of databases across tests.

## Usage

Generally the preferred usage is to re-use databases across tests. Particularly
for larger test suites, this is considerably faster than creating a new
database per-test.

To illustrate, we will look at the example in `examples/bare_bones`. Here we
have a fairly straightforward package which defines operations around a `users`
table:

```go
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
```

To test this, we will first need to create a `TestMain` to initialize the
pgtest supervisor:

```go
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
```

The `pgtest.WithResetOp(pgtest.DropAllTables())` tells the supervisor that it
should drop any existing tables on test databases before they can be used by
other tests. This is also the default behaviour. Alternatively you can instruct
the supervisor to just truncate the tables with `pgtest.TruncateAllTables()`,
or to drop/truncate a subset of the tables with `pgtest.DropAllTablesExcept`
and `pgtest.TruncateAllTablesExcept` respectively. The last two can be useful
if you're application has a more complex migration process, which there is an
example of in `examples/simple`.

Next we will create a helper to simplify our test cases:

```go
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
```

Although this is optional, having a helper like this in conjunction with
`TestMain` allows our individual tests to be written without any knowledge of
`pgtest`.

Now we are all ready to go, and can start adding tests as:

```go
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
```

Alternatively, `pgtest` does also export functionality to just create a brand
new database per-test then drop it after the test completes through
`pgtest.NewTestDB`. Although pooling generally results in significantly faster
test runs, `pgtest.NewTestDB` is simpler (no `TestMain` required) and may be
fine for smaller test suites. Additionally `pgtest.NewTestDB` might make just
make more sense if you have some tests which require different set up than the
rest of your suite. For example, if you have a more complex migration set up it
might make sense to use pooling and the `pgtest.Supervisor` for tests where the
migrations are already run then `pgtest.NewTestDB` to test the migrations
themselves.

## Configuration

The main way to configure `pgtest` is through environment variables. These
environment variables are:

1. `PGTEST_HOST` - the postgres server host. Defaults to `"localhost"`.
2. `PGTEST_PORT` - the postgres server port. Defaults to `5432`.
3. `PGTEST_USER` - the database user name. Defaults to the `"USER"` environment
   variable.
4. `PGTEST_PASSWORD` - the login password.
5. `PG_TEST_KEEP_DATABASES_FOR_FAILED` - whether or not to keep databases for
   failed tests. Defaults to `false`.

The `PG_TEST_KEEP_DATABASES_FOR_FAILED` option is provided to assist in
debugging failed tests. In particular, for more complex applications there are
sometimes cases where the easiest way to debug a failed test is to inspect the
state of the database at the time of the failure. Normally the `pgtest`
supervisor tries to re-use databases across tests, then drop them once all
tests are complete, so the `PG_TEST_KEEP_DATABASES_FOR_FAILED` option
effectively tells `pgtest` to "forget" about a database if a test fails
allowing you to inspect/modify it once the test suite completes.

Check the `Makefile` in this repo for an example of using these configuration
variables to run integration tests against a docker container.

## Caveats

Despite using `TestMain`, there is no guarantee that the test databases created
by a supervisor will actually be dropped. In particular, if a test panics or
the test times out then the cleanup will not be run. This is a side effect of
the fact that if a goroutine triggers a panic, there is no way to recover it
from a separate goroutine and the program will just crash. There are some
things we can try to do in order to improve this, such as detecting the test
timeout and trying to perform any cleanup before it occurs, we won't be able to
easily avoid the issue entirely. As such, you may want to periodically clean up
any test databases on your system with something like:

```
psql postgres --list | grep pg_test | awk '{print $1}' | xargs -I{} psql postgres -c "DROP DATABASE {};"
```
