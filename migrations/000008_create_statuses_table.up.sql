CREATE TABLE statuses (
                          id SERIAL PRIMARY KEY,
                          name TEXT,
                          code TEXT UNIQUE
);