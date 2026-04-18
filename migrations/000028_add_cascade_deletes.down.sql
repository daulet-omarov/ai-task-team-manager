ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS employees_user_id_fkey,
    ADD CONSTRAINT employees_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE board_statuses
    DROP CONSTRAINT IF EXISTS board_statuses_board_id_fkey,
    ADD CONSTRAINT board_statuses_board_id_fkey FOREIGN KEY (board_id) REFERENCES boards (id);

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_board_id_fkey,
    ADD CONSTRAINT tasks_board_id_fkey FOREIGN KEY (board_id) REFERENCES boards (id);

ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS comments_task_id_fkey,
    ADD CONSTRAINT comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks (id);

ALTER TABLE attachments
    DROP CONSTRAINT IF EXISTS attachments_task_id_fkey,
    ADD CONSTRAINT attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks (id);

ALTER TABLE task_types
    DROP CONSTRAINT IF EXISTS task_types_task_id_fkey,
    ADD CONSTRAINT task_types_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks (id);
