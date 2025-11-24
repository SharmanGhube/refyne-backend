-- Create audit_logs table for security event tracking
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(30) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    status VARCHAR(20) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    metadata JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_event_category ON audit_logs(event_category);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);

-- Composite indexes for common queries
CREATE INDEX idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_category_created ON audit_logs(event_category, created_at DESC);

-- Comment on table
COMMENT ON TABLE audit_logs IS 'Security and action audit trail for compliance and monitoring';
COMMENT ON COLUMN audit_logs.event_type IS 'Specific event (e.g., LOGIN_SUCCESS, PASSWORD_CHANGE)';
COMMENT ON COLUMN audit_logs.event_category IS 'Event category (e.g., AUTH, SECURITY, DATA)';
COMMENT ON COLUMN audit_logs.action IS 'Human-readable action description';
COMMENT ON COLUMN audit_logs.status IS 'Event outcome (SUCCESS, FAILURE, ERROR)';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context as JSON';
