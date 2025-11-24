-- Migration: Create restricted database user for application
-- This user has minimal privileges required for the application to run securely
-- Run this migration with a superuser account, then configure APP to use refyne_app_user

-- Create a restricted role for the application
-- This role will have only the necessary permissions
DO $$
BEGIN
    -- Drop role if exists (for idempotency during development)
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'refyne_app_user') THEN
        -- Revoke all privileges before dropping
        REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM refyne_app_user;
        REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM refyne_app_user;
        REVOKE ALL PRIVILEGES ON SCHEMA public FROM refyne_app_user;
        
        -- Reassign objects if any
        REASSIGN OWNED BY refyne_app_user TO postgres;
        DROP OWNED BY refyne_app_user;
        DROP ROLE refyne_app_user;
    END IF;
    
    -- Create the application user
    CREATE ROLE refyne_app_user WITH LOGIN PASSWORD 'CHANGE_THIS_PASSWORD_IN_PRODUCTION';
    
    -- Grant minimal schema access
    GRANT USAGE ON SCHEMA public TO refyne_app_user;
END $$;

-- Grant SELECT, INSERT, UPDATE, DELETE on all existing tables
-- Explicitly NO DROP, CREATE, ALTER, TRUNCATE permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO refyne_app_user;

-- Grant USAGE, SELECT on all sequences (for auto-increment IDs)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO refyne_app_user;

-- Set default privileges for future tables created by postgres user
-- This ensures new tables automatically get the same permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO refyne_app_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public 
GRANT USAGE, SELECT ON SEQUENCES TO refyne_app_user;

-- Create a comment to document the role
COMMENT ON ROLE refyne_app_user IS 'Restricted application user with minimal privileges (SELECT, INSERT, UPDATE, DELETE only). No DDL permissions.';

-- Security best practices applied:
-- 1. No SUPERUSER privilege
-- 2. No CREATEDB privilege  
-- 3. No CREATEROLE privilege
-- 4. No REPLICATION privilege
-- 5. No BYPASSRLS privilege (Row Level Security)
-- 6. Cannot execute DDL (DROP, CREATE, ALTER, TRUNCATE)
-- 7. Cannot modify schema structure
-- 8. Can only perform DML operations on existing tables
-- 9. Cannot access system catalogs beyond read-only
-- 10. Cannot modify other users or roles

-- Additional security: Set connection limit (optional)
-- ALTER ROLE refyne_app_user CONNECTION LIMIT 50;

-- Set statement timeout for this user (prevents long-running queries)
ALTER ROLE refyne_app_user SET statement_timeout = '30s';

-- Set lock timeout (prevents deadlocks)
ALTER ROLE refyne_app_user SET lock_timeout = '10s';

-- Set idle in transaction timeout (prevents holding transactions open)
ALTER ROLE refyne_app_user SET idle_in_transaction_session_timeout = '60s';

-- Revoke public schema creation (if not already done)
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

-- Note: After running this migration, you must manually grant CONNECT privilege 
-- on the current database using psql:
-- GRANT CONNECT ON DATABASE <your_database_name> TO refyne_app_user;
