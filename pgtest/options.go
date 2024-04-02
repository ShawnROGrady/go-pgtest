package pgtest

type Option interface {
	apply(*config)
}

type optFn func(*config)

func (fn optFn) apply(c *config) { fn(c) }

// WithResetOp returns an option that specifies how to reset a test db in order
// to prepare it for future tests. By default this is set to DropAllTables().
func WithResetOp(op ResetTestDBOp) Option {
	return optFn(func(c *config) {
		c.resetOp = op
	})
}

// WithKeepDatabasesForFailed returns an option which controls whether or not
// to keep test databases if a test using them fails.
func WithKeepDatabasesForFailed(v bool) Option {
	return optFn(func(c *config) {
		c.keepDatabasesForFailed = v
	})
}

// KeepDatabasesForFailed returns an option which prevents test databases from
// being dropped if the test they are used in fails. Such test databases will
// also not be re-used in future tests. By default test databases will be
// re-used regardless of whether or not tests fail, and will be automatically
// dropped when the supervisor is shutdown.
//
// This option is primarily useful for debugging. Generally it is expected to
// combine this option with the '-run' flag when running tests in order to
// inspect the state of the database for a specific failing test.
func KeepDatabasesForFailed() Option {
	return WithKeepDatabasesForFailed(true)
}

// WithKeepExistingTestDBs returns an option which controls whether or not to
// keep old test databases that were not previously dropped by the supervisor.
func WithKeepExistingTestDBs(v bool) Option {
	return optFn(func(c *config) {
		c.keepExistingTestDBs = v
	})
}

// KeepExistingTestDBs returns an option which prevents existing test databases
// from being dropped by the supervisor. Normally test databases are
// automatically dropped by the suervisor, but if the supervisor is not
// shutdown correctly (such as if a test panics) or if the test databases are
// kept for failed tests, these test databases can start to accumulate. To help
// keep this in check, the supervisor will also try to automatically drop any
// old test databases. If this is not desired for any reason though,
// KeepExistingTestDBs can be used to prevent these old test databases from
// being dropped.
func KeepExistingTestDBs() Option {
	return WithKeepExistingTestDBs(true)
}
