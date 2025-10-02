CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    content TEXT,
    parent_id INT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (NOW() at time zone 'utc'),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (NOW() at time zone 'utc'),
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);
