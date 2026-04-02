package services

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/core/repository"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type WorkspaceService interface {
	CreateWorkspace(c *gin.Context, userID, name, description string) (*models.Workspace, *errors.AppError)
	GetWorkspace(c *gin.Context, userID, workspaceID string) (*models.Workspace, *errors.AppError)
	ListWorkspaces(c *gin.Context, userID string) ([]*models.Workspace, *errors.AppError)
	UpdateWorkspace(c *gin.Context, userID, workspaceID string, updates *models.UpdateWorkspaceRequest) (*models.Workspace, *errors.AppError)
	DeleteWorkspace(c *gin.Context, userID, workspaceID string) *errors.AppError
}

type WorkspaceServiceImpl struct {
	name       string
	logger     *zap.Logger
	wsRepo     repository.WorkspaceRepository
	memberRepo repository.WorkspaceMemberRepository
}

func NewWorkspaceService(wsRepo repository.WorkspaceRepository, memberRepo repository.WorkspaceMemberRepository) WorkspaceService {
	return &WorkspaceServiceImpl{
		name:       "WorkspaceService",
		logger:     logging.GetServiceLogger("WorkspaceService"),
		wsRepo:     wsRepo,
		memberRepo: memberRepo,
	}
}

func (s *WorkspaceServiceImpl) CreateWorkspace(c *gin.Context, userID, name, description string) (*models.Workspace, *errors.AppError) {
	ws := &models.Workspace{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if appErr := s.wsRepo.CreateWorkspace(c, ws); appErr != nil {
		return nil, appErr
	}

	// Add owner as member
	member := &models.WorkspaceMember{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        "owner",
		JoinedAt:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.memberRepo.AddMember(c, member)

	s.logger.Info("Workspace created", zap.String("id", ws.ID), zap.String("user_id", userID))
	return ws, nil
}

func (s *WorkspaceServiceImpl) GetWorkspace(c *gin.Context, userID, workspaceID string) (*models.Workspace, *errors.AppError) {
	ws, appErr := s.wsRepo.GetWorkspaceByID(c, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	if ws == nil {
		return nil, errors.NewAppError(c, "WORKSPACE_NOT_FOUND", "Workspace not found",
			errors.ErrorTypeNotFound, errors.SeverityLow, "workspace")
	}

	// Verify user is member of workspace
	member, _ := s.memberRepo.GetMember(c, workspaceID, userID)
	if member == nil {
		return nil, errors.NewAppError(c, "ACCESS_DENIED", "You do not have access to this workspace",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	return ws, nil
}

func (s *WorkspaceServiceImpl) ListWorkspaces(c *gin.Context, userID string) ([]*models.Workspace, *errors.AppError) {
	return s.wsRepo.GetUserWorkspaces(c, userID)
}

func (s *WorkspaceServiceImpl) UpdateWorkspace(c *gin.Context, userID, workspaceID string, updates *models.UpdateWorkspaceRequest) (*models.Workspace, *errors.AppError) {
	ws, appErr := s.GetWorkspace(c, userID, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	// Verify user is owner
	if !ws.IsOwner(userID) {
		return nil, errors.NewAppError(c, "ACCESS_DENIED", "Only workspace owner can update workspace",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	if updates.Name != "" {
		ws.Name = updates.Name
	}
	if updates.Description != "" {
		ws.Description = updates.Description
	}
	ws.UpdatedAt = time.Now()

	if appErr := s.wsRepo.UpdateWorkspace(c, ws); appErr != nil {
		return nil, appErr
	}

	return ws, nil
}

func (s *WorkspaceServiceImpl) DeleteWorkspace(c *gin.Context, userID, workspaceID string) *errors.AppError {
	ws, appErr := s.GetWorkspace(c, userID, workspaceID)
	if appErr != nil {
		return appErr
	}

	// Verify user is owner
	if !ws.IsOwner(userID) {
		return errors.NewAppError(c, "ACCESS_DENIED", "Only workspace owner can delete workspace",
			errors.ErrorTypeUnauthorized, errors.SeverityLow, "workspace")
	}

	return s.wsRepo.SoftDeleteWorkspace(c, workspaceID)
}
