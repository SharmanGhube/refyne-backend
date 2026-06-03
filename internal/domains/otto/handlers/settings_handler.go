package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/internal/domains/otto/services"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type OttoSettingsHandler struct {
	service services.OttoService
	logger  *zap.Logger
}

func NewOttoSettingsHandler(service services.OttoService) *OttoSettingsHandler {
	return &OttoSettingsHandler{
		service: service,
		logger:  logging.GetHandlerLogger("OttoSettingsHandler"),
	}
}

func (h *OttoSettingsHandler) GetSettings(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *OttoSettingsHandler) UpsertSettings(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.UpdateOttoSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.UpsertSettings(c.Request.Context(), userID, &input)
	if err != nil {
		h.logger.Error("Failed to upsert settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
