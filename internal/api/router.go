package api

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/internal/monitoring"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth/routes"
	subscription "github.com/refynehq/refyne-backend/internal/domains/subscription/routes"
	user "github.com/refynehq/refyne-backend/internal/domains/user/routes"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

// prometheusMiddleware records HTTP request metrics
func prometheusMiddleware() gin.HandlerFunc {
	metrics := monitoring.GetMetrics()
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Skip metrics for health and metrics endpoints
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/api/health" {
			return
		}

		// Record metrics
		duration := time.Since(start).Seconds()
		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

func NewRouter(registry *handlerregistry.HandlerRegistry, db *sqlx.DB, redisClient *redis.Client) *gin.Engine {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	router := gin.New()

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize auth middleware with database for token version checking
	middlewares.InitializeAuthMiddleware(db)

	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/metrics"))
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(prometheusMiddleware())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.SecurityHeadersMiddleware())
	router.Use(middlewares.InputValidationMiddleware())
	router.Use(middlewares.ValidateRequestSize(10 * 1024 * 1024)) // 10MB max request size

	// Expose Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapF(promhttp.Handler().ServeHTTP))

	// Serve static files (checkout pages)
	router.Static("/static", "./static")
	router.StaticFile("/checkout.html", "./static/checkout.html")
	router.StaticFile("/checkout-success.html", "./static/checkout-success.html")

	// Initialize health checker
	healthChecker := NewHealthChecker(db, redisClient)

	// Register Routes
	apiRoutes := router.Group("/api")
	{
		// Public routes - Health checks
		apiRoutes.GET("/health", healthChecker.BasicHealthCheck)
		apiRoutes.GET("/health/detailed", healthChecker.DetailedHealthCheck)
		apiRoutes.GET("/health/ready", healthChecker.ReadinessCheck)
		apiRoutes.GET("/health/live", healthChecker.LivenessCheck)

		// Auth routes (contains both public and protected)
		auth.SetupAuthRoutes(apiRoutes, registry)

		// Subscription routes (checkout, webhooks, status)
		subscription.SetupSubscriptionRoutes(apiRoutes, registry)

		// User routes (profile, settings, onboarding)
		user.SetupUserRoutes(apiRoutes, registry)

		// Protected test route
		protected := apiRoutes.Group("/protected")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.GET("/me", func(ctx *gin.Context) {
				userID, _ := middlewares.GetUserID(ctx)
				email, _ := middlewares.GetUserEmail(ctx)
				username, _ := middlewares.GetUsername(ctx)

				ctx.JSON(200, gin.H{
					"message":    "Authentication successful",
					"user_id":    userID,
					"email":      email,
					"username":   username,
					"request_id": middlewares.GetRequestID(ctx),
				})
			})
		}
	}

	return router
}
