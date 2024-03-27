// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package dbqueries

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type TaskPriority string

const (
	TaskPriorityUnprioritized TaskPriority = "unprioritized"
	TaskPriorityLow           TaskPriority = "low"
	TaskPriorityMedium        TaskPriority = "medium"
	TaskPriorityHigh          TaskPriority = "high"
	TaskPriorityCritical      TaskPriority = "critical"
)

func (e *TaskPriority) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskPriority(s)
	case string:
		*e = TaskPriority(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskPriority: %T", src)
	}
	return nil
}

type NullTaskPriority struct {
	TaskPriority TaskPriority
	Valid        bool // Valid is true if TaskPriority is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTaskPriority) Scan(value interface{}) error {
	if value == nil {
		ns.TaskPriority, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TaskPriority.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTaskPriority) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TaskPriority), nil
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusValidating TaskStatus = "validating"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusObsolete   TaskStatus = "obsolete"
)

func (e *TaskStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskStatus(s)
	case string:
		*e = TaskStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskStatus: %T", src)
	}
	return nil
}

type NullTaskStatus struct {
	TaskStatus TaskStatus
	Valid      bool // Valid is true if TaskStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTaskStatus) Scan(value interface{}) error {
	if value == nil {
		ns.TaskStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TaskStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTaskStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TaskStatus), nil
}

type Task struct {
	ID          int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Description pgtype.Text
	Priority    TaskPriority
	Status      TaskStatus
	AssigneeID  int32
}

type TaskChange struct {
	ID         int32
	RecordedAt time.Time
	TaskID     int32
	Before     []byte
	After      []byte
}

type User struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Name      string
}
