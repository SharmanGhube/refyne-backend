package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type WorkspaceRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewWorkspaceRepository(db *sqlx.DB) WorkspaceRepository {
	return &WorkspaceRepositoryImpl{
		name:   "WorkspaceRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("WorkspaceRepository"),
	}
}

func (r *WorkspaceRepositoryImpl) CreateWorkspace(c *gin.Context, workspace *models.Workspace) *errors.AppError {
	query := `
		INSERT INTO workspaces (id, user_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(c, query,
		workspace.ID, workspace.UserID, workspace.Name, workspace.Description, 
		workspace.IsActive, workspace.CreatedAt, workspace.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create workspace", zap.Error(err))
		return errors.NewAppError(c, "WORKSPACE_CREATE_FAILED", "Failed to create workspace",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	return nil
}

func (r *WorkspaceRepositoryImpl) GetWorkspaceByID(c *gin.Context, id string) (*models.Workspace, *errors.AppError) {
	query := `SELECT id, user_id, name, description, is_active, created_at, updated_at, deleted_at
		FROM workspaces WHERE id = $1 AND deleted_at IS NULL`

	var workspace models.Workspace
	err := r.db.GetContext(c, &workspace, query, id)

	if err != nil {
		r.logger.Debug("Workspace not found", zap.String("id", id))
		return nil, nil // Not found is not an error
	}

	return &workspace, nil
}

func (r *WorkspaceRepositoryImpl) GetUserWorkspaces(c *gin.Context, userID string) ([]*models.Workspace, *errors.AppError) {
	query := `SELECT id, user_id, name, description, is_active, created_at, updated_at, deleted_at
		FROM workspaces WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	var workspaces []*models.Workspace
	err := r.db.SelectContext(c, &workspaces, query, userID)

	if err != nil {
		r.logger.Error("Failed to fetch user workspaces", zap.Error(err))
		return nil, errors.NewAppError(c, "WORKSPACE_FETCH_FAILED", "Failed to fetch workspaces",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	return workspaces, nil
}

func (r *WorkspaceRepositoryImpl) UpdateWorkspace(c *gin.Context, workspace *models.Workspace) *errors.AppError {
	query := `UPDATE workspaces SET name = $1, description = $2, is_active = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(c, query,
		workspace.Name, workspace.Description, workspace.IsActive, workspace.UpdatedAt, workspace.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update workspace", zap.Error(err))
		return errors.NewAppError(c, "WORKSPACE_UPDATE_FAILED", "Failed to update workspace",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	return nil
}

func (r *WorkspaceRepositoryImpl) SoftDeleteWorkspace(c *gin.Context, workspaceID string) *errors.AppError {
	query := `UPDATE workspaces SET deleted_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(c, query, workspaceID)
	if err != nil {
		r.logger.Error("Failed to delete workspace", zap.Error(err))
		return errors.NewAppError(c, "WORKSPACE_DELETE_FAILED", "Failed to delete workspace",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(c, "WORKSPACE_NOT_FOUND", "Workspace not found",
			errors.ErrorTypeNotFound, errors.SeverityLow, "workspace")
	}

	return nil
}

func (r *WorkspaceRepositoryImpl) GetDB() *sqlx.DB {
	return r.db
}
