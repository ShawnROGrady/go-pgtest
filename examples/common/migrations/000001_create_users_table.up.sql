CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	email TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,

	CONSTRAINT users_email_not_empty CHECK (email <> ''),
	CONSTRAINT users_name_not_empty CHECK (name <> '')
);
