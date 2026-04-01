CREATE TABLE teams (
                 id SERIAL PRIMARY KEY,
                 name TEXT,
                 code TEXT UNIQUE
);