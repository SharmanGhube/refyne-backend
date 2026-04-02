package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// WorkspaceRepository defines data access methods for workspaces
type WorkspaceRepository interface {
	// CreateWorkspace creates a new workspace
	CreateWorkspace(c *gin.Context, workspace *models.Workspace) *errors.AppError

	// GetWorkspaceByID retrieves a workspace by ID
	GetWorkspaceByID(c *gin.Context, id string) (*models.Workspace, *errors.AppError)

	// GetUserWorkspaces retrieves all workspaces for a user
	GetUserWorkspaces(c *gin.Context, userID string) ([]*models.Workspace, *errors.AppError)

	// UpdateWorkspace updates a workspace
	UpdateWorkspace(c *gin.Context, workspace *models.Workspace) *errors.AppError

	// SoftDeleteWorkspace soft deletes a workspace
	SoftDeleteWorkspace(c *gin.Context, workspaceID string) *errors.AppError

	// GetDB returns the database connection
	GetDB() *sqlx.DB
}

// WorkspaceMemberRepository defines data access methods for workspace members
type WorkspaceMemberRepository interface {
	// AddMember adds a member to a workspace
	AddMember(c *gin.Context, member *models.WorkspaceMember) *errors.AppError

	// GetMember retrieves a member from a workspace
	GetMember(c *gin.Context, workspaceID, userID string) (*models.WorkspaceMember, *errors.AppError)

	// GetWorkspaceMembers retrieves all members in a workspace
	GetWorkspaceMembers(c *gin.Context, workspaceID string) ([]*models.WorkspaceMember, *errors.AppError)

	// RemoveMember removes a member from a workspace
	RemoveMember(c *gin.Context, workspaceID, userID string) *errors.AppError

	// GetDB returns the database connection
	GetDB() *sqlx.DB
}
