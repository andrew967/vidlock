-- +goose Up
CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    file_name TEXT NOT NULL,
    url TEXT,
    status TEXT NOT NULL,
    size BIGINT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS videos;
