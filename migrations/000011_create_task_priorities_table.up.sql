CREATE TABLE task_priorities (
                                 id SERIAL PRIMARY KEY,
                                 name TEXT,
                                 code TEXT UNIQUE
);