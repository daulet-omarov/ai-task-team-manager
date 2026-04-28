CREATE TABLE employee_contributions
(
    id          SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
    date        DATE    NOT NULL,
    count       INTEGER NOT NULL DEFAULT 0,
    updated_at  TIMESTAMP        DEFAULT now(),
    UNIQUE (employee_id, date)
);
