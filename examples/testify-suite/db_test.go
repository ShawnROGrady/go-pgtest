package suitebased

import (
	"context"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/examples/common/db"
	"github.com/ShawnROGrady/go-pgtest/pgtest"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	supervisor pgtest.Supervisor

	// Per-test.
	client *db.Client
}

func (test *DBTestSuite) SetupSuite() {
	pgtestSupervisor, err := pgtest.NewSupervisor(
		context.Background(),
		pgtest.WithResetOp(pgtest.TruncateAllTablesExcept("schema_migrations")),
	)
	test.Require().NoError(err, "unexpected error initializing supervisor")
	test.supervisor = pgtestSupervisor
}

func (test *DBTestSuite) TearDownSuite() {
	ctx := context.Background()
	test.Require().NoError(test.supervisor.Shutdown(ctx), "unexpected error shutting down supervisor")
}

func (test *DBTestSuite) SetupTest() {
	ctx := context.Background()
	testDB := test.supervisor.GetTestDB(test.T())

	cl, err := db.Connect(ctx, testDB.DataSourceName())
	test.Require().NoError(err, "failed to connect to db")

	test.Require().NoError(cl.Migrate(ctx), "failed to run migrations")

	test.client = cl
}

func (test *DBTestSuite) TearDownTest() {
	test.client.Close()
}

func TestDB(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}
