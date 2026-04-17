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
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	var req models.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	ws, appErr := h.wsService.CreateWorkspace(c, userID, req.Name, req.Description)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusCreated, "Workspace created successfully", ws)
}

func (h *WorkspaceHandlerImpl) GetWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	ws, appErr := h.wsService.GetWorkspace(c, userID, workspaceID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Workspace retrieved successfully", ws)
}

func (h *WorkspaceHandlerImpl) ListWorkspaces(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaces, appErr := h.wsService.ListWorkspaces(c, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	if workspaces == nil {
		workspaces = []*models.Workspace{}
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Workspaces retrieved successfully", gin.H{
		"workspaces": workspaces,
	})
}

func (h *WorkspaceHandlerImpl) UpdateWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	ws, appErr := h.wsService.UpdateWorkspace(c, userID, workspaceID, &req)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Workspace updated successfully", ws)
}

func (h *WorkspaceHandlerImpl) DeleteWorkspace(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	if appErr := h.wsService.DeleteWorkspace(c, userID, workspaceID); appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Workspace deleted successfully", gin.H{
		"status": "deleted",
	})
}

func (h *WorkspaceHandlerImpl) ListMembers(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	members, appErr := h.memberService.ListMembers(c, workspaceID, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	if members == nil {
		members = []*models.WorkspaceMember{}
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Members retrieved successfully", gin.H{
		"members": members,
	})
}

func (h *WorkspaceHandlerImpl) InviteMember(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	if appErr := h.memberService.InviteMember(c, workspaceID, userID, req.Email); appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Invitation sent successfully", gin.H{
		"status": "invited",
	})
}

func (h *WorkspaceHandlerImpl) RemoveMember(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	workspaceID := c.Param("id")
	memberID := c.Param("memberId")

	if appErr := h.memberService.RemoveMember(c, workspaceID, userID, memberID); appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Member removed successfully", gin.H{
		"status": "removed",
	})
}
