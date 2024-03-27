package simple

import (
	"context"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/examples/common/db"
	"github.com/ShawnROGrady/go-pgtest/pgtest"
	"github.com/stretchr/testify/require"
)

func NewTestClient(t testing.TB) *db.Client {
	t.Helper()
	ctx := context.Background()
	testDB := pgtest.NewTestDB(t)

	cl, err := db.Connect(ctx, testDB.DataSourceName())
	require.NoError(t, err, "failed to connect to db")
	t.Cleanup(cl.Close)

	require.NoError(t, cl.Migrate(ctx), "failed to run migrations")
	return cl
}
