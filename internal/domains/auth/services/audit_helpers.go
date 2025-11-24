package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/internal/shared/audit"
	"go.uber.org/zap"
)

// Helper methods for audit logging

func (s *AuthServiceImpl) logAuditEvent(c *gin.Context, userID *string, eventType audit.EventType, category audit.EventCategory, action string, status audit.EventStatus, metadata map[string]interface{}, errorMsg *string) {
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := middlewares.GetRequestID(c)

	log := &audit.AuditLog{
		UserID:        userID,
		EventType:     eventType,
		EventCategory: category,
		Action:        action,
		Status:        status,
		IPAddress:     &ip,
		UserAgent:     &userAgent,
		RequestID:     &requestID,
		Metadata:      metadata,
		ErrorMessage:  errorMsg,
	}

	// Log asynchronously to avoid blocking the request
	go func() {
		if err := s.auditLogger.LogEvent(c.Request.Context(), log); err != nil {
			s.logger.Error("Failed to log audit event",
				zap.String("event_type", string(eventType)),
				zap.Error(err),
			)
		}
	}()
}

func (s *AuthServiceImpl) logSuccessfulLogin(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventLoginSuccess, audit.CategoryAuth,
		"User logged in successfully",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}

func (s *AuthServiceImpl) logFailedLogin(c *gin.Context, email, reason string) {
	reasonMsg := "Failed login attempt: " + reason
	s.logAuditEvent(c, nil, audit.EventLoginFailure, audit.CategoryAuth,
		"Failed login attempt",
		audit.StatusFailure,
		map[string]interface{}{"email": email, "reason": reason},
		&reasonMsg,
	)
}

func (s *AuthServiceImpl) logRegistration(c *gin.Context, userID, email, username string) {
	s.logAuditEvent(c, &userID, audit.EventRegistration, audit.CategoryAuth,
		"New user registered",
		audit.StatusSuccess,
		map[string]interface{}{"email": email, "username": username},
		nil,
	)
}

func (s *AuthServiceImpl) logPasswordChange(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventPasswordChange, audit.CategorySecurity,
		"User password changed",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}

func (s *AuthServiceImpl) logPasswordReset(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventPasswordReset, audit.CategorySecurity,
		"Password reset completed",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}

func (s *AuthServiceImpl) logAccountLocked(c *gin.Context, userID, email, reason string) {
	s.logAuditEvent(c, &userID, audit.EventAccountLocked, audit.CategorySecurity,
		"Account locked due to suspicious activity",
		audit.StatusSuccess,
		map[string]interface{}{"email": email, "reason": reason},
		nil,
	)
}

func (s *AuthServiceImpl) logOTPGenerated(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventOTPGenerated, audit.CategoryAuth,
		"OTP generated for login",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}

func (s *AuthServiceImpl) logOTPVerified(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventOTPVerified, audit.CategoryAuth,
		"OTP verified successfully",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}

func (s *AuthServiceImpl) logOTPFailed(c *gin.Context, email, reason string) {
	reasonMsg := "OTP verification failed: " + reason
	s.logAuditEvent(c, nil, audit.EventOTPFailed, audit.CategoryAuth,
		"OTP verification failed",
		audit.StatusFailure,
		map[string]interface{}{"email": email, "reason": reason},
		&reasonMsg,
	)
}

func (s *AuthServiceImpl) logLogout(c *gin.Context, userID string) {
	s.logAuditEvent(c, &userID, audit.EventLogout, audit.CategoryAuth,
		"User logged out",
		audit.StatusSuccess,
		nil,
		nil,
	)
}

func (s *AuthServiceImpl) logLogoutAll(c *gin.Context, userID string) {
	s.logAuditEvent(c, &userID, audit.EventLogoutAll, audit.CategorySecurity,
		"User logged out from all devices",
		audit.StatusSuccess,
		nil,
		nil,
	)
}

func (s *AuthServiceImpl) logAccountVerification(c *gin.Context, userID, email string) {
	s.logAuditEvent(c, &userID, audit.EventAccountVerification, audit.CategoryAuth,
		"Account verified successfully",
		audit.StatusSuccess,
		map[string]interface{}{"email": email},
		nil,
	)
}
