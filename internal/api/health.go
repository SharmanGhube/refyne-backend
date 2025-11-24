package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db    *sqlx.DB
	redis *redis.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sqlx.DB, redis *redis.Client) *HealthChecker {
	return &HealthChecker{
		db:    db,
		redis: redis,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// BasicHealthCheck returns basic health status
func (h *HealthChecker) BasicHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// DetailedHealthCheck returns detailed health status with dependency checks
func (h *HealthChecker) DetailedHealthCheck(c *gin.Context) {
	checks := make(map[string]HealthCheck)
	overallStatus := "healthy"

	// Check database
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check Redis (optional - may not be configured)
	if h.redis != nil {
		redisCheck := h.checkRedis(c.Request.Context())
		checks["redis"] = redisCheck
		if redisCheck.Status != "healthy" {
			overallStatus = "degraded" // Redis is optional, so degraded not unhealthy
		}
	}

	// System checks
	checks["api"] = HealthCheck{
		Status:  "healthy",
		Message: "API server running",
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK // Still OK, but with warning
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks:    checks,
	}

	c.JSON(statusCode, response)
}

// checkDatabase checks database connectivity
func (h *HealthChecker) checkDatabase() HealthCheck {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result int
	err := h.db.GetContext(ctx, &result, "SELECT 1")

	latency := time.Since(start)

	if err != nil {
		if err == sql.ErrNoRows || err == context.DeadlineExceeded {
			return HealthCheck{
				Status:  "unhealthy",
				Message: "Database connection timeout",
				Latency: latency.String(),
			}
		}
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Database connection failed",
			Latency: latency.String(),
		}
	}

	status := "healthy"
	if latency > 100*time.Millisecond {
		status = "degraded"
	}

	return HealthCheck{
		Status:  status,
		Latency: latency.String(),
	}
}

// checkRedis checks Redis connectivity
func (h *HealthChecker) checkRedis(ctx context.Context) HealthCheck {
	start := time.Now()

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := h.redis.Ping(timeoutCtx).Err()
	latency := time.Since(start)

	if err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "Redis connection failed",
			Latency: latency.String(),
		}
	}

	status := "healthy"
	if latency > 50*time.Millisecond {
		status = "degraded"
	}

	return HealthCheck{
		Status:  status,
		Latency: latency.String(),
	}
}

// ReadinessCheck checks if the service is ready to accept requests
func (h *HealthChecker) ReadinessCheck(c *gin.Context) {
	// Check critical dependencies
	dbCheck := h.checkDatabase()

	if dbCheck.Status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":   false,
			"message": "Service not ready - database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}

// LivenessCheck checks if the service is alive
func (h *HealthChecker) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive": true,
	})
}
