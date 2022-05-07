CREATE TABLE users (
	id CHAR(36) PRIMARY KEY,
	name VARCHAR(1024),
	email VARCHAR(1024) NOT NULL,
	password VARCHAR(1024) NOT NULL,
	created_at TIMESTAMP DEFAULT now(),
	modified_at TIMESTAMP
);

CREATE TABLE user_logins (
	id SERIAL PRIMARY KEY,
	user_id VARCHAR(36) NOT NULL,
	ip_address INET,
	created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE task_users (
	task_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	CONSTRAINT pk_taskusers PRIMARY KEY (task_id, user_id),
	FOREIGN KEY (task_id) REFERENCES tasks(id),
	FOREIGN KEY (user_id) REFERENCES users(id)
);
