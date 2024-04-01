package db

import "context"

func (q *pgxQueries) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	u, err := q.inner.CreateUser(ctx, params)
	if err != nil {
		var errReason ErrorReason

		if constraintViolation, ok := extractConstraintViolation(err); ok {
			switch constraintViolation.name {
			case "users_email_not_empty":
				errReason = ErrorReasonUserEmailEmpty
			case "users_name_not_empty":
				errReason = ErrorReasonUserNameEmpty
			case "users_email_key":
				errReason = ErrorReasonDuplicateUserEmail
			}
		}

		return nil, queryError(errReason, err)
	}

	return userFromQuery(u), nil
}

func (cl *Client) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	return cl.queries().CreateUser(ctx, params)
}

func (q *pgxQueries) GetUser(ctx context.Context, id int32) (*User, error) {
	u, err := q.inner.GetUser(ctx, id)
	if err != nil {
		var errReason ErrorReason
		if isNoRowsErr(err) {
			errReason = ErrorReasonUserNotFound
		}

		return nil, queryError(errReason, err)
	}

	return userFromQuery(u), nil
}

func (cl *Client) GetUser(ctx context.Context, id int32) (*User, error) {
	return cl.queries().GetUser(ctx, id)
}
