package pgtest

import (
	"fmt"

	"github.com/ShawnROGrady/go-pgtest/pgtest/connparams"
)

type TestDB interface {
	isTestDB()
	name() string
	DataSourceName() string
}

type testDB struct {
	connparams *connparams.ConnectionParams
}

func (db *testDB) GoString() string {
	if db == nil {
		return "nil"
	}

	return fmt.Sprintf("&testDB{%q}", db.name())
}

func (db *testDB) isTestDB() {}
func (db *testDB) name() string {
	return db.connparams.DBName()
}

func (db testDB) DataSourceName() string {
	return db.connparams.URI().String()
}
