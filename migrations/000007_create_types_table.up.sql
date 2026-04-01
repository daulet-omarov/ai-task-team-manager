CREATE TABLE types (
                       id SERIAL PRIMARY KEY,
                       name TEXT,
                       code TEXT UNIQUE
);