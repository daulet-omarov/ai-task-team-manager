CREATE TABLE employee_vacations (
                                    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
                                    start_date  DATE NOT NULL,
                                    end_date    DATE NOT NULL,
                                    CHECK (end_date >= start_date)
);