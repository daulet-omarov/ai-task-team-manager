-- employee.id must always equal user.id (1-to-1), so drop the auto-increment default.
-- IDs will be set explicitly from user.id on insert.
ALTER TABLE employees ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS employees_id_seq;
