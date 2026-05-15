-- +goose Up
CREATE TABLE employees (
    id            BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name     VARCHAR(200) NOT NULL,
    position      VARCHAR(200) NOT NULL,
    hired_at      DATE,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS employees;