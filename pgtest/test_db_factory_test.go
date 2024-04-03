package pgtest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
)

// A sequentialRandSource is an implementation of rand.Source which just
// returns sequential numbers for testing purposes.
type sequentialRandSource struct {
	d atomic.Int64
}

func (src *sequentialRandSource) Int63() int64 {
	return src.d.Add(1)
}

func (src *sequentialRandSource) Seed(seed int64) {
	src.d.Store(seed)
}

func TestDBFactoryCreateTestDBImmediateSuccess(t *testing.T) {
	var (
		ctx = context.Background()

		randSource = new(sequentialRandSource)
		rng        = rand.New(randSource)

		user = "foo"
		host = "localhost"
		port = 5432

		paramFactory = func(dbName string) connparams.ConnectionParams {
			return connparams.New(
				dbName,
				connparams.WithUser(user),
				connparams.WithHost(host),
				connparams.WithPort(port),
			)
		}
	)

	// Set up: create a rootDB with a mockPool.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}
	defer mockPool.Close()

	rootDB := &rootDB{db: mockPool}
	defer rootDB.close()

	// Set up: create the factory.
	factory := &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       rootDB,
		rng:          rng,
	}

	// Set up: mock out the operations performed by the rootDB.
	mockPool.
		ExpectExec(`CREATE DATABASE "pg_test_\d"`).
		WillReturnResult(pgxmock.NewResult("CREATE DATABASE", 1)).
		Times(1)

	// Create the test database using the factory.
	created, err := factory.createTestDB(ctx)
	if err != nil {
		t.Fatalf("unexpected error from factory.createTestDB: %s", err)
	}

	expectedCreated := &testDB{
		connparams: connparams.New(
			"pg_test_1",
			connparams.WithUser(user),
			connparams.WithHost(host),
			connparams.WithPort(5432),
		),
	}

	if diff := cmp.Diff(
		created, expectedCreated,
		cmp.AllowUnexported(testDB{}),
	); diff != "" {
		t.Errorf("unexpected created TestDB (-got, +want):\n%s", diff)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("mock pool has unfulfilled expectations: %s", err)
	}
}

func TestDBFactoryCreateTestDBSucceedsSecondTime(t *testing.T) {
	var (
		ctx = context.Background()

		randSource = new(sequentialRandSource)
		rng        = rand.New(randSource)

		user = "foo"
		host = "localhost"
		port = 5432

		paramFactory = func(dbName string) connparams.ConnectionParams {
			return connparams.New(
				dbName,
				connparams.WithUser(user),
				connparams.WithHost(host),
				connparams.WithPort(port),
			)
		}
	)

	// Set up: create a rootDB with a mockPool.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}
	defer mockPool.Close()

	rootDB := &rootDB{db: mockPool}
	defer rootDB.close()

	// Set up: create the factory.
	factory := &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       rootDB,
		rng:          rng,
	}

	// Set up: mock out the operations performed by the rootDB. The first
	// call to create the database fails with a duplicate database error,
	// but the second succeeds.
	mockPool.
		ExpectExec(`CREATE DATABASE "pg_test_\d"`).
		Times(1).
		WillReturnError(&pgconn.PgError{
			Severity: "ERROR",
			Code:     "42P04",
			Message:  `database "pg_test_1" already exists"`,
		})

	mockPool.
		ExpectExec(`CREATE DATABASE "pg_test_\d"`).
		WillReturnResult(pgxmock.NewResult("CREATE", 1)).
		Times(1)

	// Create the test database using the factory.
	created, err := factory.createTestDB(ctx)
	if err != nil {
		t.Fatalf("unexpected error from factory.createTestDB: %s", err)
	}

	expectedCreated := &testDB{
		connparams: connparams.New(
			"pg_test_2",
			connparams.WithUser(user),
			connparams.WithHost(host),
			connparams.WithPort(5432),
		),
	}

	if diff := cmp.Diff(
		created, expectedCreated,
		cmp.AllowUnexported(testDB{}),
	); diff != "" {
		t.Errorf("unexpected created TestDB (-got, +want):\n%s", diff)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("mock pool has unfulfilled expectations: %s", err)
	}
}

