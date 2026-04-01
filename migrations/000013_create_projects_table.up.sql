CREATE TABLE projects
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    board_id   INTEGER REFERENCES boards (id),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);