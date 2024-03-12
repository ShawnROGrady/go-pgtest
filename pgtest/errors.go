package pgtest

import "fmt"

type databaseAlreadyExistsWithName struct {
	name  string
	cause error
}

func (e *databaseAlreadyExistsWithName) Error() string {
	msg := fmt.Sprintf("database already exists with name=%q", e.name)
	if e.cause != nil {
		msg += ": " + e.cause.Error()
	}

	return msg
}

func (e *databaseAlreadyExistsWithName) Unwrap() error {
	return e.cause
}

func (e *databaseAlreadyExistsWithName) Is(target error) bool {
	other, ok := target.(*databaseAlreadyExistsWithName)
	if !ok {
		return false
	}

	if other.name != "" && e.name != other.name {
		return false
	}

	return true
}
