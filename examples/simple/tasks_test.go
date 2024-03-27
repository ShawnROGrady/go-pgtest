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

func TestCreateTaskSuccess(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		cl  = NewTestClient(t)

		// Input attributes.
		name        = "task 1"
		description = "the first task"
		priority    = db.TaskPriorityLow
	)

	// Create the assignee.
	assignee, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: "foo@example.com",
		Name:  "foo",
	})
	require.NoError(t, err, "failed to create assignee")

	// Create the task
	newTask, err := cl.CreateTask(ctx, db.CreateTaskParams{
		Name:        name,
		Description: &description,
		Priority:    priority,
		AssigneeID:  assignee.ID,
	})
	require.NoError(t, err, "failed to create task")

	// Verify the new task has the expected attributes.
	expectedNewTask := &db.Task{
		Name:        name,
		Description: &description,
		Priority:    priority,
		AssigneeID:  assignee.ID,
		Status:      db.TaskStatusPending,
	}

	if diff := cmp.Diff(
		expectedNewTask, newTask,
		cmpopts.IgnoreFields(db.Task{}, "ID", "CreatedAt", "UpdatedAt"),
	); diff != "" {
		assert.Failf(t, "Unexpected newTask\nDiff (-expected +actual):\n%s", diff)
	}

	// Verify the new task can be retrieved.
	foundTask, err := cl.GetTask(ctx, newTask.ID)
	require.NoError(t, err, "failed to retrieve new task")

	// Verify the found task is the same as the newly created task.
	expectedFoundTask := newTask
	if diff := cmp.Diff(
		expectedFoundTask, foundTask,
	); diff != "" {
		assert.Failf(t, "Unexpected foundTask\nDiff (-expected +actual):\n%s", diff)
	}
}

func TestUpdateTaskSuccess(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		cl  = NewTestClient(t)

		// Input attributes.
		name = "task 1"

		newDescription = "the first task"

		initialPriority = db.TaskPriorityLow
		newPriority     = db.TaskPriorityMedium
	)

	// Create the assignee.
	assignee, err := cl.CreateUser(ctx, db.CreateUserParams{
		Email: "foo@example.com",
		Name:  "foo",
	})
	require.NoError(t, err, "failed to create assignee")

	// Create the initial task.
	newTask, err := cl.CreateTask(ctx, db.CreateTaskParams{
		Name:       name,
		Priority:   initialPriority,
		AssigneeID: assignee.ID,
	})
	require.NoError(t, err, "failed to create task")

	// Update the task.
	var updates db.TaskUpdates
	updates.SetDescription(&newDescription)
	updates.SetPriority(newPriority)

	updatedTask, err := cl.UpdateTask(ctx, newTask.ID, updates)
	require.NoError(t, err, "failed to update task")

	// Verify the updated task has the expected attributes.
	expectedUpdatedTask := &db.Task{
		ID:          newTask.ID,
		CreatedAt:   newTask.CreatedAt,
		Name:        newTask.Name,
		Description: &newDescription,
		Priority:    newPriority,
		AssigneeID:  assignee.ID,
		Status:      db.TaskStatusPending,
	}

	if diff := cmp.Diff(
		expectedUpdatedTask, updatedTask,
		cmpopts.IgnoreFields(db.Task{}, "UpdatedAt"),
	); diff != "" {
		assert.Failf(t, "Unexpected updatedTask\nDiff (-expected +actual):\n%s", diff)
	}

	// Verify the task reflects the update.
	foundTask, err := cl.GetTask(ctx, newTask.ID)
	require.NoError(t, err, "failed to retrieve task")

	// Verify the found task is the same as the updated created task.
	expectedFoundTask := updatedTask
	if diff := cmp.Diff(
		expectedFoundTask, foundTask,
	); diff != "" {
		assert.Failf(t, "Unexpected foundTask\nDiff (-expected +actual):\n%s", diff)
	}

	// Verify the task history reflects the change.
	taskHistory, err := cl.GetTaskHistory(ctx, newTask.ID)
	require.NoError(t, err, "failed to get task history")

	expectedTaskHistory := []db.TaskHistoryEvent{{
		TaskID: newTask.ID,
		Changes: db.TaskChanges{
			Description: &db.Change[*string]{
				Before: nil,
				After:  &newDescription,
			},
			Priority: &db.Change[db.TaskPriority]{
				Before: initialPriority,
				After:  newPriority,
			},
		},
	}}

	if diff := cmp.Diff(
		expectedTaskHistory, taskHistory,
		cmpopts.IgnoreFields(db.TaskHistoryEvent{}, "ID", "RecordedAt"),
	); diff != "" {
		assert.Failf(t, "Unexpected taskHistory\nDiff (-expected +actual):\n%s", diff)
	}
}
