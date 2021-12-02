CREATE TABLE task_links ( 
	id SERIAL PRIMARY KEY,
	task_id VARCHAR(36) NOT NULL,
	uri VARCHAR(1024) NOT NULL,
	nameOffset INT,
	nameLength INT,
	created_at TIMESTAMP DEFAULT now(),
	modified_at TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id)	
);