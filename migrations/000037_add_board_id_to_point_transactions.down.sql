DROP INDEX IF EXISTS idx_point_transactions_board;
ALTER TABLE point_transactions DROP COLUMN IF EXISTS board_id;