func TestDBFactoryCreateTestDBEventuallyFailsOnAlreadyExists(t *testing.T) {
	var (
		ctx = context.Background()

		randSource = new(sequentialRandSource)
		rng        = rand.New(randSource)

		user = "foo"
		host = "localhost"
		port = 5432

		paramFactory = func(dbName string) connparams.ConnectionParams {
			return connparams.New(
				dbName,
				connparams.WithUser(user),
				connparams.WithHost(host),
				connparams.WithPort(port),
			)
		}
	)

	// Set up: create a rootDB with a mockPool.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}
	defer mockPool.Close()

	rootDB := &rootDB{db: mockPool}
	defer rootDB.close()

	// Set up: create the factory.
	factory := &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       rootDB,
		rng:          rng,
	}

	// Set up: mock out the operations performed by the rootDB. Here the
	// create call continously fails.
	mockPool.
		ExpectExec(`CREATE DATABASE "pg_test_\d"`).
		Times(5).
		WillReturnError(&pgconn.PgError{
			Severity: "ERROR",
			Code:     "42P04",
			Message:  `database already exists"`,
		})

	// Attempt to create a test database using the factory.
	created, err := factory.createTestDB(ctx)
	if err == nil {
		t.Fatalf("unexpectedly error from factory.createTestDB; got %#v", created)
	}

	// Verify error message contains the duplicate db name.
	expectedDuplicateDBName := "pg_test_5"
	if !strings.Contains(err.Error(), expectedDuplicateDBName) {
		t.Errorf(`unexpectedly not strings.Contains(%q, %q)`, err, expectedDuplicateDBName)
	}

	// Verify error satisfies errors.Is.
	if !errors.Is(err, &databaseAlreadyExistsWithName{name: expectedDuplicateDBName}) {
		t.Errorf(`unexpectedly not errors.Is(%q, &databaseAlreadyExistsWithName{name: %q}); err = %#v`, err, expectedDuplicateDBName, err)
	}

	// Verify the mock was called as expected.
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("mock pool has unfulfilled expectations: %s", err)
	}
}

func TestDBFactoryDestroyTestDBSuccess(t *testing.T) {
	var (
		ctx = context.Background()

		randSource = new(sequentialRandSource)
		rng        = rand.New(randSource)

		paramFactory = func(dbName string) connparams.ConnectionParams {
			return connparams.New(dbName, connparams.WithUser("foo"), connparams.WithHost("localhost"), connparams.WithPort(5432))
		}

		toDrop = &testDB{
			connparams: paramFactory("pg_test_1234"),
		}
	)

	// Set up: create a rootDB with a mockPool.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}
	defer mockPool.Close()

	rootDB := &rootDB{db: mockPool}
	defer rootDB.close()

	// Set up: create the factory.
	factory := &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       rootDB,
		rng:          rng,
	}
	defer factory.close()

	// Set up: mock out the operations performed by the rootDB.
	mockPool.
		ExpectExec(regexp.QuoteMeta(
			`DROP DATABASE "pg_test_1234"`,
		)).
		WillReturnResult(pgxmock.NewResult("DROP DATABASE", 1)).
		Times(1)

	// Attempt to drop the test db.
	if err := factory.destroyTestDB(ctx, toDrop); err != nil {
		t.Fatalf("factory.destroyTestDB(ctx, toDrop) = %s; want nil", err)
	}

	// Verify the mock was called as expected.
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("mock pool has unfulfilled expectations: %s", err)
	}
}

func TestDBFactoryDestroyAllTestDBsSuccess(t *testing.T) {
	var (
		ctx = context.Background()

		randSource = new(sequentialRandSource)
		rng        = rand.New(randSource)

		paramFactory = func(dbName string) connparams.ConnectionParams {
			return connparams.New(dbName, connparams.WithUser("foo"), connparams.WithHost("localhost"), connparams.WithPort(5432))
		}

		existingDBNames = []string{
			"pg_test_1",
			"postgres",
			"some_db",
			"pg_test_2",
			"pg_test_3",
			"another_db",
			"pg_test_456",
		}

		expectedDropped = []string{
			"pg_test_1",
			"pg_test_2",
			"pg_test_3",
			"pg_test_456",
		}
	)

	// Set up: create a rootDB with a mockPool.
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}
	defer mockPool.Close()

	// Allowing expectations to be run out of order since we drop each
	// database individually, and there's really no reason to drop them in
	// a particular order.
	mockPool.MatchExpectationsInOrder(false)

	rootDB := &rootDB{db: mockPool}
	defer rootDB.close()

	// Set up: create the factory.
	factory := &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       rootDB,
		rng:          rng,
	}
	defer factory.close()

	// Set up: mock the query to get the current databases.
	getCurrentDBsRows := pgxmock.NewRows([]string{"datname"})
	for _, name := range existingDBNames {
		getCurrentDBsRows.AddRow(name)
	}

	mockPool.
		ExpectQuery(regexp.QuoteMeta(
			`SELECT datname FROM pg_database;`,
		)).
		WillReturnRows(
			getCurrentDBsRows,
		).
		RowsWillBeClosed().
		Times(1)

	// Set up: mock the queries to drop the databases.
	for _, name := range expectedDropped {
		mockPool.
			ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
				"DROP DATABASE %q;", name,
			))).
			WillReturnResult(pgxmock.NewResult("DROP DATABASE", 1)).
			Times(1)
	}

	// Attempt to drop the test dbs.
	if err := factory.destroyAllTestDBs(ctx); err != nil {
		t.Fatalf("factory.destroyAllTestDBs(ctx) = %s; want nil", err)
	}

	// Verify the mock was called as expected.
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("mock pool has unfulfilled expectations: %s", err)
	}
}
