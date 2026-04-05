CREATE TABLE employees
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER REFERENCES users (id),
    full_name    VARCHAR(255),
    photo        VARCHAR(255),
    email        VARCHAR(255) unique,
    team_id      INTEGER REFERENCES teams (id),
    birthday     DATE,
    phone_number VARCHAR(11),
    gender_id    INTEGER references genders (id),
    created_at   TIMESTAMP DEFAULT now(),
    updated_at   TIMESTAMP DEFAULT now(),
    UNIQUE (user_id),
    UNIQUE (email)
);