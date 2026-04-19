CREATE TABLE sprints
(
    id         SERIAL PRIMARY KEY,
    name       TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);