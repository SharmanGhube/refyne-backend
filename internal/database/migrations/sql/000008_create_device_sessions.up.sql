-- Create device_sessions table for tracking user login devices
CREATE TABLE IF NOT EXISTS device_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    device_type VARCHAR(50), -- 'mobile', 'tablet', 'desktop', 'unknown'
    browser VARCHAR(100),
    os VARCHAR(100),
    ip_address INET NOT NULL,
    country VARCHAR(100),
    city VARCHAR(100),
    is_suspicious BOOLEAN DEFAULT false,
    suspicion_reason TEXT,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true
);

-- Create indexes for efficient queries
CREATE INDEX idx_device_sessions_user_id ON device_sessions(user_id);
CREATE INDEX idx_device_sessions_fingerprint ON device_sessions(device_fingerprint);
CREATE INDEX idx_device_sessions_user_fingerprint ON device_sessions(user_id, device_fingerprint);
CREATE INDEX idx_device_sessions_last_used ON device_sessions(last_used_at DESC);
CREATE INDEX idx_device_sessions_active ON device_sessions(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_device_sessions_suspicious ON device_sessions(user_id, is_suspicious) WHERE is_suspicious = true;

-- Create table for tracking login attempts with location
CREATE TABLE IF NOT EXISTS login_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    country VARCHAR(100),
    city VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    login_count INTEGER DEFAULT 1,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_trusted BOOLEAN DEFAULT false
);

-- Create indexes for location tracking
CREATE INDEX idx_login_locations_user_id ON login_locations(user_id);
CREATE INDEX idx_login_locations_ip ON login_locations(ip_address);
CREATE INDEX idx_login_locations_user_ip ON login_locations(user_id, ip_address);
CREATE INDEX idx_login_locations_trusted ON login_locations(user_id, is_trusted) WHERE is_trusted = true;

-- Add comments
COMMENT ON TABLE device_sessions IS 'Tracks user sessions across different devices for security monitoring';
COMMENT ON TABLE login_locations IS 'Tracks login locations for detecting unusual access patterns';
COMMENT ON COLUMN device_sessions.device_fingerprint IS 'Hash of user-agent + IP address for device identification';
COMMENT ON COLUMN device_sessions.is_suspicious IS 'Flag for potentially compromised sessions';
COMMENT ON COLUMN login_locations.is_trusted IS 'Locations the user regularly logs in from';
