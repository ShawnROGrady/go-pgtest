package pgtest

import (
	"context"
	"fmt"

	"github.com/ShawnROGrady/go-pgtest/pgtest/internal/pool"
)

type supervisor struct {
	factory *testDBFactory
	pool    *pool.Pool[TestDB]
	resetOp ResetTestDBOp
}

func newSupervisor(factory *testDBFactory, opts ...Option) *supervisor {
	resourceConf := &pool.ResourceConf[TestDB]{
		Create: func(ctx context.Context) (TestDB, error) {
			return factory.createTestDB(ctx)
		},
		Destroy: func(testDB TestDB) error {
			return factory.destroyTestDB(context.Background(), testDB)
		},
	}
	pool := pool.New[TestDB](resourceConf)

	s := &supervisor{
		factory: factory,
		pool:    pool,
		resetOp: DropAllTables(),
	}

	for _, opt := range opts {
		opt.apply(s)
	}

	return s
}

func (s *supervisor) shutdown(ctx context.Context) error {
	if err := s.pool.Close(ctx); err != nil {
		s.factory.close()
		return err
	}

	s.factory.close()
	return nil
}

func (s *supervisor) getTestDB(ctx context.Context) (*pool.Resource[TestDB], error) {
	db, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire: %w", err)
	}

	if s.resetOp != nil {
		if err := runResetTestDBOp(ctx, s.resetOp, db.Data()); err != nil {
			return nil, fmt.Errorf("reset test db: %w", err)
		}
	}

	return db, nil
}
