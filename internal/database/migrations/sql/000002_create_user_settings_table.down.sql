-- Drop user_settings table and related objects
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
DROP INDEX IF EXISTS idx_user_settings_created_at;
DROP INDEX IF EXISTS idx_user_settings_user_id;
DROP TABLE IF EXISTS user_settings;
