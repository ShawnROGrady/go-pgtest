package pgtest

type Option interface {
	apply(*supervisor)
}

type optFn func(*supervisor)

func (fn optFn) apply(s *supervisor) { fn(s) }

// WithResetOp returns an option that specifies how to reset a test db in order
// to prepare it for future tests. By default this is set to DropAllTables().
func WithResetOp(op ResetTestDBOp) Option {
	return optFn(func(s *supervisor) {
		s.resetOp = op
	})
}
