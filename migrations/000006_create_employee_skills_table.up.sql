CREATE TABLE employee_skills (
                                 id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
                                 skill_name      VARCHAR(100) NOT NULL,
                                 proficiency     INT NOT NULL CHECK (proficiency BETWEEN 1 AND 5),
                                 verified_tasks  INT NOT NULL DEFAULT 0,  -- how many tasks confirmed this skill
                                 created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
                                 UNIQUE(employee_id, skill_name)
);