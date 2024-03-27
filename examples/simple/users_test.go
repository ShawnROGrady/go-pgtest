package simple

import (
	"context"
	"testing"

	"github.com/ShawnROGrady/go-pgtest/examples/common/db"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserSucces(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		cl  = NewTestClient(t)

		// Input attributes.
		name  = "foo"
		email = "foo@example.com"
	)

	// Add the user.
	newUser, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  name,
	})
	require.NoError(t, err, "failed to create user")

	// Verify the new user has the expected attributes.
	expectedNewUser := &db.User{
		Email: email,
		Name:  name,
	}

	if diff := cmp.Diff(
		expectedNewUser, newUser,
		cmpopts.IgnoreFields(db.User{}, "ID", "CreatedAt", "UpdatedAt"),
	); diff != "" {
		assert.Failf(t, "Unexpected newUser\nDiff (-expected +actual):\n%s", diff)
	}

	// Verify the new user can be retrieved.
	foundUser, err := cl.GetUser(ctx, newUser.ID)
	require.NoError(t, err, "failed to retrieve new user")

	// Verify the found user is the same as the newly created user.
	expectedFoundUser := newUser
	if diff := cmp.Diff(
		expectedFoundUser, foundUser,
	); diff != "" {
		assert.Failf(t, "Unexpected foundUser\nDiff (-expected +actual):\n%s", diff)
	}
}

func TestCreateUserConstraintViolations(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		cl  = NewTestClient(t)
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
		tt := tt
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			newUser, err := cl.CreateUser(ctx, tt.params)
			require.Errorf(t, err, "unexpectedly no error from cl.CreateUser(ctx, %#v); got newUser = %#v", tt.params, newUser)

			assert.ErrorIs(t, err, &db.QueryError{Reason: tt.expectedErrReason}, "unexpected error value returned from cl.CreateUser(ctx, %#v)", tt.params)
		})
	}
}
