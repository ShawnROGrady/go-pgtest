package simple

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/examples/common/db"
	"github.com/ShawnROGrady/go-pgtest/pgtest"
	"github.com/stretchr/testify/require"
)

var (
	pgtestSupervisor         pgtest.Supervisor
	pgtestSupervisorInitOnce sync.Once
)

func initPgTestSupervisor() error {
	var err error

	pgtestSupervisorInitOnce.Do(func() {
		pgtestSupervisor, err = pgtest.NewSupervisor(
			context.Background(),
			pgtest.WithResetOp(pgtest.DropAllTablesExcept("schema_migations")),
		)
	})

	return err
}

func NewTestClient(t testing.TB) *db.Client {
	t.Helper()
	ctx := context.Background()
	testDB := pgtestSupervisor.GetTestDB(t)

	cl, err := db.Connect(ctx, testDB.DataSourceName())
	require.NoError(t, err, "failed to connect to db")
	t.Cleanup(cl.Close)

	require.NoError(t, cl.Migrate(ctx), "failed to run migrations")
	return cl
}

func TestMain(m *testing.M) {
	var (
		ctx = context.Background()
	)

	if err := initPgTestSupervisor(); err != nil {
		log.Fatalf("initialize pgtest supervisor: %s", err)
	}
	code := m.Run()

	if err := pgtestSupervisor.Shutdown(ctx); err != nil {
		if code == 0 {
			log.Fatalf("shutdown pgtest supervisor: %s", err)
		} else {
			log.Printf("shutdown pgtest supervisor: %s", err)
		}
	}

	os.Exit(code)
}
