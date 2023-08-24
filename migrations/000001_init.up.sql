CREATE TABLE user (
    user_id SERIAL PRIMARY KEY
);

CREATE TABLE segment (
    segment_id SERIAL PRIMARY KEY,
    segment_name VARCHAR(255) NOT NULL
);

CREATE TABLE user_segment (
    user_segment_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES user (user_id) ON DELETE CASCADE,
    segment_id INT REFERENCES segment (segment_id)
);
