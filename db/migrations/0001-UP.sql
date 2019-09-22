CREATE TABLE migrations (
	version_applied INT NOT NULL,
	file_applied VARCHAR(1024),
    created_at TIMESTAMP DEFAULT now()
);

ALTER TABLE tasks
	ADD COLUMN today BOOLEAN;

ALTER TABLE tasks	
	ALTER created_at SET DEFAULT now();
