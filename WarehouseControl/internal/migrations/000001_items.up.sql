CREATE TABLE users (
    id bigserial PRIMARY KEY,
    nickname TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- Убрано поле tokens_id, связь через таблицу tokens
);

CREATE TABLE tokens (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT NOT NULL,
    access_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL 
);


CREATE TABLE items (
    id bigserial PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE item_history (
    id bigserial PRIMARY KEY,
    item_id bigserial NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    changed_by_user_id bigint NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    change TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);