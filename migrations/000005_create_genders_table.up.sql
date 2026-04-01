CREATE TABLE genders (
                         id SERIAL PRIMARY KEY,
                         name TEXT,
                         code TEXT UNIQUE
);