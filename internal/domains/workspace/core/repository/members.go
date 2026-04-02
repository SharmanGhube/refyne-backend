package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	workspaceModels "github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// WorkspaceMemberRepository defines operations for workspace members
type WorkspaceMemberRepository interface {
	AddMember(c *gin.Context, member *workspaceModels.WorkspaceMember) *errors.AppError
	GetMember(c *gin.Context, workspaceID, userID string) (*workspaceModels.WorkspaceMember, *errors.AppError)
	GetWorkspaceMembers(c *gin.Context, workspaceID string) ([]*workspaceModels.WorkspaceMember, *errors.AppError)
	RemoveMember(c *gin.Context, workspaceID, userID string) *errors.AppError
	GetDB() *sqlx.DB
}

// WorkspaceMemberRepositoryImpl implements WorkspaceMemberRepository
type WorkspaceMemberRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

// NewWorkspaceMemberRepository creates a new workspace member repository
func NewWorkspaceMemberRepository(db *sqlx.DB) WorkspaceMemberRepository {
	return &WorkspaceMemberRepositoryImpl{
		name:   "WorkspaceMemberRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("WorkspaceMemberRepository"),
	}
}

// AddMember adds a new member to a workspace
func (r *WorkspaceMemberRepositoryImpl) AddMember(c *gin.Context, member *workspaceModels.WorkspaceMember) *errors.AppError {
	query := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if err := r.db.ExecContext(c, query,
		member.ID,
		member.WorkspaceID,
		member.UserID,
		member.Role,
		member.JoinedAt,
		member.CreatedAt,
		member.UpdatedAt,
	); err != nil {
		r.logger.Error("Failed to add member to workspace",
			zap.Error(err),
			zap.String("workspace_id", member.WorkspaceID),
			zap.String("user_id", member.UserID),
		)
		return errors.NewAppError(
			c,
			"ADD_MEMBER_FAILED",
			"Failed to add member to workspace",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	return nil
}

// GetMember retrieves a specific workspace member
func (r *WorkspaceMemberRepositoryImpl) GetMember(c *gin.Context, workspaceID, userID string) (*workspaceModels.WorkspaceMember, *errors.AppError) {
	member := &workspaceModels.WorkspaceMember{}

	query := `
		SELECT id, workspace_id, user_id, role, joined_at, created_at, updated_at
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`

	if err := r.db.GetContext(c, member, query, workspaceID, userID); err != nil {
		r.logger.Debug("Member not found in workspace",
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, nil
	}

	return member, nil
}

// GetWorkspaceMembers retrieves all members of a workspace
func (r *WorkspaceMemberRepositoryImpl) GetWorkspaceMembers(c *gin.Context, workspaceID string) ([]*workspaceModels.WorkspaceMember, *errors.AppError) {
	members := []*workspaceModels.WorkspaceMember{}

	query := `
		SELECT id, workspace_id, user_id, role, joined_at, created_at, updated_at
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY role DESC, joined_at ASC
	`

	if err := r.db.SelectContext(c, &members, query, workspaceID); err != nil {
		r.logger.Error("Failed to get workspace members",
			zap.Error(err),
			zap.String("workspace_id", workspaceID),
		)
		return nil, errors.NewAppError(
			c,
			"GET_MEMBERS_FAILED",
			"Failed to retrieve workspace members",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	return members, nil
}

// RemoveMember removes a member from a workspace
func (r *WorkspaceMemberRepositoryImpl) RemoveMember(c *gin.Context, workspaceID, userID string) *errors.AppError {
	query := `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(c, query, workspaceID, userID)
	if err != nil {
		r.logger.Error("Failed to remove member from workspace",
			zap.Error(err),
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
		)
		return errors.NewAppError(
			c,
			"REMOVE_MEMBER_FAILED",
			"Failed to remove member from workspace",
			errors.ErrorTypeDatabase,
			errors.SeverityMedium,
			"workspace",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"MEMBER_NOT_FOUND",
			"Member not found in workspace",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"workspace",
		)
	}

	return nil
}

// GetDB returns the database connection
func (r *WorkspaceMemberRepositoryImpl) GetDB() *sqlx.DB {
	return r.db
}
