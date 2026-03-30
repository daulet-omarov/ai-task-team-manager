CREATE TABLE employees (
                           id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           email                VARCHAR(255) NOT NULL UNIQUE,
                           name                 VARCHAR(200) NOT NULL,
                           role                 VARCHAR(50) NOT NULL CHECK (role IN ('developer', 'senior_developer', 'tech_lead', 'manager', 'pm')),
                           team_id              UUID REFERENCES teams(id),
                           hire_date            DATE NOT NULL,
                           is_active            BOOLEAN NOT NULL DEFAULT true,
                           max_concurrent_tasks INT NOT NULL DEFAULT 5,
                           available_hours_week NUMERIC(4,1) NOT NULL DEFAULT 40.0,
                           timezone             VARCHAR(50) NOT NULL DEFAULT 'UTC',
                           created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
                           updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE teams ADD CONSTRAINT fk_teams_manager
    FOREIGN KEY (manager_id) REFERENCES employees(id);