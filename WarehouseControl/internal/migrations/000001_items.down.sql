-- rollback

DROP TABLE IF EXISTS item_history;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS tokens;

DROP TRIGGER IF EXISTS trigger_log_item_changes ON items;