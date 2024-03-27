-- name: CreateTask :one
INSERT INTO tasks (
	priority,
	name,
	description,
	assignee_id
)
VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id=$1;

-- name: GetTaskHistory :many
SELECT * FROM task_changes WHERE task_id = $1;

-- name: GetTaskWithHistory :one
SELECT
	task.*,
	array_agg(history.recorded_at)::TIMESTAMPTZ[] AS history_recorded_ats,
	array_agg(history.before)::JSON[] AS history_befores,
	array_agg(history.after)::JSON[] AS history_after
FROM
	tasks task
	LEFT JOIN task_changes history ON (
		history.task_id = task.id
	)
WHERE
	task.id=$1
ORDER BY
	history.id ASC
;

-- name: UpdateTask :one
UPDATE
	tasks
SET
	updated_at = now(),
	name = coalesce(sqlc.narg('name'), name),
	description = (CASE WHEN sqlc.arg('set_description')::bool
				THEN sqlc.arg('description')
				ELSE description
			END),
	priority = coalesce(sqlc.narg('priority'), priority),
	status = coalesce(sqlc.narg('status'), status),
	assignee_id = coalesce(sqlc.narg('assignee_id'), assignee_id)
WHERE
	id = $1
RETURNING *;

-- name: ListTasks :many
SELECT
	tasks.*
FROM
	tasks
WHERE
	status = coalesce(sqlc.narg('status'), status)
	AND assignee_id = coalesce(sqlc.narg('assignee_id'), assignee_id)
LIMIT $1
OFFSET $2
;

-- name: ListOpenTasksByAssignee :many
SELECT
	tasks.*
FROM
	tasks
WHERE
	assignee_id = $1
	AND status NOT IN ('done', 'obsolete')
;

-- name: ListClosedTasksByAssignee :many
SELECT
	tasks.*
FROM
	tasks
WHERE
	assignee_id = $1
	AND status IN ('done', 'obsolete')
;
