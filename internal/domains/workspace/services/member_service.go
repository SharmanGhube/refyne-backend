package services

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/core/repository"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type MemberService interface {
	InviteMember(c *gin.Context, workspaceID, userID, email string) *errors.AppError
	RemoveMember(c *gin.Context, workspaceID, ownerID, memberID string) *errors.AppError
	ListMembers(c *gin.Context, workspaceID, userID string) ([]*models.WorkspaceMember, *errors.AppError)
	GetMember(c *gin.Context, workspaceID, userID string) (*models.WorkspaceMember, *errors.AppError)
}

type MemberServiceImpl struct {
	name       string
	logger     *zap.Logger
	wsRepo     repository.WorkspaceRepository
	memberRepo repository.WorkspaceMemberRepository
}

func NewMemberService(wsRepo repository.WorkspaceRepository, memberRepo repository.WorkspaceMemberRepository) MemberService {
	return &MemberServiceImpl{
		name:       "MemberService",
		logger:     logging.GetServiceLogger("MemberService"),
		wsRepo:     wsRepo,
		memberRepo: memberRepo,
	}
}

func (s *MemberServiceImpl) InviteMember(c *gin.Context, workspaceID, userID, email string) *errors.AppError {
	ws, appErr := s.wsRepo.GetWorkspaceByID(c, workspaceID)
	if appErr != nil {
		return appErr
	}

	if ws == nil {
		return errors.NewAppError(c, "WORKSPACE_NOT_FOUND", "Workspace not found",
			errors.ErrorTypeNotFound, errors.SeverityLow, "workspace")
	}

	// Verify requester is owner
	if !ws.IsOwner(userID) {
		return errors.NewAppError(c, "ACCESS_DENIED", "Only workspace owner can invite members",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	// TODO: Implement email invitation via River queue
	s.logger.Info("Member invitation queued",
		zap.String("workspace_id", workspaceID),
		zap.String("email", email),
	)

	return nil
}

func (s *MemberServiceImpl) RemoveMember(c *gin.Context, workspaceID, ownerID, memberID string) *errors.AppError {
	ws, appErr := s.wsRepo.GetWorkspaceByID(c, workspaceID)
	if appErr != nil {
		return appErr
	}

	if ws == nil {
		return errors.NewAppError(c, "WORKSPACE_NOT_FOUND", "Workspace not found",
			errors.ErrorTypeNotFound, errors.SeverityLow, "workspace")
	}

	// Verify requester is owner
	if !ws.IsOwner(ownerID) {
		return errors.NewAppError(c, "ACCESS_DENIED", "Only workspace owner can remove members",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	// Prevent owner from removing themselves
	if ownerID == memberID {
		return errors.NewAppError(c, "CANNOT_REMOVE_OWNER", "Cannot remove workspace owner",
			errors.ErrorTypeValidation, errors.SeverityLow, "workspace")
	}

	return s.memberRepo.RemoveMember(c, workspaceID, memberID)
}

func (s *MemberServiceImpl) ListMembers(c *gin.Context, workspaceID, userID string) ([]*models.WorkspaceMember, *errors.AppError) {
	// Verify user is member of workspace
	member, appErr := s.GetMember(c, workspaceID, userID)
	if appErr != nil {
		return nil, appErr
	}

	if member == nil {
		return nil, errors.NewAppError(c, "ACCESS_DENIED", "You are not a member of this workspace",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	return s.memberRepo.GetWorkspaceMembers(c, workspaceID)
}

func (s *MemberServiceImpl) GetMember(c *gin.Context, workspaceID, userID string) (*models.WorkspaceMember, *errors.AppError) {
	return s.memberRepo.GetMember(c, workspaceID, userID)
}
