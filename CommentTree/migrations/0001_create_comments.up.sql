-- Active: 1751554201026@@localhost@5432@postgres

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    content VARCHAR(255),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (NOW() at time zone 'utc'),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT, 
    parent_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

/* CREATE TABLE users (
     id SERIAL PRIMARY KEY,
     nickname VARCHAR(255)
);

CREATE TABLE posts (
     id SERIAL PRIMARY KEY,
     post_name VARCHAR(255)
); */