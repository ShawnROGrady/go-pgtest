package pgtest

import "github.com/ShawnROGrady/go-pgtest/pgtest/connparams"

// config describes the configuration for pgtest.
type config struct {
	// resetOp is the operation to reset a testDB for use in further tests.
	//
	// The default is to run DropAllTables.
	//
	// Currently this is called after retrieving a testDB from the pool,
	// however in the future it might be better to call it before releasing
	// the testDB back to the pool.
	resetOp ResetTestDBOp

	// keepDatabasesForFailed prevents a testDB from being released to the
	// pool if the test that acquired it fails. As a result, such testDBs
	// will not be re-used for future tests and will not be automatically
	// dropped.
	//
	// This is intended to be used for debugging failed tests.
	keepDatabasesForFailed bool

	// keepExistingTestDBs prevents existing testDBs from being dropped by
	// the supervisor. By default, the supervisor immediately drops any old
	// test databases to clean up any test databases that weren't
	// previously dropped (such as if keepDatabasesForFailed was specified
	// or if the supervisor didn't shutdown correctly).
	keepExistingTestDBs bool

	paramFactory connparams.Factory
}
