CREATE TABLE invites
(
    id         SERIAL PRIMARY KEY,
    board_id   INTEGER NOT NULL REFERENCES boards (id) ON DELETE CASCADE,
    inviter_id BIGINT  NOT NULL REFERENCES users (id)  ON DELETE CASCADE,
    invitee_id BIGINT  NOT NULL REFERENCES users (id)  ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending' | 'accepted' | 'rejected'
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_invites_invitee_id ON invites (invitee_id);
CREATE INDEX idx_invites_board_id   ON invites (board_id);
