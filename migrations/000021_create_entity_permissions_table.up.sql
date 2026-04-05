CREATE TABLE entity_permissions
(
    id   SERIAL PRIMARY KEY,
    entity_id INTEGER REFERENCES entities (id),
    name TEXT,
    code TEXT UNIQUE
);