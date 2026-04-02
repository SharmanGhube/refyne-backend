package workspace

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

func SetupWorkspaceRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	handler := registry.Workspace

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	// All workspace routes require authentication
	workspaceGroup := router.Group("/workspaces")
	workspaceGroup.Use(middlewares.AuthMiddleware())
	workspaceGroup.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
	{
		// Workspace CRUD
		workspaceGroup.POST("", handler.CreateWorkspace)
		workspaceGroup.GET("", handler.ListWorkspaces)
		workspaceGroup.GET("/:id", handler.GetWorkspace)
		workspaceGroup.PUT("/:id", handler.UpdateWorkspace)
		workspaceGroup.DELETE("/:id", handler.DeleteWorkspace)

		// Workspace members
		workspaceGroup.GET("/:id/members", handler.ListMembers)
		workspaceGroup.POST("/:id/members", handler.InviteMember)
		workspaceGroup.DELETE("/:id/members/:memberId", handler.RemoveMember)
	}
}
