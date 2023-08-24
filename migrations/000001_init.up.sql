BEGIN;

CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    email VARCHAR(254) NOT NULL
);

CREATE TABLE IF NOT EXISTS segments (
    segment_id SERIAL PRIMARY KEY,
    segment_name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_segments (
    user_segment_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users (user_id) ON DELETE CASCADE,
    segment_id INT REFERENCES segments (segment_id)
);

COMMIT;