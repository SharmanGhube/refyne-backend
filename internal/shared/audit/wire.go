package audit

import (
	"github.com/jmoiron/sqlx"
)

// ProvideAuditLogger provides the audit logger for wire dependency injection
func ProvideAuditLogger(db *sqlx.DB) *AuditLogger {
	return NewAuditLogger(db)
}
