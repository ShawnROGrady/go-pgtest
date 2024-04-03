package pgtest

import (
	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
)

type TestDB interface {
	isTestDB()

	name() string
	Name() string

	DataSourceName() string
}

type testDB struct {
	connparams connparams.ConnectionParams
}

func (db *testDB) isTestDB() {}
func (db *testDB) name() string {
	return db.connparams.DBName()
}
func (db *testDB) Name() string { return db.name() }

func (db *testDB) DataSourceName() string {
	return db.connparams.URI().String()
}
