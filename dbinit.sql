CREATE TABLE tasks (
	id SERIAL PRIMARY KEY,
	name VARCHAR(1024) NOT NULL,
	state INT NOT NULL,
	created_at DATE,
	modified_at DATE
);

CREATE TABLE task_parents (
	parent_id INT NOT NULL,
	child_id INT NOT NULL,
	created_at DATE,
	modified_at DATE,
	CONSTRAINT pk_parents PRIMARY KEY (parent_id,child_id),
	FOREIGN KEY (parent_id) REFERENCES tasks(id),
	FOREIGN KEY (child_id) REFERENCES tasks(id)
);