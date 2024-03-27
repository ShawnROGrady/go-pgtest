package db

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorReason uint

// The possible values for ErrorReason.
const (
	ErrorReasonUnknown ErrorReason = iota
	ErrorReasonUserNotFound
	ErrorReasonUserEmailEmpty
	ErrorReasonUserNameEmpty
	ErrorReasonTaskNotFound
	ErrorReasonTaskNameEmpty
	ErrorReasonTaskNameUniquePerAssignee
	ErrorReasonTaskDescriptionEmpty
	ErrorReasonInvalidTaskStatusUpdate
)

func (e ErrorReason) String() string {
	//exhaustive:enforce
	switch e {
	case ErrorReasonUserNotFound:
		return "user not found"
	case ErrorReasonUserEmailEmpty:
		return "user email empty"
	case ErrorReasonUserNameEmpty:
		return "user name empty"
	case ErrorReasonTaskNotFound:
		return "task not found"
	case ErrorReasonTaskNameEmpty:
		return "task name empty"
	case ErrorReasonTaskNameUniquePerAssignee:
		return "duplicate task name for assignee"
	case ErrorReasonTaskDescriptionEmpty:
		return "task description empty"
	case ErrorReasonInvalidTaskStatusUpdate:
		return "invalid task status update"
	case ErrorReasonUnknown:
		fallthrough
	default:
		return "unknown"
	}
}

type QueryError struct {
	Reason ErrorReason
	Cause  error
}

func (e *QueryError) Unwrap() error {
	return e.Cause
}

func (e *QueryError) Error() string {
	msg := e.Reason.String()
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}

	return msg
}

func (e *QueryError) Is(target error) bool {
	if other, ok := target.(*QueryError); ok {
		return other.Reason == e.Reason
	}
	return false
}

func queryError(reason ErrorReason, cause error) error {
	if reason == ErrorReasonUnknown {
		return cause
	}

	return &QueryError{
		Cause:  cause,
		Reason: reason,
	}
}

type constraintViolation struct {
	code   string
	name   string
	detail string
}

func extractConstraintViolation(err error) (*constraintViolation, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil, false
	}

	if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return nil, false
	}

	return &constraintViolation{
		code:   pgErr.Code,
		name:   pgErr.ConstraintName,
		detail: pgErr.Detail,
	}, true
}

func isNoRowsErr(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}
