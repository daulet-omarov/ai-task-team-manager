CREATE TABLE priorities
(
    id   SERIAL PRIMARY KEY,
    name TEXT,
    code TEXT UNIQUE
);