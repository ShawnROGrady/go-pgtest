package pgtest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
)

type supervisor struct {
	paramFactory connparams.Factory

	rootDB *rootDB
	mut    sync.Mutex
	rng    *rand.Rand
}

func (s *supervisor) randomDBName() string {
	s.mut.Lock()
	defer s.mut.Unlock()

	return fmt.Sprintf("pg_test_%d", s.rng.Int())
}

func (s *supervisor) createTestDB(ctx context.Context) (*testDB, error) {
	var (
		retryCount = 5
		err        error
	)

	for retryCount > 0 {
		dbName := s.randomDBName()

		err = s.rootDB.createDatabase(ctx, dbName)
		if err == nil {
			return &testDB{
				connparams: s.paramFactory(dbName),
			}, nil
		}

		if !errors.Is(err, &databaseAlreadyExistsWithName{name: dbName}) {
			break
		}

		retryCount--
	}

	return nil, err
}

func (s *supervisor) destroyTestDB(ctx context.Context, testDB TestDB) error {
	return s.rootDB.dropDatabase(ctx, testDB.name())
}
