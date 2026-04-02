package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/models"
	"github.com/refynehq/refyne-backend/internal/domains/workspace/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type WorkspaceHandler interface {
	CreateWorkspace(c *gin.Context)
	GetWorkspace(c *gin.Context)
	ListWorkspaces(c *gin.Context)
	UpdateWorkspace(c *gin.Context)
	DeleteWorkspace(c *gin.Context)
	ListMembers(c *gin.Context)
	InviteMember(c *gin.Context)
	RemoveMember(c *gin.Context)
}

type WorkspaceHandlerImpl struct {
	logger        *zap.Logger
	wsService     services.WorkspaceService
	memberService services.MemberService
}

func NewWorkspaceHandler(wsService services.WorkspaceService, memberService services.MemberService) WorkspaceHandler {
	return &WorkspaceHandlerImpl{
		logger:        logging.GetHandlerLogger("WorkspaceHandler"),
		wsService:     wsService,
		memberService: memberService,
	}
}

func (h *WorkspaceHandlerImpl) CreateWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	ws, appErr := h.wsService.CreateWorkspace(c, userID, req.Name, req.Description)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusCreated, ws)
}

func (h *WorkspaceHandlerImpl) GetWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	ws, appErr := h.wsService.GetWorkspace(c, userID, workspaceID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, ws)
}

func (h *WorkspaceHandlerImpl) ListWorkspaces(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaces, appErr := h.wsService.ListWorkspaces(c, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	if workspaces == nil {
		workspaces = []*models.Workspace{}
	}

	c.JSON(http.StatusOK, workspaces)
}

func (h *WorkspaceHandlerImpl) UpdateWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	ws, appErr := h.wsService.UpdateWorkspace(c, userID, workspaceID, &req)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, ws)
}

func (h *WorkspaceHandlerImpl) DeleteWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	if appErr := h.wsService.DeleteWorkspace(c, userID, workspaceID); appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted"})
}

func (h *WorkspaceHandlerImpl) ListMembers(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	members, appErr := h.memberService.ListMembers(c, workspaceID, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	if members == nil {
		members = []*models.WorkspaceMember{}
	}

	c.JSON(http.StatusOK, members)
}

func (h *WorkspaceHandlerImpl) InviteMember(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if appErr := h.memberService.InviteMember(c, workspaceID, userID, req.Email); appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent"})
}

func (h *WorkspaceHandlerImpl) RemoveMember(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Param("id")
	memberID := c.Param("memberId")

	if appErr := h.memberService.RemoveMember(c, workspaceID, userID, memberID); appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}
