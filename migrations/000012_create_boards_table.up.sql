CREATE TABLE boards
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    sprint_id  INTEGER REFERENCES sprints (id),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);