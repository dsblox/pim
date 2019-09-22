CREATE TABLE migrations (
	version_applied INT NOT NULL,
	file_applied VARCHAR(1024),
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE tasks ( 
	id CHAR(36) PRIMARY KEY,
	name VARCHAR(1024) NOT NULL,
	state INT NOT NULL,
	target_start_time TIMESTAMP,
	actual_start_time TIMESTAMP,
	actual_completion_time TIMESTAMP,
	estimate_minutes INT,
	today BOOLEAN,
	thisweek BOOLEAN,
	created_at TIMESTAMP DEFAULT now(),
	modified_at TIMESTAMP
);

CREATE TABLE task_parents (
	parent_id CHAR(36) NOT NULL,
	child_id CHAR(36) NOT NULL,
	created_at TIMESTAMP DEFAULT now(),
	modified_at TIMESTAMP,
	CONSTRAINT pk_parents PRIMARY KEY (parent_id,child_id),
	FOREIGN KEY (parent_id) REFERENCES tasks(id),
	FOREIGN KEY (child_id) REFERENCES tasks(id) 
);

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
