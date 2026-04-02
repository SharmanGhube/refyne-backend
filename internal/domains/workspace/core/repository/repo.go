package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	workspaceModels "github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// WorkspaceRepository defines CRUD operations for workspaces
type WorkspaceRepository interface {
	CreateWorkspace(c *gin.Context, workspace *workspaceModels.Workspace) *errors.AppError
	GetWorkspaceByID(c *gin.Context, id string) (*workspaceModels.Workspace, *errors.AppError)
	GetUserWorkspaces(c *gin.Context, userID string) ([]*workspaceModels.Workspace, *errors.AppError)
	UpdateWorkspace(c *gin.Context, workspace *workspaceModels.Workspace) *errors.AppError
	SoftDeleteWorkspace(c *gin.Context, id string) *errors.AppError
	GetDB() *sqlx.DB
}

// WorkspaceRepositoryImpl implements WorkspaceRepository
type WorkspaceRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(db *sqlx.DB) WorkspaceRepository {
	return &WorkspaceRepositoryImpl{
		name:   "WorkspaceRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("WorkspaceRepository"),
	}
}

// CreateWorkspace creates a new workspace
func (r *WorkspaceRepositoryImpl) CreateWorkspace(c *gin.Context, workspace *workspaceModels.Workspace) *errors.AppError {
	query := `
		INSERT INTO workspaces (id, user_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if err := r.db.GetContext(c, workspace, query,
		workspace.ID,
		workspace.UserID,
		workspace.Name,
		workspace.Description,
		workspace.IsActive,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	); err != nil {
		r.logger.Error("Failed to create workspace", zap.Error(err), zap.String("workspace_id", workspace.ID))
		return errors.NewAppError(
			c,
			"WORKSPACE_CREATION_FAILED",
			"Failed to create workspace",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	return nil
}

// GetWorkspaceByID retrieves a workspace by ID
func (r *WorkspaceRepositoryImpl) GetWorkspaceByID(c *gin.Context, id string) (*workspaceModels.Workspace, *errors.AppError) {
	workspace := &workspaceModels.Workspace{}

	query := `
		SELECT id, user_id, name, description, is_active, created_at, updated_at, deleted_at
		FROM workspaces
		WHERE id = $1 AND deleted_at IS NULL
	`

	if err := r.db.GetContext(c, workspace, query, id); err != nil {
		r.logger.Debug("Workspace not found", zap.String("workspace_id", id), zap.Error(err))
		return nil, nil
	}

	return workspace, nil
}

// GetUserWorkspaces retrieves all workspaces for a user
func (r *WorkspaceRepositoryImpl) GetUserWorkspaces(c *gin.Context, userID string) ([]*workspaceModels.Workspace, *errors.AppError) {
	workspaces := []*workspaceModels.Workspace{}

	query := `
		SELECT id, user_id, name, description, is_active, created_at, updated_at, deleted_at
		FROM workspaces
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	if err := r.db.SelectContext(c, &workspaces, query, userID); err != nil {
		r.logger.Error("Failed to get user workspaces", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewAppError(
			c,
			"GET_WORKSPACES_FAILED",
			"Failed to retrieve workspaces",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	return workspaces, nil
}

// UpdateWorkspace updates an existing workspace
func (r *WorkspaceRepositoryImpl) UpdateWorkspace(c *gin.Context, workspace *workspaceModels.Workspace) *errors.AppError {
	query := `
		UPDATE workspaces
		SET name = $1, description = $2, is_active = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(c, query,
		workspace.Name,
		workspace.Description,
		workspace.IsActive,
		workspace.UpdatedAt,
		workspace.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update workspace", zap.Error(err), zap.String("workspace_id", workspace.ID))
		return errors.NewAppError(
			c,
			"WORKSPACE_UPDATE_FAILED",
			"Failed to update workspace",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"WORKSPACE_NOT_FOUND",
			"Workspace not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"workspace",
		)
	}

	return nil
}

// SoftDeleteWorkspace performs a soft delete on a workspace
func (r *WorkspaceRepositoryImpl) SoftDeleteWorkspace(c *gin.Context, id string) *errors.AppError {
	query := `
		UPDATE workspaces
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(c, query, id)
	if err != nil {
		r.logger.Error("Failed to delete workspace", zap.Error(err), zap.String("workspace_id", id))
		return errors.NewAppError(
			c,
			"WORKSPACE_DELETE_FAILED",
			"Failed to delete workspace",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"WORKSPACE_NOT_FOUND",
			"Workspace not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"workspace",
		)
	}

	return nil
}

// GetDB returns the database connection
func (r *WorkspaceRepositoryImpl) GetDB() *sqlx.DB {
	return r.db
}
