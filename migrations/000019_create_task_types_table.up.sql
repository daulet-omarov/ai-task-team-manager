CREATE TABLE task_types
(
    id      SERIAL PRIMARY KEY,
    task_id INTEGER REFERENCES tasks (id),
    type_id INTEGER REFERENCES types (id),
    UNIQUE (task_id, type_id)
);