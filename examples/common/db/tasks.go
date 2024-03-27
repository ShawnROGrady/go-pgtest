package db

import (
	"context"
)

func (q *pgxQueries) CreateTask(ctx context.Context, params CreateTaskParams) (*Task, error) {
	t, err := q.inner.CreateTask(ctx, params.toInternal())
	if err != nil {
		return nil, err
	}

	return taskFromQuery(t), nil
}

func (cl *Client) CreateTask(ctx context.Context, params CreateTaskParams) (*Task, error) {
	return cl.queries().CreateTask(ctx, params)
}

func (q *pgxQueries) GetTask(ctx context.Context, id int32) (*Task, error) {
	t, err := q.inner.GetTask(ctx, id)
	if err != nil {
		var errReason ErrorReason
		if isNoRowsErr(err) {
			errReason = ErrorReasonTaskNotFound
		}

		return nil, queryError(errReason, err)
	}

	return taskFromQuery(t), nil
}

func (cl *Client) GetTask(ctx context.Context, id int32) (*Task, error) {
	return cl.queries().GetTask(ctx, id)
}

func (q *pgxQueries) GetTaskHistory(ctx context.Context, id int32) ([]TaskHistoryEvent, error) {
	rawChanges, err := q.inner.GetTaskHistory(ctx, id)
	if err != nil {
		return nil, err
	}

	events := make([]TaskHistoryEvent, len(rawChanges))
	for i, rawChange := range rawChanges {
		event, err := parseTaskChange(rawChange)
		if err != nil {
			return nil, &QueryError{
				Cause: err,
			}
		}

		events[i] = event
	}

	return events, nil
}

func (cl *Client) GetTaskHistory(ctx context.Context, id int32) ([]TaskHistoryEvent, error) {
	return cl.queries().GetTaskHistory(ctx, id)
}

func (q *pgxQueries) UpdateTask(ctx context.Context, id int32, updates TaskUpdates) (*Task, error) {
	t, err := q.inner.UpdateTask(ctx, updates.toInternal(id))
	if err != nil {
		return nil, err
	}

	return taskFromQuery(t), nil
}

func (cl *Client) UpdateTask(ctx context.Context, id int32, updates TaskUpdates) (*Task, error) {
	return cl.queries().UpdateTask(ctx, id, updates)
}
