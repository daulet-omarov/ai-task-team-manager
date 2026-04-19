CREATE TABLE board_statuses
(
    id         SERIAL PRIMARY KEY,
    status_id  INTEGER REFERENCES statuses (id),
    board_id   INTEGER REFERENCES boards (id),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    UNIQUE (status_id, board_id)
);