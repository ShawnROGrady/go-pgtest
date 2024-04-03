package pgtest

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultRootDBName = "postgres"

func newConfig(opts ...Option) (*config, error) {
	var connParamOpts []connparams.Option

	if host := os.Getenv("PGTEST_HOST"); host != "" {
		connParamOpts = append(connParamOpts, connparams.WithHost(host))
	}

	if portRaw := os.Getenv("PGTEST_PORT"); portRaw != "" {
		port, err := strconv.Atoi(portRaw)
		if err != nil {
			return nil, fmt.Errorf("parse PGTEST_PORT %q: %w", portRaw, err)
		}

		connParamOpts = append(connParamOpts, connparams.WithPort(port))
	}

	if u := os.Getenv("PGTEST_USER"); u != "" {
		connParamOpts = append(connParamOpts, connparams.WithUser(u))
	}

	if p := os.Getenv("PGTEST_PASSWORD"); p != "" {
		connParamOpts = append(connParamOpts, connparams.WithPassword(p))
	}

	paramFactory := func(dbName string) connparams.ConnectionParams {
		return connparams.NewWithDefaults(dbName, connParamOpts...)
	}

	var keepDatabasesForFailed bool
	if o := os.Getenv("PG_TEST_KEEP_DATABASES_FOR_FAILED"); o != "" {
		var err error
		keepDatabasesForFailed, err = strconv.ParseBool(o)
		if err != nil {
			return nil, fmt.Errorf("parse PG_TEST_KEEP_DATABASES_FOR_FAILED %q: %w", o, err)
		}
	}

	/*
		var keepExistingTestDBs bool
		if o := os.Getenv("PG_TEST_KEEP_EXISTING_TEST_DBS"); o != "" {
			var err error
			keepExistingTestDBs, err = strconv.ParseBool(o)
			if err != nil {
				return nil, fmt.Errorf("parse PG_TEST_KEEP_EXISTING_TEST_DBS %q: %w", o, err)
			}
		}
	*/

	c := &config{
		resetOp:                DropAllTables(),
		keepDatabasesForFailed: keepDatabasesForFailed,
		//keepExistingTestDBs:    keepExistingTestDBs,
		paramFactory: paramFactory,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c, nil
}

func newTestDBFactory(ctx context.Context, conf *config) (*testDBFactory, error) {
	var (
		paramFactory = conf.paramFactory
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
	conf, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	factory, err := newTestDBFactory(ctx, conf)
	if err != nil {
		return nil, err
	}

	inner := newSupervisor(conf, factory)
	s := &testSupervisor{
		inner:                  inner,
		keepDatabasesForFailed: conf.keepDatabasesForFailed,
	}

	/*
		if !s.keepExistingTestDBs {
			if err := factory.destroyAllTestDBs(ctx); err != nil {
				return nil, fmt.Errorf("destroy old test dbs: %w", err)
			}
		}
	*/

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

	conf, err := newConfig()
	if err != nil {
		t.Fatalf("load config: %s", err)
	}

	state, err := newTestDBFactory(ctx, conf)
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
