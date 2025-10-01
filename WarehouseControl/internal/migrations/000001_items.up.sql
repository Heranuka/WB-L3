CREATE OR REPLACE FUNCTION log_item_changes() RETURNS TRIGGER AS $$
DECLARE
    changed_user_id INT;
    new_version INT;
    change_desc TEXT := '';
    change_diff JSONB := '{}';
BEGIN
    -- Получаем ID пользователя из переменной сессии
    changed_user_id := current_setting('myapp.current_user_id')::INT;

    -- Определяем следующую версию для сущности
    SELECT COALESCE(MAX(version), 0) + 1 INTO new_version FROM item_history WHERE item_id = COALESCE(NEW.id, OLD.id);

    IF TG_OP = 'UPDATE' THEN
        -- Сравнение полей и формирование описания изменений
        IF OLD.name IS DISTINCT FROM NEW.name THEN
            change_desc := change_desc || format('name: %s -> %s; ', OLD.name, NEW.name);
            change_diff := change_diff || jsonb_build_object(
                'name', jsonb_build_object('old', to_jsonb(OLD.name), 'new', to_jsonb(NEW.name))
            );
        END IF;

        IF OLD.description IS DISTINCT FROM NEW.description THEN
            change_desc := change_desc || format('description: %s -> %s; ', OLD.description, NEW.description);
            change_diff := change_diff || jsonb_build_object(
                'description', jsonb_build_object('old', to_jsonb(OLD.description), 'new', to_jsonb(NEW.description))
            );
        END IF;

        IF OLD.price IS DISTINCT FROM NEW.price THEN
            change_desc := change_desc || format('price: %s -> %s; ', OLD.price, NEW.price);
            change_diff := change_diff || jsonb_build_object(
                'price', jsonb_build_object('old', to_jsonb(OLD.price), 'new', to_jsonb(NEW.price))
            );
        END IF;

        IF OLD.stock IS DISTINCT FROM NEW.stock THEN
            change_desc := change_desc || format('stock: %s -> %s; ', OLD.stock, NEW.stock);
            change_diff := change_diff || jsonb_build_object(
                'stock', jsonb_build_object('old', to_jsonb(OLD.stock), 'new', to_jsonb(NEW.stock))
            );
        END IF;

        IF change_desc = '' THEN
            change_desc := 'No changes detected';
        END IF;

        INSERT INTO item_history(
            item_id, changed_by_user_id, change_description, changed_at, version, change_diff
        ) VALUES (
            NEW.id, changed_user_id, change_desc, NOW(), new_version, change_diff
        );

        RETURN NEW;

    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO item_history(
            item_id, changed_by_user_id, change_description, changed_at, version, change_diff
        ) VALUES (
            NEW.id, changed_user_id, 'Created item', NOW(), 1, to_jsonb(NEW)
        );

        RETURN NEW;

    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;


CREATE EXTENSION IF NOT EXISTS hstore;

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
   /*  access_token TEXT NOT NULL, */
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now()
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
CREATE TRIGGER trigger_log_item_changes
AFTER INSERT OR UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_changes();

CREATE TABLE item_history (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    changed_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    change_description TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 0,
    change_diff JSONB NULL
);
