CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'validating', 'done', 'obsolete');
CREATE TYPE task_priority AS ENUM ('unprioritized', 'low', 'medium', 'high', 'critical');

CREATE TABLE tasks (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	name TEXT NOT NULL,
	description TEXT DEFAULT NULL,
	priority task_priority NOT NULL DEFAULT 'unprioritized',
	status task_status NOT NULL DEFAULT 'pending',
	assignee_id INTEGER NOT NULL REFERENCES users (id),

	CONSTRAINT tasks_name_not_empty CHECK (name <> ''),
	CONSTRAINT tasks_name_unique_per_assignee UNIQUE (assignee_id, name),
	CONSTRAINT tasks_description_not_empty CHECK (description <> '')
);

CREATE FUNCTION tasks_status_update_allowed() RETURNS trigger AS $tasks_status_update_allowed$
	BEGIN
		-- Allow any task to be marked as 'obsolete'.
		IF NEW.status = 'obsolete' THEN
			RETURN NEW;
		END IF;

		-- Only allow single state transition per-update (besides
		-- marking 'obsolete'). So you can move an 'in_progress' task
		-- to 'validating' but not 'done' for example.
		IF (OLD.status = 'pending' AND NEW.status <> 'in_progress')
			OR (OLD.status = 'in_progress' AND NEW.status NOT IN ('pending', 'validating'))
			OR (OLD.status = 'validating' AND NEW.status NOT IN ('in_progress', 'done'))
			OR (OLD.status = 'done' AND NEW.status NOT IN ('pending', 'validating'))
			OR (OLD.status = 'obsolete' AND NEW.status <> 'pending')
		THEN
			RAISE EXCEPTION integrity_constraint_violation USING
				MESSAGE = 'Invalid task status update',
				DETAIL = 'Cannot transition a ' || OLD.status || ' task to ' || NEW.status
			;
		END IF;

		RETURN NEW;
	END;
$tasks_status_update_allowed$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER tasks_status_update_allowed
	AFTER UPDATE OF status ON tasks
	FOR EACH ROW
		WHEN (OLD.status IS DISTINCT FROM NEW.status)
		EXECUTE FUNCTION tasks_status_update_allowed()
;
