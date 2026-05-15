-- +goose Up
CREATE TABLE departments (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    parent_id   BIGINT REFERENCES departments(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(name, parent_id)
);

-- +goose Down
DROP TABLE IF EXISTS departments;