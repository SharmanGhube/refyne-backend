package auth

import (
	"github.com/gin-gonic/gin"
	registry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *registry.HandlerRegistry) {
	authhandler := registry.AuthHandler
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authhandler.Login)
		authGroup.POST("/register", authhandler.Register)
		authGroup.POST("/refresh", nil)
		authGroup.POST("/logout", nil)
	}
}
