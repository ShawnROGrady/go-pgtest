package pgtest

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultRootDBName = "postgres"

func pgTestParamFactory() connparams.Factory {
	// TODO: let this be overridden.
	return connparams.DefaultFactory()
}

func createPgTestSupervisor(ctx context.Context, t testing.TB) *supervisor {
	var (
		paramFactory = pgTestParamFactory()
		rootDBName   = defaultRootDBName
	)

	rootDBPool, err := pgxpool.New(ctx, paramFactory(rootDBName).URI().String())
	if err != nil {
		t.Fatalf("open %q: %s", rootDBName, err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &supervisor{
		paramFactory: paramFactory,
		rootDB:       &rootDB{db: rootDBPool},
		rng:          rng,
	}
}

func NewTestDB(t testing.TB) TestDB {
	ctx := context.Background()

	supervisor := createPgTestSupervisor(ctx, t)
	testDB, err := supervisor.createTestDB(ctx)
	if err != nil {
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		if err := supervisor.destroyTestDB(ctx, testDB); err != nil {
			t.Fatalf("destroy test db: %v", err)
		}
	})

	return testDB
}
