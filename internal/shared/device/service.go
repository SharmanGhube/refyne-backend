package device

import (
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/internal/shared/utils"
	"go.uber.org/zap"
)

// DeviceSession represents a user's device session
type DeviceSession struct {
	ID                string     `json:"id" db:"id"`
	UserID            string     `json:"user_id" db:"user_id"`
	DeviceFingerprint string     `json:"device_fingerprint" db:"device_fingerprint"`
	DeviceName        string     `json:"device_name" db:"device_name"`
	DeviceType        string     `json:"device_type" db:"device_type"`
	Browser           string     `json:"browser" db:"browser"`
	OS                string     `json:"os" db:"os"`
	IPAddress         string     `json:"ip_address" db:"ip_address"`
	Country           *string    `json:"country,omitempty" db:"country"`
	City              *string    `json:"city,omitempty" db:"city"`
	IsSuspicious      bool       `json:"is_suspicious" db:"is_suspicious"`
	SuspicionReason   *string    `json:"suspicion_reason,omitempty" db:"suspicion_reason"`
	LastUsedAt        time.Time  `json:"last_used_at" db:"last_used_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive          bool       `json:"is_active" db:"is_active"`
}

// LoginLocation represents a user's login location
type LoginLocation struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	Country     *string   `json:"country,omitempty" db:"country"`
	City        *string   `json:"city,omitempty" db:"city"`
	Latitude    *float64  `json:"latitude,omitempty" db:"latitude"`
	Longitude   *float64  `json:"longitude,omitempty" db:"longitude"`
	LoginCount  int       `json:"login_count" db:"login_count"`
	FirstSeenAt time.Time `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt  time.Time `json:"last_seen_at" db:"last_seen_at"`
	IsTrusted   bool      `json:"is_trusted" db:"is_trusted"`
}

// DeviceSessionService handles device session operations
type DeviceSessionService struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewDeviceSessionService creates a new device session service
func NewDeviceSessionService(db *sqlx.DB, logger *zap.Logger) *DeviceSessionService {
	return &DeviceSessionService{
		db:     db,
		logger: logger,
	}
}

