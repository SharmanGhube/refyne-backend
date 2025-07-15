package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/shared/registry"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *registry.HandlerRegistry) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", registry.AuthHandler.Login)
		authGroup.POST("/register", registry.AuthHandler.Register)
		authGroup.POST("/refresh", nil)
		authGroup.POST("/logout", nil)
	}
}
