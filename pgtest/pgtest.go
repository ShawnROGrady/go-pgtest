package pgtest

import (
	"context"
	"fmt"
	"log"
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

// A Supervisor is used to manage databases for use in a test suite.
type Supervisor interface {
	// GetTestDB returns a db for use in testing.
	GetTestDB(t testing.TB) TestDB

	// Shutdown shuts down the supervisor, dropping any test databases it
	// owns.
	Shutdown(ctx context.Context) error
}

type testSupervisor struct {
	inner                  *supervisor
	keepDatabasesForFailed bool
}

// GetTestDB returns a db for use in testing.
func (s *testSupervisor) GetTestDB(t testing.TB) TestDB {
	ctx := context.Background()
	dbResource, err := s.inner.getTestDB(ctx)
	if err != nil {
		t.Fatalf("get test db: %s", err)
	}

	t.Cleanup(func() {
		if t.Failed() && s.keepDatabasesForFailed {
			dbResource.Hijack()
			t.Logf("keeping test db: %s", dbResource.Data().name())
			return
		}

		dbResource.Release()
	})
	return dbResource.Data()
}

// Shutdown shuts down the supervisor, dropping any test databases it owns.
func (s *testSupervisor) Shutdown(ctx context.Context) error {
	return s.inner.shutdown(ctx)
}

// NewSupervisor returns a new supervisor, which maintains a pool of test
// databases for use in testing.
func NewSupervisor(ctx context.Context, opts ...Option) (Supervisor, error) {
	factory, err := newTestDBFactory(ctx)
	if err != nil {
		return nil, err
	}

	inner := newSupervisor(factory)
	s := &testSupervisor{inner: inner}

	for _, opt := range opts {
		opt.apply(s)
	}

	return s, nil
}

func RunMain(ctx context.Context, m *testing.M, supervisor Supervisor) (code int) {
	defer func() {
		if err := supervisor.Shutdown(ctx); err != nil {
			log.Printf("ERROR: pgtest: shutdown pgtest supervisor: %s", err)
			if code == 0 {
				code = 11
			}
		}
	}()

	return m.Run()
}

// NewTestDB returns a brand new TestDB that is dropped at the end of the test.
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
