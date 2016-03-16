CREATE TABLE tasks (
	id SERIAL PRIMARY KEY,
	name VARCHAR(1024) NOT NULL,
	state INT NOT NULL,
	parent_id INT,
	created_at DATE,
	modified_at DATE
);