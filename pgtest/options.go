package pgtest

type Option interface {
	apply(*testSupervisor)
}

type optFn func(*testSupervisor)

func (fn optFn) apply(s *testSupervisor) { fn(s) }

// WithResetOp returns an option that specifies how to reset a test db in order
// to prepare it for future tests. By default this is set to DropAllTables().
func WithResetOp(op ResetTestDBOp) Option {
	return optFn(func(s *testSupervisor) {
		s.inner.resetOp = op
	})
}

// WithKeepDatabasesForFailed returns an option which controls whether or not
// to keep test databases if a test using them fails.
func WithKeepDatabasesForFailed(v bool) Option {
	return optFn(func(s *testSupervisor) {
		s.keepDatabasesForFailed = v
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
