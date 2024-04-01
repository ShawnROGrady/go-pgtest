package pgtest

import (
	"context"
	"fmt"
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

func newTestDBFactory(ctx context.Context) (*testDBFactory, error) {
	var (
		paramFactory = pgTestParamFactory()
		rootDBName   = defaultRootDBName
	)

	rootDBPool, err := pgxpool.New(ctx, paramFactory(rootDBName).URI().String())
	if err != nil {
		return nil, fmt.Errorf("open %q: %s", rootDBName, err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &testDBFactory{
		paramFactory: paramFactory,
		rootDB:       &rootDB{db: rootDBPool},
		rng:          rng,
	}, nil
}

type Supervisor interface {
	GetTestDB(t testing.TB) TestDB
	Shutdown(ctx context.Context) error
}

type testSupervisor struct {
	inner *supervisor
}

func (s *testSupervisor) GetTestDB(t testing.TB) TestDB {
	ctx := context.Background()
	dbResource, err := s.inner.getTestDB(ctx)
	if err != nil {
		t.Fatalf("get test db: %s", err)
	}

	t.Cleanup(func() {
		dbResource.Release()
	})
	return dbResource.Data()
}

func (s *testSupervisor) Shutdown(ctx context.Context) error {
	return s.inner.shutdown(ctx)
}

func NewSupervisor(ctx context.Context, opts ...Option) (Supervisor, error) {
	factory, err := newTestDBFactory(ctx)
	if err != nil {
		return nil, err
	}

	inner := newSupervisor(factory, opts...)
	return &testSupervisor{inner: inner}, nil
}

func NewTestDB(t testing.TB) TestDB {
	ctx := context.Background()

	state, err := newTestDBFactory(ctx)
	if err != nil {
		t.Fatalf("create supervisor state: %s", err)
	}

	testDB, err := state.createTestDB(ctx)
	if err != nil {
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		if err := state.destroyTestDB(ctx, testDB); err != nil {
			t.Fatalf("destroy test db: %v", err)
		}
		state.close()
	})

	return testDB
}
