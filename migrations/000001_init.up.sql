BEGIN;

CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    email VARCHAR(254) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS segments (
    segment_id BIGSERIAL PRIMARY KEY,
    automatic_percentage INT,
    segment_name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_segments (
    user_id BIGINT,
    segment_id BIGINT,
    expired_at TIMESTAMPTZ NOT NULL DEFAULT 'infinity',
    PRIMARY KEY (user_id, segment_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id),
    FOREIGN KEY (segment_id) REFERENCES segments (segment_id)
);

CREATE TYPE operation_enum AS ENUM ('added', 'deleted');

CREATE TABLE IF NOT EXISTS segment_history (
    history_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (user_id),
    segment_id BIGINT REFERENCES segments (segment_id),
    operation operation_enum NOT NULL,
    operation_timestamp TIMESTAMPTZ NOT NULL 
);

COMMIT;