// CreateOrUpdateSession creates a new device session or updates existing one
func (s *DeviceSessionService) CreateOrUpdateSession(c *gin.Context, userID string, deviceInfo *utils.DeviceInfo) (*DeviceSession, error) {
	ctx := c.Request.Context()
	requestID := middlewares.GetRequestID(c)

	// Check if session already exists
	existingSession, err := s.GetSessionByFingerprint(ctx, userID, deviceInfo.Fingerprint)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("Failed to check existing session",
			zap.String("requestID", requestID),
			zap.String("userID", userID),
			zap.Error(err))
		return nil, err
	}

	// Update existing session
	if existingSession != nil {
		query := `UPDATE device_sessions 
		          SET last_used_at = CURRENT_TIMESTAMP,
		              ip_address = $1,
		              is_active = true
		          WHERE id = $2
		          RETURNING id, user_id, device_fingerprint, device_name, device_type, browser, os, 
		                    ip_address, country, city, is_suspicious, suspicion_reason, 
		                    last_used_at, created_at, expires_at, is_active`

		var session DeviceSession
		err = s.db.GetContext(ctx, &session, query, deviceInfo.IPAddress, existingSession.ID)
		if err != nil {
			s.logger.Error("Failed to update device session",
				zap.String("requestID", requestID),
				zap.Error(err))
			return nil, err
		}

		s.logger.Info("Device session updated",
			zap.String("requestID", requestID),
			zap.String("sessionID", session.ID))

		return &session, nil
	}

	// Check for suspicious login
	isSuspicious, reason := s.checkSuspiciousLogin(ctx, userID, deviceInfo)

	// Create new session
	query := `INSERT INTO device_sessions 
	          (user_id, device_fingerprint, device_name, device_type, browser, os, 
	           ip_address, is_suspicious, suspicion_reason, expires_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	          RETURNING id, user_id, device_fingerprint, device_name, device_type, browser, os, 
	                    ip_address, country, city, is_suspicious, suspicion_reason, 
	                    last_used_at, created_at, expires_at, is_active`

	var session DeviceSession
	expiresAt := time.Now().Add(90 * 24 * time.Hour) // 90 days
	var suspicionReasonPtr *string
	if isSuspicious {
		suspicionReasonPtr = &reason
	}

	err = s.db.GetContext(ctx, &session, query,
		userID, deviceInfo.Fingerprint, deviceInfo.DeviceName, deviceInfo.DeviceType,
		deviceInfo.Browser, deviceInfo.OS, deviceInfo.IPAddress,
		isSuspicious, suspicionReasonPtr, expiresAt)

	if err != nil {
		s.logger.Error("Failed to create device session",
			zap.String("requestID", requestID),
			zap.String("userID", userID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Device session created",
		zap.String("requestID", requestID),
		zap.String("sessionID", session.ID),
		zap.Bool("suspicious", isSuspicious))

	// Track login location
	go s.trackLoginLocation(context.Background(), userID, deviceInfo.IPAddress)

	return &session, nil
}

// GetSessionByFingerprint retrieves a session by device fingerprint
func (s *DeviceSessionService) GetSessionByFingerprint(ctx context.Context, userID, fingerprint string) (*DeviceSession, error) {
	query := `SELECT id, user_id, device_fingerprint, device_name, device_type, browser, os, 
	                 ip_address, country, city, is_suspicious, suspicion_reason, 
	                 last_used_at, created_at, expires_at, is_active
	          FROM device_sessions
	          WHERE user_id = $1 AND device_fingerprint = $2 AND is_active = true
	          ORDER BY last_used_at DESC
	          LIMIT 1`

	var session DeviceSession
	err := s.db.GetContext(ctx, &session, query, userID, fingerprint)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *DeviceSessionService) GetUserSessions(ctx context.Context, userID string) ([]*DeviceSession, error) {
	query := `SELECT id, user_id, device_fingerprint, device_name, device_type, browser, os, 
	                 ip_address, country, city, is_suspicious, suspicion_reason, 
	                 last_used_at, created_at, expires_at, is_active
	          FROM device_sessions
	          WHERE user_id = $1 AND is_active = true
	          ORDER BY last_used_at DESC`

	var sessions []*DeviceSession
	err := s.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeSession revokes a specific device session
func (s *DeviceSessionService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	query := `UPDATE device_sessions 
	          SET is_active = false 
	          WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	s.logger.Info("Device session revoked",
		zap.String("userID", userID),
		zap.String("sessionID", sessionID))

	return nil
}

// RevokeAllSessions revokes all sessions for a user except current one
func (s *DeviceSessionService) RevokeAllSessions(ctx context.Context, userID, exceptSessionID string) error {
	query := `UPDATE device_sessions 
	          SET is_active = false 
	          WHERE user_id = $1 AND id != $2 AND is_active = true`

	result, err := s.db.ExecContext(ctx, query, userID, exceptSessionID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	s.logger.Info("All device sessions revoked",
		zap.String("userID", userID),
		zap.Int64("count", rows))

	return nil
}

// checkSuspiciousLogin checks if a login is suspicious
func (s *DeviceSessionService) checkSuspiciousLogin(ctx context.Context, userID string, deviceInfo *utils.DeviceInfo) (bool, string) {
	// Get known IPs
	knownIPs, err := s.getKnownIPs(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get known IPs", zap.Error(err))
		return false, ""
	}

	// Get known device fingerprints
	knownFingerprints, err := s.getKnownFingerprints(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get known fingerprints", zap.Error(err))
		return false, ""
	}

	return utils.IsSuspiciousLogin(deviceInfo.IPAddress, knownIPs, deviceInfo.Fingerprint, knownFingerprints)
}

// getKnownIPs retrieves known IP addresses for a user
func (s *DeviceSessionService) getKnownIPs(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT DISTINCT ip_address FROM device_sessions 
	          WHERE user_id = $1 AND is_suspicious = false
	          ORDER BY last_used_at DESC LIMIT 10`

	var ips []string
	err := s.db.SelectContext(ctx, &ips, query, userID)
	return ips, err
}

// getKnownFingerprints retrieves known device fingerprints for a user
func (s *DeviceSessionService) getKnownFingerprints(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT DISTINCT device_fingerprint FROM device_sessions 
	          WHERE user_id = $1 AND is_suspicious = false
	          ORDER BY last_used_at DESC LIMIT 10`

	var fingerprints []string
	err := s.db.SelectContext(ctx, &fingerprints, query, userID)
	return fingerprints, err
}

// trackLoginLocation tracks login location for future analysis
func (s *DeviceSessionService) trackLoginLocation(ctx context.Context, userID, ipAddress string) {
	// Check if location exists
	var location LoginLocation
	query := `SELECT id, login_count FROM login_locations 
	          WHERE user_id = $1 AND ip_address = $2`

	err := s.db.GetContext(ctx, &location, query, userID, ipAddress)
	if err == nil {
		// Update existing location
		updateQuery := `UPDATE login_locations 
		                SET login_count = login_count + 1,
		                    last_seen_at = CURRENT_TIMESTAMP
		                WHERE id = $1`
		_, _ = s.db.ExecContext(ctx, updateQuery, location.ID)
		return
	}

	// Create new location
	insertQuery := `INSERT INTO login_locations (user_id, ip_address) 
	                VALUES ($1, $2)`
	_, _ = s.db.ExecContext(ctx, insertQuery, userID, ipAddress)
}

// CleanupExpiredSessions removes expired device sessions
func (s *DeviceSessionService) CleanupExpiredSessions(ctx context.Context) error {
	query := `UPDATE device_sessions 
	          SET is_active = false 
	          WHERE expires_at < CURRENT_TIMESTAMP AND is_active = true`

	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	s.logger.Info("Expired device sessions cleaned up", zap.Int64("count", rows))
	return nil
}
