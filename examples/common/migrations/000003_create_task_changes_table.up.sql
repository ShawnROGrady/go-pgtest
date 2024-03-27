CREATE TABLE task_changes (
	id SERIAL PRIMARY KEY,
	recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	task_id INTEGER NOT NULL REFERENCES tasks (id),
	before JSONB NOT NULL,
	after JSONB NOT NULL
);

CREATE FUNCTION record_task_change() RETURNS trigger AS $record_task_change$
	DECLARE
		before JSONB;
		after JSONB;
		k TEXT;
	BEGIN
		before := to_jsonb(OLD) - '{created_at,updated_at}'::text[];
		after := to_jsonb(NEW) - '{created_at,updated_at}'::text[];

		FOR k IN
			SELECT * FROM jsonb_object_keys(before)
		LOOP
			IF (before->k = after->k) THEN
				before := before - k;
				after := after - k;
			END IF;
		END LOOP;

		IF (before = '{}'::JSONB OR after = '{}'::JSONB) THEN
			RETURN NEW;
		END IF;
		
		INSERT INTO task_changes (
			task_id,
			before,
			after
		)
		VALUES (
			OLD.id,
			before,
			after
		);

		RETURN NEW;
	END;
$record_task_change$ LANGUAGE plpgsql;

CREATE TRIGGER record_task_change
	AFTER UPDATE ON tasks
	FOR EACH ROW EXECUTE FUNCTION record_task_change()
;
