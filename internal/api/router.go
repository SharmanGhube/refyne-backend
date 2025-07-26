package api

import (
	"os"

	"github.com/gin-gonic/gin"
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/routes"
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

	// Load routes
	apiRoutes := router.Group("/api/v1")
	{
		auth.SetupAuthRoutes(apiRoutes, registry)
	}

	return router
}
