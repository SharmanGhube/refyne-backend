-- Drop indexes
DROP INDEX IF EXISTS idx_login_locations_trusted;
DROP INDEX IF EXISTS idx_login_locations_user_ip;
DROP INDEX IF EXISTS idx_login_locations_ip;
DROP INDEX IF EXISTS idx_login_locations_user_id;

DROP INDEX IF EXISTS idx_device_sessions_suspicious;
DROP INDEX IF EXISTS idx_device_sessions_active;
DROP INDEX IF EXISTS idx_device_sessions_last_used;
DROP INDEX IF EXISTS idx_device_sessions_user_fingerprint;
DROP INDEX IF EXISTS idx_device_sessions_fingerprint;
DROP INDEX IF EXISTS idx_device_sessions_user_id;

-- Drop tables
DROP TABLE IF EXISTS login_locations;
DROP TABLE IF EXISTS device_sessions;
