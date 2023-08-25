BEGIN;

CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    email VARCHAR(254) NOT NULL
);

CREATE TABLE IF NOT EXISTS segments (
    segment_id BIGSERIAL PRIMARY KEY,
    segment_name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_segments (
    user_segment_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (user_id) ON DELETE CASCADE,
    segment_id BIGINT REFERENCES segments (segment_id) ON DELETE CASCADE,
    expired_at TIMESTAMPTZ NOT NULL,
    INDEX (user_id, segment_id)
);

CREATE TYPE operation_enum AS ENUM ('добавление', 'удаление');

CREATE TABLE IF NOT EXISTS segment_history (
    history_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (user_id),
    segment_id BIGINT REFERENCES segments (segment_id),
    operation operation_enum NOT NULL,
    operation_timestamp TIMESTAMPTZ NOT NULL 
);

COMMIT;