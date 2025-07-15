package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	authRoutes "github.com/refynehq/refyne-backend/internal/domain/auth/routes"
	"github.com/refynehq/refyne-backend/internal/shared/registry"
)

func NewRouter(registry *registry.HandlerRegistry) *gin.Engine {
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

	router.Use(gin.LoggerWithWriter(gin.DefaultWriter))
	router.Use(gin.Recovery())

	// Add request ID middleware using Google UUID
	// router.Use(middlewares.RequestIDMiddleware())
	// router.Use(middlewares.PrometheusMiddleware())

	// Define your routes here
	apiRoutes := router.Group("/api/v1")
	{
		authRoutes.SetupAuthRoutes(apiRoutes, registry)
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}
