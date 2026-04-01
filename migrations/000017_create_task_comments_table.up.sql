CREATE TABLE task_comments
(
    id         SERIAL PRIMARY KEY,
    content    TEXT,
    task_id    INTEGER REFERENCES tasks (id),
    author_id  INTEGER REFERENCES employees (id),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);