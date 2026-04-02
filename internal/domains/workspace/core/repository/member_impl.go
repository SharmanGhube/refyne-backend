package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type WorkspaceMemberRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewWorkspaceMemberRepository(db *sqlx.DB) WorkspaceMemberRepository {
	return &WorkspaceMemberRepositoryImpl{
		name:   "WorkspaceMemberRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("WorkspaceMemberRepository"),
	}
}

func (r *WorkspaceMemberRepositoryImpl) AddMember(c *gin.Context, member *models.WorkspaceMember) *errors.AppError {
	query := `INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(c, query,
		member.ID, member.WorkspaceID, member.UserID, member.Role, 
		member.JoinedAt, member.CreatedAt, member.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to add member", zap.Error(err))
		return errors.NewAppError(c, "MEMBER_ADD_FAILED", "Failed to add member to workspace",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	return nil
}

func (r *WorkspaceMemberRepositoryImpl) GetMember(c *gin.Context, workspaceID, userID string) (*models.WorkspaceMember, *errors.AppError) {
	query := `SELECT id, workspace_id, user_id, role, joined_at, created_at, updated_at
		FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`

	var member models.WorkspaceMember
	err := r.db.GetContext(c, &member, query, workspaceID, userID)

	if err != nil {
		r.logger.Debug("Member not found", zap.String("workspace_id", workspaceID), zap.String("user_id", userID))
		return nil, nil // Not found is not an error
	}

	return &member, nil
}

func (r *WorkspaceMemberRepositoryImpl) GetWorkspaceMembers(c *gin.Context, workspaceID string) ([]*models.WorkspaceMember, *errors.AppError) {
	query := `SELECT id, workspace_id, user_id, role, joined_at, created_at, updated_at
		FROM workspace_members WHERE workspace_id = $1 ORDER BY role DESC, created_at ASC`

	var members []*models.WorkspaceMember
	err := r.db.SelectContext(c, &members, query, workspaceID)

	if err != nil {
		r.logger.Error("Failed to fetch workspace members", zap.Error(err))
		return nil, errors.NewAppError(c, "MEMBERS_FETCH_FAILED", "Failed to fetch workspace members",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	return members, nil
}

func (r *WorkspaceMemberRepositoryImpl) RemoveMember(c *gin.Context, workspaceID, userID string) *errors.AppError {
	query := `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(c, query, workspaceID, userID)
	if err != nil {
		r.logger.Error("Failed to remove member", zap.Error(err))
		return errors.NewAppError(c, "MEMBER_REMOVE_FAILED", "Failed to remove member from workspace",
			errors.ErrorTypeInternal, errors.SeverityHigh, "workspace")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(c, "MEMBER_NOT_FOUND", "Member not found",
			errors.ErrorTypeNotFound, errors.SeverityLow, "workspace")
	}

	return nil
}

func (r *WorkspaceMemberRepositoryImpl) GetDB() *sqlx.DB {
	return r.db
}
