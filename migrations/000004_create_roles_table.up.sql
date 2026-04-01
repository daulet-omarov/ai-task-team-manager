CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT,
    code TEXT UNIQUE
);