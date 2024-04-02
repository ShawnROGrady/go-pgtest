package suitebased

import (
	"context"

	"github.com/ShawnROGrady/go-pgtest/examples/common/db"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func (test *DBTestSuite) TestCreateUserSucces() {
	var (
		ctx = context.Background()

		cl = test.client

		// Input attributes.
		name  = "foo"
		email = "foo@example.com"
	)

	// Add the user.
	newUser, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  name,
	})
	test.Require().NoError(err, "failed to create user")

	// Verify the new user has the expected attributes.
	expectedNewUser := &db.User{
		Email: email,
		Name:  name,
	}

	if diff := cmp.Diff(
		expectedNewUser, newUser,
		cmpopts.IgnoreFields(db.User{}, "ID", "CreatedAt", "UpdatedAt"),
	); diff != "" {
		test.Failf("Unexpected newUser\nDiff (-expected +actual):\n%s", diff)
	}

	// Verify the new user can be retrieved.
	foundUser, err := cl.GetUser(ctx, newUser.ID)
	test.Require().NoError(err, "failed to retrieve new user")

	// Verify the found user is the same as the newly created user.
	expectedFoundUser := newUser
	if diff := cmp.Diff(
		expectedFoundUser, foundUser,
	); diff != "" {
		test.Failf("Unexpected foundUser\nDiff (-expected +actual):\n%s", diff)
	}
}

func (test *DBTestSuite) TestCreateUserWithDuplicateEmail() {
	var (
		ctx = context.Background()
		cl  = test.client

		// Input attributes.
		email = "foo@example.com"
	)

	// Add a user with the email.
	_, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  "first",
	})
	test.Require().NoError(err, "failed to create first user")

	// Attempt to add another user with the same email.
	user2, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  "second",
	})
	test.Require().Errorf(err, "unexpectedly no error crating second user; created %#v", user2)

	// Verify the expected error is returned.
	test.ErrorIs(err, &db.QueryError{Reason: db.ErrorReasonDuplicateUserEmail}, "unexpected error value returned creating second user")
}

func (test *DBTestSuite) TestCreateUserConstraintViolations() {
	var (
		ctx = context.Background()
		cl  = test.client
	)

	testCases := map[string]struct {
		params            db.CreateUserParams
		expectedErrReason db.ErrorReason
	}{
		"name_empty": {
			params:            db.CreateUserParams{Email: "foo@example.com"},
			expectedErrReason: db.ErrorReasonUserNameEmpty,
		},
		"email_empty": {
			params:            db.CreateUserParams{Name: "foo"},
			expectedErrReason: db.ErrorReasonUserEmailEmpty,
		},
	}

	for testName, tt := range testCases {
		test.Run(testName, func() {
			newUser, err := cl.CreateUser(ctx, tt.params)
			test.Require().Errorf(err, "unexpectedly no error from cl.CreateUser(ctx, %#v); got newUser = %#v", tt.params, newUser)

			test.ErrorIs(err, &db.QueryError{Reason: tt.expectedErrReason}, "unexpected error value returned from cl.CreateUser(ctx, %#v)", tt.params)
		})
	}
}
