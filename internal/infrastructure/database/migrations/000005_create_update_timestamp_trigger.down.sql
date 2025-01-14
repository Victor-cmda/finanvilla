-- 000005_create_update_timestamp_trigger.down.sql
DROP TRIGGER IF EXISTS update_users_timestamp ON users;
DROP TRIGGER IF EXISTS update_user_settings_timestamp ON user_settings;
DROP TRIGGER IF EXISTS update_permissions_timestamp ON permissions;
DROP FUNCTION IF EXISTS update_timestamp();
