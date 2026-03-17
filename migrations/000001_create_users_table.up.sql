CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email TEXT NOT NULL UNIQUE,
                       password TEXT NOT NULL,
                       is_verified BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at TIMESTAMP DEFAULT now()
);