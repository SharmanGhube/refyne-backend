package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector holds all Prometheus metrics for the Refyne backend
type MetricsCollector struct {
	// HTTP Request Metrics
	HTTPRequestsTotal   prometheus.CounterVec
	HTTPRequestDuration prometheus.HistogramVec

	// Database Metrics
	DBConnectionsActive prometheus.GaugeVec
	DBConnectionsUsed   prometheus.GaugeVec

	// Redis Metrics
	RedisOperationsTotal prometheus.CounterVec
	RedisErrors          prometheus.CounterVec

	// Authentication Metrics
	AuthLoginAttempts   prometheus.CounterVec
	AuthLoginFailures   prometheus.CounterVec
	AuthTokensGenerated prometheus.CounterVec

	// Subscription Metrics
	PaddleAPICalls      prometheus.CounterVec
	PaddleAPIErrors     prometheus.CounterVec
	SubscriptionsByTier prometheus.GaugeVec

	// Email Metrics
	EmailJobsProcessed prometheus.CounterVec
	EmailJobsFailures  prometheus.CounterVec

	// Rate Limiting Metrics
	RateLimitExceeded prometheus.CounterVec
}

// NewMetricsCollector initializes all Prometheus metrics
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP Request Metrics
		HTTPRequestsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_http_requests_total",
				Help: "Total number of HTTP requests received",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "refyne_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // Default: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint"},
		),

		// Database Metrics
		DBConnectionsActive: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "refyne_db_connections_active",
				Help: "Number of active database connections",
			},
			[]string{"pool"},
		),
		DBConnectionsUsed: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "refyne_db_connections_used",
				Help: "Number of database connections in use",
			},
			[]string{"pool"},
		),

		// Redis Metrics
		RedisOperationsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_redis_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),
		RedisErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_redis_errors_total",
				Help: "Total number of Redis errors",
			},
			[]string{"operation", "error_type"},
		),

		// Authentication Metrics
		AuthLoginAttempts: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_auth_login_attempts_total",
				Help: "Total number of login attempts",
			},
			[]string{"method"}, // otp, password, refresh
		),
		AuthLoginFailures: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_auth_login_failures_total",
				Help: "Total number of failed login attempts",
			},
			[]string{"reason"}, // invalid_credentials, user_not_found, account_locked
		),
		AuthTokensGenerated: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_auth_tokens_generated_total",
				Help: "Total number of tokens generated",
			},
			[]string{"token_type"}, // access, refresh
		),

		// Subscription Metrics
		PaddleAPICalls: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_paddle_api_calls_total",
				Help: "Total number of Paddle API calls",
			},
			[]string{"operation", "status"},
		),
		PaddleAPIErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_paddle_api_errors_total",
				Help: "Total number of Paddle API errors",
			},
			[]string{"operation", "error_code"},
		),
		SubscriptionsByTier: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "refyne_subscriptions_by_tier",
				Help: "Number of active subscriptions by tier",
			},
			[]string{"tier", "status"},
		),

		// Email Metrics
		EmailJobsProcessed: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_email_jobs_processed_total",
				Help: "Total number of email jobs processed",
			},
			[]string{"email_type", "status"},
		),
		EmailJobsFailures: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_email_jobs_failures_total",
				Help: "Total number of failed email jobs",
			},
			[]string{"email_type", "reason"},
		),

		// Rate Limiting Metrics
		RateLimitExceeded: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "refyne_rate_limit_exceeded_total",
				Help: "Total number of requests rejected by rate limiting",
			},
			[]string{"endpoint", "client_ip"},
		),
	}
}

// Global metrics instance
var globalMetrics *MetricsCollector

// Initialize creates the global metrics collector
func Initialize() {
	globalMetrics = NewMetricsCollector()
}

// GetMetrics returns the global metrics collector
func GetMetrics() *MetricsCollector {
	if globalMetrics == nil {
		Initialize()
	}
	return globalMetrics
}

// RecordHTTPRequest records HTTP request metrics
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint string, statusCode int, durationSeconds float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(statusCode))).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(durationSeconds)
}

// RecordAuthLoginAttempt records login attempt
func (m *MetricsCollector) RecordAuthLoginAttempt(method string) {
	m.AuthLoginAttempts.WithLabelValues(method).Inc()
}

// RecordAuthLoginFailure records login failure
func (m *MetricsCollector) RecordAuthLoginFailure(reason string) {
	m.AuthLoginFailures.WithLabelValues(reason).Inc()
}

// RecordTokenGenerated records token generation
func (m *MetricsCollector) RecordTokenGenerated(tokenType string) {
	m.AuthTokensGenerated.WithLabelValues(tokenType).Inc()
}

// RecordPaddleAPICall records Paddle API call
func (m *MetricsCollector) RecordPaddleAPICall(operation, status string) {
	m.PaddleAPICalls.WithLabelValues(operation, status).Inc()
}

// RecordPaddleAPIError records Paddle API error
func (m *MetricsCollector) RecordPaddleAPIError(operation, errorCode string) {
	m.PaddleAPIErrors.WithLabelValues(operation, errorCode).Inc()
}

// RecordEmailJobProcessed records email job completion
func (m *MetricsCollector) RecordEmailJobProcessed(emailType, status string) {
	m.EmailJobsProcessed.WithLabelValues(emailType, status).Inc()
}

// RecordEmailJobFailure records email job failure
func (m *MetricsCollector) RecordEmailJobFailure(emailType, reason string) {
	m.EmailJobsFailures.WithLabelValues(emailType, reason).Inc()
}

// RecordRateLimitExceeded records rate limit exceeded
func (m *MetricsCollector) RecordRateLimitExceeded(endpoint, clientIP string) {
	m.RateLimitExceeded.WithLabelValues(endpoint, clientIP).Inc()
}

// UpdateDBConnections updates database connection metrics
func (m *MetricsCollector) UpdateDBConnections(poolName string, active, used float64) {
	m.DBConnectionsActive.WithLabelValues(poolName).Set(active)
	m.DBConnectionsUsed.WithLabelValues(poolName).Set(used)
}

// UpdateSubscriptionCounts updates subscription count by tier
func (m *MetricsCollector) UpdateSubscriptionCounts(tier, status string, count float64) {
	m.SubscriptionsByTier.WithLabelValues(tier, status).Set(count)
}

// RecordRedisOperation records Redis operation
func (m *MetricsCollector) RecordRedisOperation(operation, status string) {
	m.RedisOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordRedisError records Redis error
func (m *MetricsCollector) RecordRedisError(operation, errorType string) {
	m.RedisErrors.WithLabelValues(operation, errorType).Inc()
}
