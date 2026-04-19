CREATE TABLE tasks
(
    id            SERIAL PRIMARY KEY,
    title         VARCHAR(255) NOT NULL,
    status_id     INTEGER REFERENCES statuses (id),
    priority_id   INTEGER REFERENCES priorities (id),
    difficulty_id INTEGER REFERENCES difficulties (id),
    board_id      INTEGER REFERENCES boards (id),
    developer_id  INTEGER REFERENCES employees (id),
    tester_id     INTEGER REFERENCES employees (id),
    reporter_id   INTEGER REFERENCES employees (id),
    description   TEXT,
    time_spent    INTEGER,
    created_at    TIMESTAMP DEFAULT now(),
    updated_at    TIMESTAMP DEFAULT now()
);