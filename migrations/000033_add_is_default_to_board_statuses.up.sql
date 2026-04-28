ALTER TABLE board_statuses ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT false;

UPDATE board_statuses
SET is_default = true
WHERE id IN (
    SELECT DISTINCT ON (board_id) id
    FROM board_statuses
    ORDER BY board_id, position ASC
);
