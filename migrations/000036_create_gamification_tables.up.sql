CREATE TABLE user_gamification (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_points    INTEGER NOT NULL DEFAULT 0,
    current_level   INTEGER NOT NULL DEFAULT 1,
    current_streak  INTEGER NOT NULL DEFAULT 0,
    longest_streak  INTEGER NOT NULL DEFAULT 0,
    last_active_date DATE,
    updated_at      TIMESTAMP DEFAULT now()
);

CREATE TABLE point_transactions (
    id        SERIAL PRIMARY KEY,
    user_id   BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id   INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    points    INTEGER NOT NULL,
    reason    VARCHAR(50) NOT NULL,
    earned_at TIMESTAMP NOT NULL DEFAULT now(),
    metadata  JSONB
);

CREATE INDEX idx_point_transactions_user_earned ON point_transactions(user_id, earned_at);

CREATE TABLE kudos (
    id           SERIAL PRIMARY KEY,
    from_user_id BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id   BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id      INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    message      TEXT,
    created_at   TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_kudos_from_user_week ON kudos(from_user_id, created_at);
