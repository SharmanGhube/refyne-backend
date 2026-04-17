package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/internal/domains/otto/repository"
	"github.com/refynehq/refyne-backend/internal/domains/otto/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// OttoHandler handles Otto AI assistant endpoints
type OttoHandler struct {
	conversationService services.OttoConversationManager
	messageRepo         repository.OttoMessageRepository
	logger              *zap.Logger
}

// NewOttoHandler creates a new Otto handler
func NewOttoHandler(
	conversationService services.OttoConversationManager,
	messageRepo repository.OttoMessageRepository,
) *OttoHandler {
	return &OttoHandler{
		conversationService: conversationService,
		messageRepo:         messageRepo,
		logger:              logging.GetHandlerLogger("OttoHandler"),
	}
}

// CreateConversation creates a new conversation
// POST /api/otto/conversations?workspace_id=...
func (h *OttoHandler) CreateConversation(c *gin.Context) {
	type CreateRequest struct {
		Title       string                   `json:"title" binding:"required,max=255"`
		Description string                   `json:"description" binding:"max=1000"`
		Context     models.ConversationContext `json:"context" binding:"required"`
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("userID")
	workspaceID := c.Query("workspace_id")

	if userID == "" || workspaceID == "" {
		c.JSON(400, gin.H{"error": "workspace_id query parameter is required"})
		return
	}

	conversation, err := h.conversationService.CreateConversation(
		c,
		userID,
		workspaceID,
		req.Title,
		&req.Context,
	)

	if err != nil {
		h.logger.Error("Failed to create conversation", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to create conversation"})
		return
	}

	h.logger.Info("Conversation created", zap.String("conversation_id", conversation.ID), zap.String("user_id", userID))

	c.JSON(201, gin.H{
		"status": "ok",
		"data": gin.H{
			"id":             conversation.ID,
			"workspace_id":   conversation.WorkspaceID,
			"title":          conversation.Title,
			"status":         conversation.Status,
			"message_count":  conversation.MessageCount,
			"created_at":     conversation.CreatedAt,
		},
	})
}

// ListConversations lists all conversations for the user
// GET /api/otto/conversations?workspace_id=...&limit=20&offset=0
func (h *OttoHandler) ListConversations(c *gin.Context) {
	userID := c.GetString("userID")
	workspaceID := c.Query("workspace_id")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	if userID == "" || workspaceID == "" {
		c.JSON(400, gin.H{"error": "workspace_id query parameter is required"})
		return
	}

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	conversations, err := h.conversationService.ListConversations(c, userID, workspaceID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list conversations", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to list conversations"})
		return
	}

	data := make([]gin.H, 0)
	for _, conv := range conversations {
		data = append(data, gin.H{
			"id":              conv.ID,
			"title":           conv.Title,
			"status":          conv.Status,
			"is_bookmarked":   conv.IsBookmarked,
			"message_count":   conv.MessageCount,
			"last_message_at": conv.LastMessageAt,
			"created_at":      conv.CreatedAt,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   data,
	})
}

// GetConversation gets a specific conversation
// GET /api/otto/conversations/:id
func (h *OttoHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	if conversationID == "" {
		c.JSON(400, gin.H{"error": "Conversation ID is required"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		h.logger.Warn("Conversation not found", zap.String("conversation_id", conversationID))
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		h.logger.Warn("Unauthorized access to conversation", zap.String("user_id", userID), zap.String("conversation_id", conversationID))
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"id":              conversation.ID,
			"title":           conversation.Title,
			"description":     conversation.Description.String,
			"status":          conversation.Status,
			"is_bookmarked":   conversation.IsBookmarked,
			"message_count":   conversation.MessageCount,
			"last_message_at": conversation.LastMessageAt,
			"created_at":      conversation.CreatedAt,
		},
	})
}

// UpdateConversation updates a conversation
// PUT /api/otto/conversations/:id
func (h *OttoHandler) UpdateConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	var req models.UpdateOttoConversationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	if err := h.conversationService.UpdateConversation(c, conversationID, &req); err != nil {
		h.logger.Error("Failed to update conversation", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to update conversation"})
		return
	}

	h.logger.Info("Conversation updated", zap.String("conversation_id", conversationID))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Conversation updated",
	})
}

