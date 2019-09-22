CREATE TABLE tags (
	id SERIAL PRIMARY KEY,
	name VARCHAR(1024) NOT NULL,
	system BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT now(),
	modified_at TIMESTAMP
);

CREATE TABLE task_tags (
	task_id VARCHAR(36) NOT NULL,
	tag_id INT NOT NULL,
	created_at TIMESTAMP DEFAULT now(),
	CONSTRAINT pk_tasktags PRIMARY KEY (task_id, tag_id),
	FOREIGN KEY (task_id) REFERENCES tasks(id),
	FOREIGN KEY (tag_id) REFERENCES tags(id)
);

INSERT INTO tags ( name, system ) 
VALUES ( 'today' , true ), 
       ( 'thisweek', true ), 
       ( 'dontforget', true );
ALTER SEQUENCE tags_id_seq RESTART WITH 1000;

INSERT INTO task_tags ( task_id, tag_id )
SELECT id, 1 FROM tasks WHERE today IS true;

INSERT INTO task_tags ( task_id, tag_id )
SELECT id, 2 FROM tasks WHERE thisweek IS true;

ALTER TABLE tasks
	DROP COLUMN today,
	DROP COLUMN thisweek;

SELECT t.id, t.name, t.state,
       count(tt.tag_id = 0) > 0 as today
FROM   tasks t
JOIN   task_tags tt on tt.task_id = t.id
GROUP BY t.id