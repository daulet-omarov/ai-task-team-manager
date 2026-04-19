CREATE TABLE types
(
    id         SERIAL PRIMARY KEY,
    name       TEXT,
    code       TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);