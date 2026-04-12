package middlewares

import (
	"time"
)

// GetCurrentTimestamp returns current time in RFC3339 format
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
