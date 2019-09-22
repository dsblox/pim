ALTER TABLE tasks
	ADD COLUMN today BOOLEAN,
	ADD COLUMN thisweek BOOLEAN;

UPDATE tasks AS t SET today    = TRUE FROM task_tags AS tt WHERE tt.task_id = t.id AND tt.tag_id = 1;
UPDATE tasks AS t SET thisweek = TRUE FROM task_tags AS tt WHERE tt.task_id = t.id AND tt.tag_id = 2;

DROP TABLE task_tags;
DROP TABLE tags;