CREATE TABLE difficulties
(
    id   SERIAL PRIMARY KEY,
    name TEXT,
    code TEXT UNIQUE
);