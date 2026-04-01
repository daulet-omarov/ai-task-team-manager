CREATE TABLE task_attachments
(
    id         SERIAL PRIMARY KEY,
    task_id    INTEGER REFERENCES tasks (id),
    file_path  VARCHAR(255),
    file_name  VARCHAR(255),
    file_size  INTEGER,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);