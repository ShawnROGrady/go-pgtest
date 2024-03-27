package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ShawnROGrady/go-pgtest/examples/common/dbqueries"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserParams = dbqueries.CreateUserParams

type User struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Name      string
}

func userFromQuery(u dbqueries.User) *User {
	return &User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Name:      u.Name,
	}
}

type TaskStatus = dbqueries.TaskStatus

const (
	TaskStatusPending    = dbqueries.TaskStatusPending
	TaskStatusInProgress = dbqueries.TaskStatusInProgress
	TaskStatusValidating = dbqueries.TaskStatusValidating
	TaskStatusDone       = dbqueries.TaskStatusDone
	TaskStatusObsolete   = dbqueries.TaskStatusObsolete
)

type TaskPriority = dbqueries.TaskPriority

const (
	TaskPriorityUnprioritized = dbqueries.TaskPriorityUnprioritized
	TaskPriorityLow           = dbqueries.TaskPriorityLow
	TaskPriorityMedium        = dbqueries.TaskPriorityMedium
	TaskPriorityHigh          = dbqueries.TaskPriorityHigh
	TaskPriorityCritical      = dbqueries.TaskPriorityCritical
)

type CreateTaskParams struct {
	Priority    TaskPriority
	Name        string
	Description *string
	AssigneeID  int32
}

func (p *CreateTaskParams) toInternal() dbqueries.CreateTaskParams {
	var description pgtype.Text
	if p.Description != nil {
		description.Valid = true
		description.String = *p.Description
	}

	return dbqueries.CreateTaskParams{
		Priority:    p.Priority,
		Name:        p.Name,
		Description: description,
		AssigneeID:  p.AssigneeID,
	}
}

type Task struct {
	ID          int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Description *string
	Priority    TaskPriority
	Status      TaskStatus
	AssigneeID  int32
}

func taskFromQuery(task dbqueries.Task) *Task {
	var description *string
	if task.Description.Valid {
		description = &task.Description.String
	}

	return &Task{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Name:        task.Name,
		Description: description,
		Priority:    task.Priority,
		Status:      task.Status,
		AssigneeID:  task.AssigneeID,
	}
}

type TaskUpdates struct {
	setName bool
	name    string

	setDescription bool
	description    pgtype.Text

	setPriority bool
	priority    TaskPriority

	setStatus bool
	status    TaskStatus

	setAssigneeID bool
	assigneeID    int32
}

func (u *TaskUpdates) SetName(name string) {
	u.setName = true
	u.name = name
}

func (u *TaskUpdates) SetDescription(description *string) {
	u.setDescription = true
	if description != nil {
		u.description.Valid = true
		u.description.String = *description
	}
}

func (u *TaskUpdates) SetPriority(priority TaskPriority) {
	u.setPriority = true
	u.priority = priority
}

func (u *TaskUpdates) SetStatus(status TaskStatus) {
	u.setStatus = true
	u.status = status
}

func (u *TaskUpdates) SetAssigneeID(assigneeID int32) {
	u.setAssigneeID = true
	u.assigneeID = assigneeID
}

func (u *TaskUpdates) apply(params *dbqueries.UpdateTaskParams) {
	if u.setName {
		params.Name = pgtype.Text{String: u.name, Valid: true}
	}

	if u.setDescription {
		params.SetDescription = true
		params.Description = u.description
	}

	if u.setPriority {
		params.Priority = dbqueries.NullTaskPriority{TaskPriority: u.priority, Valid: true}
	}

	if u.setStatus {
		params.Status = dbqueries.NullTaskStatus{TaskStatus: u.status, Valid: true}
	}

	if u.setAssigneeID {
		params.AssigneeID = pgtype.Int4{Int32: u.assigneeID, Valid: true}
	}
}

func (u *TaskUpdates) toInternal(taskID int32) dbqueries.UpdateTaskParams {
	params := dbqueries.UpdateTaskParams{ID: taskID}
	u.apply(&params)
	return params
}

type maybeDefined[T any] struct {
	data    T
	defined bool
}

func (u *maybeDefined[T]) UnmarshalJSON(b []byte) error {
	u.defined = true
	return json.Unmarshal(b, &u.data)
}

type taskChangeContent struct {
	Name        maybeDefined[string]       `json:"name"`
	Description maybeDefined[*string]      `json:"description"`
	Priority    maybeDefined[TaskPriority] `json:"priority"`
	Status      maybeDefined[TaskStatus]   `json:"status"`
	AssigneeID  maybeDefined[int32]        `json:"assignee_id"`
}

type Change[T any] struct {
	Before T
	After  T
}

type TaskChanges struct {
	Name        *Change[string]
	Description *Change[*string]
	Priority    *Change[TaskPriority]
	Status      *Change[TaskStatus]
	AssigneeID  *Change[int32]
}

func taskChangesFromChangeContent(before, after taskChangeContent) TaskChanges {
	var changes TaskChanges

	if before.Name.defined {
		changes.Name = &Change[string]{Before: before.Name.data, After: after.Name.data}
	}

	if before.Description.defined {
		changes.Description = &Change[*string]{Before: before.Description.data, After: after.Description.data}
	}

	if before.Priority.defined {
		changes.Priority = &Change[TaskPriority]{Before: before.Priority.data, After: after.Priority.data}
	}

	if before.Status.defined {
		changes.Status = &Change[TaskStatus]{Before: before.Status.data, After: after.Status.data}
	}

	if before.AssigneeID.defined {
		changes.AssigneeID = &Change[int32]{Before: before.AssigneeID.data, After: after.AssigneeID.data}
	}

	return changes
}

type TaskHistoryEvent struct {
	ID         int32
	RecordedAt time.Time
	TaskID     int32
	Changes    TaskChanges
}

func parseTaskChange(rawChange dbqueries.TaskChange) (TaskHistoryEvent, error) {
	var contentBefore taskChangeContent
	if err := json.Unmarshal(rawChange.Before, &contentBefore); err != nil {
		return TaskHistoryEvent{}, fmt.Errorf("parse 'before' content for task change %d: %w", rawChange.ID, err)
	}

	var contentAfter taskChangeContent
	if err := json.Unmarshal(rawChange.After, &contentAfter); err != nil {
		return TaskHistoryEvent{}, fmt.Errorf("parse 'before' content for task change %d: %w", rawChange.ID, err)
	}

	return TaskHistoryEvent{
		ID:         rawChange.ID,
		RecordedAt: rawChange.RecordedAt,
		TaskID:     rawChange.TaskID,
		Changes:    taskChangesFromChangeContent(contentBefore, contentAfter),
	}, nil
}
