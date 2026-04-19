CREATE TABLE board_members
(
    id        SERIAL PRIMARY KEY,
    board_id  INTEGER NOT NULL REFERENCES boards (id) ON DELETE CASCADE,
    user_id   BIGINT  NOT NULL REFERENCES users (id)  ON DELETE CASCADE,
    role      VARCHAR(20) NOT NULL DEFAULT 'member', -- 'owner' | 'member'
    joined_at TIMESTAMP DEFAULT now(),
    UNIQUE (board_id, user_id)
);

CREATE INDEX idx_board_members_board_id ON board_members (board_id);
CREATE INDEX idx_board_members_user_id  ON board_members (user_id);
