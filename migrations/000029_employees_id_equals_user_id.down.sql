CREATE SEQUENCE IF NOT EXISTS employees_id_seq;
ALTER TABLE employees ALTER COLUMN id SET DEFAULT nextval('employees_id_seq');
