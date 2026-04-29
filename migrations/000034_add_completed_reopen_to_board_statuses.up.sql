ALTER TABLE board_statuses ADD COLUMN is_completed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE board_statuses ADD COLUMN is_reopen BOOLEAN NOT NULL DEFAULT false;

-- Mark the 'done' status as completed for every board that has it
UPDATE board_statuses bs
SET is_completed = true
FROM statuses s
WHERE s.id = bs.status_id AND s.code = 'done';

-- Mark the 'to_do' status as reopen for every board that has it
UPDATE board_statuses bs
SET is_reopen = true
FROM statuses s
WHERE s.id = bs.status_id AND s.code = 'to_do';
