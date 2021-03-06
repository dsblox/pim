CREATE TABLE tasks ( 
	id CHAR(36) PRIMARY KEY,
	name VARCHAR(1024) NOT NULL,
	state INT NOT NULL,
	target_start_time TIMESTAMP,
	actual_start_time TIMESTAMP,
	actual_completion_time TIMESTAMP,
	estimate_minutes INT,
	created_at TIMESTAMP,
	modified_at TIMESTAMP
);

CREATE TABLE task_parents (
	parent_id CHAR(36) NOT NULL,
	child_id CHAR(36) NOT NULL,
	created_at TIMESTAMP,
	modified_at TIMESTAMP,
	CONSTRAINT pk_parents PRIMARY KEY (parent_id,child_id),
	FOREIGN KEY (parent_id) REFERENCES tasks(id),
	FOREIGN KEY (child_id) REFERENCES tasks(id) 
);

