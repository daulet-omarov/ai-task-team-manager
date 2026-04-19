CREATE TABLE entities
(
    id   SERIAL PRIMARY KEY,
    name TEXT,
    code TEXT UNIQUE
);