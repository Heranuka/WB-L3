-- Active: 1752029824357@@127.0.0.1@5432
CREATE TABLE short_urls (
    id SERIAL PRIMARY KEY,
    short_code TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    custom BOOLEAN DEFAULT FALSE
);

CREATE TABLE clicks (
    id SERIAL PRIMARY KEY,
    short_url_id INTEGER REFERENCES short_urls(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    user_agent TEXT,
    ip_address TEXT
);
