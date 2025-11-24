package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// EventCategory represents the category of audit event
type EventCategory string

const (
	CategoryAuth     EventCategory = "AUTH"
	CategorySecurity EventCategory = "SECURITY"
	CategoryData     EventCategory = "DATA"
	CategorySystem   EventCategory = "SYSTEM"
	CategoryUser     EventCategory = "USER"
)

// EventType represents specific audit event types
type EventType string

const (
	// Auth events
	EventLoginSuccess        EventType = "LOGIN_SUCCESS"
	EventLoginFailure        EventType = "LOGIN_FAILURE"
	EventLogout              EventType = "LOGOUT"
	EventLogoutAll           EventType = "LOGOUT_ALL"
	EventRegistration        EventType = "REGISTRATION"
	EventPasswordChange      EventType = "PASSWORD_CHANGE"
	EventPasswordReset       EventType = "PASSWORD_RESET"
	EventOTPGenerated        EventType = "OTP_GENERATED"
	EventOTPVerified         EventType = "OTP_VERIFIED"
	EventOTPFailed           EventType = "OTP_FAILED"
	EventTokenRefresh        EventType = "TOKEN_REFRESH"
	EventAccountVerification EventType = "ACCOUNT_VERIFICATION"

	// Security events
	EventAccountLocked      EventType = "ACCOUNT_LOCKED"
	EventAccountUnlocked    EventType = "ACCOUNT_UNLOCKED"
	EventSuspiciousActivity EventType = "SUSPICIOUS_ACTIVITY"
	EventRateLimitExceeded  EventType = "RATE_LIMIT_EXCEEDED"
	EventUnauthorizedAccess EventType = "UNAUTHORIZED_ACCESS"
	EventPermissionDenied   EventType = "PERMISSION_DENIED"

	// Data events
	EventDataAccess       EventType = "DATA_ACCESS"
	EventDataModification EventType = "DATA_MODIFICATION"
	EventDataDeletion     EventType = "DATA_DELETION"

	// System events
	EventSystemError  EventType = "SYSTEM_ERROR"
	EventConfigChange EventType = "CONFIG_CHANGE"
)

// EventStatus represents the outcome of an event
type EventStatus string

const (
	StatusSuccess EventStatus = "SUCCESS"
	StatusFailure EventStatus = "FAILURE"
	StatusError   EventStatus = "ERROR"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string                 `db:"id" json:"id"`
	UserID        *string                `db:"user_id" json:"user_id,omitempty"`
	EventType     EventType              `db:"event_type" json:"event_type"`
	EventCategory EventCategory          `db:"event_category" json:"event_category"`
	Action        string                 `db:"action" json:"action"`
	ResourceType  *string                `db:"resource_type" json:"resource_type,omitempty"`
	ResourceID    *string                `db:"resource_id" json:"resource_id,omitempty"`
	Status        EventStatus            `db:"status" json:"status"`
	IPAddress     *string                `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent     *string                `db:"user_agent" json:"user_agent,omitempty"`
	RequestID     *string                `db:"request_id" json:"request_id,omitempty"`
	Metadata      map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	ErrorMessage  *string                `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time              `db:"created_at" json:"created_at"`
}

// AuditLogger handles audit logging operations
type AuditLogger struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(db *sqlx.DB) *AuditLogger {
	return &AuditLogger{
		db:     db,
		logger: logging.GetLogger(),
	}
}

// LogEvent logs an audit event
func (a *AuditLogger) LogEvent(ctx context.Context, log *AuditLog) error {
	// Serialize metadata to JSON
	var metadataJSON []byte
	var err error
	if log.Metadata != nil {
		metadataJSON, err = json.Marshal(log.Metadata)
		if err != nil {
			a.logger.Error("Failed to marshal audit log metadata",
				zap.Error(err),
				zap.String("event_type", string(log.EventType)),
			)
			metadataJSON = []byte("{}")
		}
	}

	query := `
		INSERT INTO audit_logs (
			user_id, event_type, event_category, action, 
			resource_type, resource_id, status, ip_address, 
			user_agent, request_id, metadata, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err = a.db.ExecContext(ctx, query,
		log.UserID,
		log.EventType,
		log.EventCategory,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.Status,
		log.IPAddress,
		log.UserAgent,
		log.RequestID,
		metadataJSON,
		log.ErrorMessage,
	)

	if err != nil {
		a.logger.Error("Failed to insert audit log",
			zap.Error(err),
			zap.String("event_type", string(log.EventType)),
		)
		return err
	}

	a.logger.Info("Audit event logged",
		zap.String("event_type", string(log.EventType)),
		zap.String("event_category", string(log.EventCategory)),
		zap.String("status", string(log.Status)),
	)

	return nil
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (a *AuditLogger) GetUserAuditLogs(ctx context.Context, userID string, limit int, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, event_type, event_category, action,
		       resource_type, resource_id, status, ip_address,
		       user_agent, request_id, metadata, error_message, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var logs []AuditLog
	err := a.db.SelectContext(ctx, &logs, query, userID, limit, offset)
	if err != nil {
		a.logger.Error("Failed to retrieve user audit logs",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, err
	}

	return logs, nil
}

// GetAuditLogsByCategory retrieves audit logs by category
func (a *AuditLogger) GetAuditLogsByCategory(ctx context.Context, category EventCategory, limit int, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, event_type, event_category, action,
		       resource_type, resource_id, status, ip_address,
		       user_agent, request_id, metadata, error_message, created_at
		FROM audit_logs
		WHERE event_category = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var logs []AuditLog
	err := a.db.SelectContext(ctx, &logs, query, category, limit, offset)
	if err != nil {
		a.logger.Error("Failed to retrieve audit logs by category",
			zap.Error(err),
			zap.String("category", string(category)),
		)
		return nil, err
	}

	return logs, nil
}

// GetSecurityEvents retrieves security-related audit logs (failed logins, lockouts, etc.)
func (a *AuditLogger) GetSecurityEvents(ctx context.Context, limit int, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, event_type, event_category, action,
		       resource_type, resource_id, status, ip_address,
		       user_agent, request_id, metadata, error_message, created_at
		FROM audit_logs
		WHERE event_category = $1 OR status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	var logs []AuditLog
	err := a.db.SelectContext(ctx, &logs, query, CategorySecurity, StatusFailure, limit, offset)
	if err != nil {
		a.logger.Error("Failed to retrieve security events",
			zap.Error(err),
		)
		return nil, err
	}

	return logs, nil
}

// CleanupOldLogs removes audit logs older than the specified duration
func (a *AuditLogger) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	query := `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := a.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		a.logger.Error("Failed to cleanup old audit logs",
			zap.Error(err),
		)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()

	a.logger.Info("Cleaned up old audit logs",
		zap.Int64("rows_deleted", rowsAffected),
		zap.Time("cutoff_time", cutoffTime),
	)

	return rowsAffected, nil
}