// ArchiveConversation archives a conversation
// POST /api/otto/conversations/:id/archive
func (h *OttoHandler) ArchiveConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	if err := h.conversationService.ArchiveConversation(c, conversationID); err != nil {
		h.logger.Error("Failed to archive conversation", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to archive conversation"})
		return
	}

	h.logger.Info("Conversation archived", zap.String("conversation_id", conversationID))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Conversation archived",
	})
}

// BookmarkConversation toggles bookmark status
// POST /api/otto/conversations/:id/bookmark
func (h *OttoHandler) BookmarkConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	type BookmarkRequest struct {
		IsBookmarked bool `json:"is_bookmarked" binding:"required"`
	}

	var req BookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	if err := h.conversationService.BookmarkConversation(c, conversationID, req.IsBookmarked); err != nil {
		h.logger.Error("Failed to update bookmark", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to update bookmark"})
		return
	}

	h.logger.Info("Bookmark toggled", zap.String("conversation_id", conversationID), zap.Bool("is_bookmarked", req.IsBookmarked))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Bookmark updated",
	})
}

// DeleteConversation deletes a conversation
// DELETE /api/otto/conversations/:id
func (h *OttoHandler) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	if err := h.conversationService.DeleteConversation(c, conversationID); err != nil {
		h.logger.Error("Failed to delete conversation", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to delete conversation"})
		return
	}

	h.logger.Info("Conversation deleted", zap.String("conversation_id", conversationID))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Conversation deleted",
	})
}

// SendMessage sends a message in a conversation
// POST /api/otto/conversations/:id/messages
func (h *OttoHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	type MessageRequest struct {
		Content string `json:"content" binding:"required,max=5000"`
	}

	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Create user message
	message := models.NewOttoMessage(conversationID, userID, "user", req.Content)
	if err := h.messageRepo.CreateMessage(c, message); err != nil {
		h.logger.Error("Failed to create message", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to send message"})
		return
	}

	h.logger.Info("Message sent", zap.String("message_id", message.ID), zap.String("conversation_id", conversationID))

	c.JSON(201, gin.H{
		"status": "ok",
		"data": gin.H{
			"id":              message.ID,
			"conversation_id": message.ConversationID,
			"role":            message.Role,
			"content":         message.Content,
			"created_at":      message.CreatedAt,
		},
	})
}

// GetMessages retrieves messages from a conversation
// GET /api/otto/conversations/:id/messages
func (h *OttoHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	limit := 50
	offset := 0
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	messages, err := h.messageRepo.GetConversationMessages(c, conversationID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get messages", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	data := make([]gin.H, 0)
	for _, msg := range messages {
		likeStatus := (*bool)(nil)
		if msg.IsLiked.Valid {
			likeStatus = &msg.IsLiked.Bool
		}
		data = append(data, gin.H{
			"id":              msg.ID,
			"role":            msg.Role,
			"content":         msg.Content,
			"tokens_used":     msg.TokensUsed,
			"model_used":      msg.ModelUsed,
			"is_liked":        likeStatus,
			"created_at":      msg.CreatedAt,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   data,
	})
}

// AddMessageFeedback adds feedback to a message
// POST /api/otto/messages/:id/feedback
func (h *OttoHandler) AddMessageFeedback(c *gin.Context) {
	messageID := c.Param("id")

	type FeedbackRequest struct {
		IsLiked bool   `json:"is_liked" binding:"required"`
		Notes   string `json:"notes" binding:"max=500"`
	}

	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.messageRepo.AddFeedback(c, messageID, req.IsLiked, req.Notes); err != nil {
		h.logger.Error("Failed to add feedback", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to add feedback"})
		return
	}

	h.logger.Info("Feedback added", zap.String("message_id", messageID), zap.Bool("is_liked", req.IsLiked))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Feedback recorded",
	})
}

// GetConversationContext retrieves enriched context for a conversation
// GET /api/otto/conversations/:id/context
func (h *OttoHandler) GetConversationContext(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetString("userID")

	conversation, err := h.conversationService.GetConversation(c, conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	// Verify ownership
	if conversation.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Parse context JSON
	var context models.ConversationContext
	if conversation.Context != "" {
		if err := json.Unmarshal([]byte(conversation.Context), &context); err != nil {
			h.logger.Warn("Failed to parse context JSON", zap.Error(err))
		}
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id":        context.AccountID,
			"platform_type":     context.PlatformType,
			"metrics_scope":     context.MetricsScope,
			"include_media":     context.IncludeMedia,
			"include_insights":  context.IncludeInsights,
			"related_documents": context.RelatedDocuments,
		},
	})
}
