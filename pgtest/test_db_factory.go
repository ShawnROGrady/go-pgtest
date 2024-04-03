package pgtest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"sync"
)

const testDBNamePrefix = "pg_test_"

type testDBFactory struct {
	paramFactory connparamsFactory

	rootDB *rootDB
	mut    sync.Mutex
	rng    *rand.Rand
}

func (s *testDBFactory) randomDBName() string {
	s.mut.Lock()
	defer s.mut.Unlock()

	return fmt.Sprintf("%s%d", testDBNamePrefix, s.rng.Int())
}

func (s *testDBFactory) createTestDB(ctx context.Context) (TestDB, error) {
	var (
		retryCount = 5
		err        error
	)

	for retryCount > 0 {
		dbName := s.randomDBName()

		err = s.rootDB.createDatabase(ctx, dbName)
		if err == nil {
			ps := s.paramFactory(dbName)
			return &testDB{
				connparams: ps,
			}, nil
		}

		if !errors.Is(err, &databaseAlreadyExistsWithName{name: dbName}) {
			break
		}

		retryCount--
	}

	return nil, err
}

func (s *testDBFactory) destroyTestDB(ctx context.Context, testDB TestDB) error {
	return s.rootDB.dropDatabase(ctx, testDB.name())
}

func (s *testDBFactory) close() {
	s.rootDB.close()
}

func (s *testDBFactory) destroyAllTestDBs(ctx context.Context) error {
	dbNames, err := s.rootDB.getAllDatabases(ctx)
	if err != nil {
		return fmt.Errorf("get all databases: %w", err)
	}

	toDrop := slices.DeleteFunc(dbNames, func(name string) bool {
		return !strings.HasPrefix(name, testDBNamePrefix)
	})

	for _, dbName := range toDrop {
		if err := s.rootDB.dropDatabase(ctx, dbName); err != nil {
			return fmt.Errorf("drop database %q: %w", dbName, err)
		}
	}

	return nil
}
