CREATE OR REPLACE FUNCTION log_item_changes() RETURNS TRIGGER AS $$
DECLARE
    changed_user_id INT;
BEGIN
    -- Получаем user_id из переменной
    changed_user_id := current_setting('myapp.current_user_id')::INT;

    -- Лог уведомление
    RAISE NOTICE 'User ID: %, Operation: %, Item ID: %', changed_user_id, TG_OP, COALESCE(NEW.id, OLD.id);

    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history(item_id, changed_by_user_id, change_description, changed_at, version, change_diff)
        VALUES (NEW.id, changed_user_id, 'Created item', NOW(), 1, to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO item_history(item_id, changed_by_user_id, change_description, changed_at, version, change_diff)
        VALUES (NEW.id, changed_user_id, 'Updated item', NOW(), OLD.version + 1, to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history(item_id, changed_by_user_id, change_description, changed_at, version, change_diff)
        VALUES (OLD.id, changed_user_id, 'Deleted item', NOW(), OLD.version + 1, to_jsonb(OLD));
        RETURN OLD;
    END IF;

    RETURN NULL; -- или можно вернуть NULL, если не обрабатываем операцию
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_item_changes
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_changes();
