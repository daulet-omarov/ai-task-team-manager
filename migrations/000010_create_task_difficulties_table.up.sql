CREATE TABLE task_difficulties
(
    id   SERIAL PRIMARY KEY,
    name TEXT,
    code TEXT UNIQUE
);