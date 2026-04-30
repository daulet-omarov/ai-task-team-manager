ALTER TABLE point_transactions ADD COLUMN board_id INTEGER REFERENCES boards(id) ON DELETE SET NULL;
CREATE INDEX idx_point_transactions_board ON point_transactions(board_id, user_id, earned_at);
