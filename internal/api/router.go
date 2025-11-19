package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth/routes"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

func NewRouter(registry *handlerregistry.HandlerRegistry) *gin.Engine {
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

	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/metrics"))
	router.Use(gin.Recovery())
	router.Use(middlewares.RequestIDMiddleware())

	// Register Routes
	apiRoutes := router.Group("/api")
	{
		// Public routes
		apiRoutes.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"status": "ok"})
		})

		// Auth routes (contains both public and protected)
		auth.SetupAuthRoutes(apiRoutes, registry)

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
