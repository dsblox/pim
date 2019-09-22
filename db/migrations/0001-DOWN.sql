DROP TABLE migrations;

ALTER TABLE tasks
	DROP COLUMN today;

ALTER TABLE tasks
	ALTER created_at DROP DEFAULT;
