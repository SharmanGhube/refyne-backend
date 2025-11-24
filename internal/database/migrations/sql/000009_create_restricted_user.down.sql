-- Rollback: Remove restricted database user

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'refyne_app_user') THEN
        -- Revoke all privileges
        REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM refyne_app_user;
        REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM refyne_app_user;
        REVOKE ALL PRIVILEGES ON SCHEMA public FROM refyne_app_user;
        
        -- Reassign owned objects
        REASSIGN OWNED BY refyne_app_user TO postgres;
        DROP OWNED BY refyne_app_user;
        
        -- Drop the role
        DROP ROLE refyne_app_user;
    END IF;
END $$;

-- Note: Default privileges are automatically removed when the role is dropped
-- Note: Database CONNECT privilege must be manually revoked if needed using:
-- REVOKE CONNECT ON DATABASE <your_database_name> FROM refyne_app_user;
