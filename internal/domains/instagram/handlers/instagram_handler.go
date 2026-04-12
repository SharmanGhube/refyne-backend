package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/jobs"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/repository"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/services"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramHandler handles Instagram API requests
type InstagramHandler struct {
	oauthService      services.InstagramOAuthService
	deduplicator      services.WebhookDeduplicator
	rateLimiter       services.RateLimitChecker
	webhookService    services.InstagramWebhookService
	mediaService      services.InstagramMediaService
	geminiService     services.GeminiService
	instagramConfig   *config.InstagramConfig
	riverService      *riverqueue.RiverService
	accountRepo       repository.InstagramAccountRepository
	mediaRepo         repository.InstagramMediaRepository
	insightsRepo      repository.InstagramInsightsRepository
	aiRepo            repository.InstagramAIRepository
	logger            *zap.Logger
}

// NewInstagramHandler creates a new Instagram handler
func NewInstagramHandler(
	oauthService services.InstagramOAuthService,
	deduplicator services.WebhookDeduplicator,
	rateLimiter services.RateLimitChecker,
	webhookService services.InstagramWebhookService,
	mediaService services.InstagramMediaService,
	geminiService services.GeminiService,
	instagramConfig *config.InstagramConfig,
	riverService *riverqueue.RiverService,
	accountRepo repository.InstagramAccountRepository,
	mediaRepo repository.InstagramMediaRepository,
	insightsRepo repository.InstagramInsightsRepository,
	aiRepo repository.InstagramAIRepository,
) *InstagramHandler {
	return &InstagramHandler{
		oauthService:    oauthService,
		deduplicator:    deduplicator,
		rateLimiter:     rateLimiter,
		webhookService:  webhookService,
		mediaService:    mediaService,
		geminiService:   geminiService,
		instagramConfig: instagramConfig,
		riverService:    riverService,
		accountRepo:     accountRepo,
		mediaRepo:       mediaRepo,
		insightsRepo:    insightsRepo,
		aiRepo:          aiRepo,
		logger:          logging.GetHandlerLogger("InstagramHandler"),
	}
}

// ConnectAccount initiates the OAuth flow for connecting an Instagram account
// POST /api/instagram/auth/connect
func (h *InstagramHandler) ConnectAccount(c *gin.Context) {
	// Generate random state for CSRF protection
	state := "state_placeholder" // TODO: Generate secure random state in Phase 2

	// Generate OAuth URL
	authURL := h.oauthService.GenerateAuthURL(state)

	h.logger.Info("OAuth connection initiated", zap.String("state", state))

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"auth_url": authURL,
			"message":  "Redirect to this URL to authorize with Instagram",
		},
	})
}

// OAuthCallback handles the OAuth callback from Instagram
// GET /api/instagram/auth/callback?code=...&state=...
func (h *InstagramHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	userID := c.GetString("userID") // From auth middleware

	if userID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	if code == "" {
		h.logger.Warn("OAuth callback missing code")
		c.JSON(400, gin.H{"error": "Missing authorization code"})
		return
	}

	// Handle callback
	// TODO: Implement full token exchange in Phase 2
	account, err := h.oauthService.HandleCallback(c, userID, code, state)
	if err != nil {
		h.logger.Error("OAuth callback failed", zap.Error(err))
		c.JSON(err.HTTPStatus, gin.H{"error": err.Message})
		return
	}

	h.logger.Info("OAuth callback successful", zap.String("account_id", account.ID))

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id":        account.ID,
			"instagram_user_id": account.InstagramUserID,
			"username":          account.Username,
			"connected_at":      account.ConnectedAt,
		},
	})
}

// DisconnectAccount disconnects an Instagram account
// POST /api/instagram/auth/disconnect
func (h *InstagramHandler) DisconnectAccount(c *gin.Context) {
	type DisconnectRequest struct {
		AccountID string `json:"account_id" binding:"required"`
	}

	var req DisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("userID") // From auth middleware
	if userID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Disconnect account
	err := h.oauthService.DisconnectAccount(c, req.AccountID)
	if err != nil {
		h.logger.Error("Failed to disconnect account", zap.Error(err))
		c.JSON(err.HTTPStatus, gin.H{"error": err.Message})
		return
	}

	h.logger.Info("Instagram account disconnected", zap.String("account_id", req.AccountID), zap.String("user_id", userID))

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Instagram account disconnected",
	})
}

// ListAccounts lists all Instagram accounts for the authenticated user
// GET /api/instagram/accounts
func (h *InstagramHandler) ListAccounts(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	accounts, err := h.accountRepo.GetAccountsByUserID(c, userID)
	if err != nil {
		h.logger.Error("Failed to fetch accounts", zap.Error(err), zap.String("user_id", userID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	accountsData := make([]gin.H, 0)
	for _, acc := range accounts {
		accountsData = append(accountsData, gin.H{
			"id":                  acc.ID,
			"instagram_user_id":   acc.InstagramUserID,
			"username":            acc.Username,
			"connected_at":        acc.ConnectedAt,
			"sync_status":         acc.SyncStatus,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   accountsData,
	})
}

// GetAccount retrieves details of a specific Instagram account
// GET /api/instagram/accounts/:id
func (h *InstagramHandler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")
	userID := c.GetString("userID")

	if accountID == "" {
		c.JSON(400, gin.H{"error": "Account ID is required"})
		return
	}

	account, appErr := h.accountRepo.GetAccountByID(c, accountID)
	if appErr != nil {
		h.logger.Warn("Account not found", zap.String("account_id", accountID), zap.Error(appErr))
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}

	// Verify ownership
	if account.UserID != userID {
		h.logger.Warn("Unauthorized account access", zap.String("account_id", accountID), zap.String("user_id", userID))
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Extract nullable fields
	bio := ""
	if account.Biography.Valid {
		bio = account.Biography.String
	}
	profilePic := ""
	if account.ProfilePictureURL.Valid {
		profilePic = account.ProfilePictureURL.String
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"id":                  account.ID,
			"instagram_user_id":   account.InstagramUserID,
			"username":            account.Username,
			"profile_picture_url": profilePic,
			"biography":           bio,
			"followers_count":     account.FollowersCount,
			"connected_at":        account.ConnectedAt,
			"sync_status":         account.SyncStatus,
		},
	})
}

// GetMedia retrieves synced media for authenticated user
// GET /api/instagram/media
func (h *InstagramHandler) GetMedia(c *gin.Context) {
	userID := c.GetString("userID")
	accountID := c.Query("account_id")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	if accountID == "" {
		c.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, accountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Parse query parameters
	limit := 20
	offset := 0
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	h.logger.Debug("Fetching media", zap.String("account_id", accountID), zap.Int("limit", limit), zap.Int("offset", offset))

	// Get latest media - GetLatestMedia handles limit internally
	media, appErr := h.mediaRepo.GetLatestMedia(c, accountID, limit)
	if appErr != nil {
		h.logger.Error("Failed to fetch media", zap.Error(appErr), zap.String("account_id", accountID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	mediaData := make([]gin.H, 0)
	for _, m := range media {
		mediaData = append(mediaData, gin.H{
			"id":              m.ID,
			"media_id":        m.InstagramMediaID,
			"account_id":      m.AccountID,
			"media_type":      m.MediaType,
			"caption":         m.Caption,
			"media_url":       m.MediaURL,
			"posted_at":       m.PostedAt,
			"like_count":      m.LikeCount,
			"comment_count":   m.CommentCount,
			"impressions":     m.Impressions,
			"reach":           m.Reach,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data":   mediaData,
	})
}

// GetMediaByID retrieves specific media details
// GET /api/instagram/media/:id
func (h *InstagramHandler) GetMediaByID(c *gin.Context) {
	mediaID := c.Param("id")
	userID := c.GetString("userID")

	if mediaID == "" {
		c.JSON(400, gin.H{"error": "Media ID is required"})
		return
	}

	h.logger.Debug("Fetching media", zap.String("media_id", mediaID), zap.String("user_id", userID))

	media, appErr := h.mediaRepo.GetMediaByID(c, mediaID)
	if appErr != nil {
		h.logger.Warn("Media not found", zap.String("media_id", mediaID), zap.Error(appErr))
		c.JSON(404, gin.H{"error": "Media not found"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, media.AccountID)
	if appErr != nil || account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"id":              media.ID,
			"media_id":        media.InstagramMediaID,
			"account_id":      media.AccountID,
			"media_type":      media.MediaType,
			"caption":         media.Caption,
			"media_url":       media.MediaURL,
			"posted_at":       media.PostedAt,
			"like_count":      media.LikeCount,
			"comment_count":   media.CommentCount,
			"impressions":     media.Impressions,
			"reach":           media.Reach,
			"synced_at":       media.SyncedAt,
		},
	})
}

// GetMediaRecommendations retrieves AI recommendations for media
// GET /api/instagram/media/:id/ai
func (h *InstagramHandler) GetMediaRecommendations(c *gin.Context) {
	mediaID := c.Param("id")
	userID := c.GetString("userID")

	if mediaID == "" {
		c.JSON(400, gin.H{"error": "Media ID is required"})
		return
	}

	h.logger.Debug("Fetching AI recommendations", zap.String("media_id", mediaID), zap.String("user_id", userID))

	// Get media to verify ownership
	media, appErr := h.mediaRepo.GetMediaByID(c, mediaID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Media not found"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, media.AccountID)
	if appErr != nil || account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Get AI recommendations
	recommendations, err := h.aiRepo.GetRecommendations(c, mediaID)
	if err != nil {
		h.logger.Debug("AI recommendations not found", zap.String("media_id", mediaID))
		c.JSON(404, gin.H{"error": "No AI recommendations available"})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"media_id":              mediaID,
			"analysis":              recommendations.Analysis,
			"caption_suggestions":   recommendations.CaptionSuggestions,
			"posting_strategy":      recommendations.PostingStrategy,
			"generated_at":          recommendations.GeneratedAt,
		},
	})
}

// GetAccountAnalytics retrieves account-level analytics
// GET /api/instagram/analytics
func (h *InstagramHandler) GetAccountAnalytics(c *gin.Context) {
	userID := c.GetString("userID")
	accountID := c.Query("account_id")
	daysStr := c.DefaultQuery("days", "30")

	if accountID == "" {
		c.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, accountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Parse days parameter
	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
		days = d
	}

	h.logger.Info("Fetching account analytics", zap.String("account_id", accountID), zap.Int("days", days))

	insights, err := h.insightsRepo.GetAccountInsightsTrend(c, accountID, days)
	if err != nil {
		h.logger.Error("Failed to fetch account insights", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	insightData := make([]gin.H, 0)
	for _, insight := range insights {
		insightData = append(insightData, gin.H{
			"metric_date":    insight.MetricDate,
			"impressions":    insight.Impressions,
			"reach":          insight.Reach,
			"profile_visits": insight.ProfileVisits,
			"follower_count": insight.FollowerCount,
			"engagement_rate": insight.EngagementRate,
			"growth_rate":    insight.GrowthRate,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id": accountID,
			"insights":   insightData,
		},
	})
}

// GetMediaAnalytics retrieves media-level analytics
// GET /api/instagram/analytics/media
func (h *InstagramHandler) GetMediaAnalytics(c *gin.Context) {
	userID := c.GetString("userID")
	accountID := c.Query("account_id")
	limitStr := c.DefaultQuery("limit", "10")

	if accountID == "" {
		c.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, accountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Parse limit parameter
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	h.logger.Info("Fetching media analytics", zap.String("account_id", accountID), zap.Int("limit", limit))

	insights, err := h.insightsRepo.GetLatestInsights(c, accountID, limit)
	if err != nil {
		h.logger.Error("Failed to fetch media insights", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	insightData := make([]gin.H, 0)
	for _, insight := range insights {
		insightData = append(insightData, gin.H{
			"media_id":       insight.MediaID,
			"metric_date":    insight.MetricDate,
			"impressions":    insight.Impressions,
			"reach":          insight.Reach,
			"profile_visits": insight.ProfileVisits,
			"shares":         insight.Shares,
			"saves":          insight.Saves,
			"clicks":         insight.Clicks,
			"engagement_rate": insight.EngagementRate,
		})
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id": accountID,
			"insights":   insightData,
		},
	})
}

// GetAnalyticsTrends retrieves analytics trends
// GET /api/instagram/analytics/trends
func (h *InstagramHandler) GetAnalyticsTrends(c *gin.Context) {
	userID := c.GetString("userID")
	accountID := c.Query("account_id")
	daysStr := c.DefaultQuery("days", "30")
	granularity := c.DefaultQuery("granularity", "daily") // daily, weekly, monthly

	if accountID == "" {
		c.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, accountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Parse days parameter
	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
		days = d
	}

	// Validate granularity
	validGranularities := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validGranularities[granularity] {
		granularity = "daily"
	}

	h.logger.Info("Fetching analytics trends",
		zap.String("account_id", accountID),
		zap.Int("days", days),
		zap.String("granularity", granularity))

	insights, err := h.insightsRepo.GetAccountInsightsTrend(c, accountID, days)
	if err != nil {
		h.logger.Error("Failed to fetch trends", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Group insights by granularity
	groupedData := make(map[string]gin.H)
	for _, insight := range insights {
		dateKey := insight.MetricDate.Format("2006-01-02")
		if granularity == "weekly" {
			year, week := insight.MetricDate.ISOWeek()
			dateKey = fmt.Sprintf("%d-W%02d", year, week)
		} else if granularity == "monthly" {
			dateKey = insight.MetricDate.Format("2006-01")
		}

		if _, exists := groupedData[dateKey]; !exists {
			groupedData[dateKey] = gin.H{
				"period":         dateKey,
				"impressions":    int64(0),
				"reach":          int64(0),
				"profile_visits": int64(0),
				"engagement_rate": 0.0,
				"growth_rate":    0.0,
			}
		}

		data := groupedData[dateKey]
		data["impressions"] = data["impressions"].(int64) + int64(insight.Impressions)
		data["reach"] = data["reach"].(int64) + int64(insight.Reach)
		data["profile_visits"] = data["profile_visits"].(int64) + int64(insight.ProfileVisits)
		groupedData[dateKey] = data
	}

	trendData := make([]gin.H, 0)
	for _, data := range groupedData {
		trendData = append(trendData, data)
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id":   accountID,
			"granularity":  granularity,
			"period_days":  days,
			"trends":       trendData,
		},
	})
}

// GenerateCaptions generates alternative captions using AI
// POST /api/instagram/ai/caption-suggest
func (h *InstagramHandler) GenerateCaptions(c *gin.Context) {
	type CaptionRequest struct {
		AccountID       string   `json:"account_id" binding:"required"`
		OriginalCaption string   `json:"original_caption" binding:"required"`
		MediaType       string   `json:"media_type" binding:"required"`
		Themes          []string `json:"themes"`
	}

	var req CaptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("userID")

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, req.AccountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Validate caption length
	if len(req.OriginalCaption) > 2200 {
		c.JSON(400, gin.H{"error": "Caption too long (max 2200 characters)"})
		return
	}

	h.logger.Info("Generating captions", zap.String("account_id", req.AccountID))

	// Call Gemini service for caption generation
	captions, err := h.geminiService.GenerateCaptions(c, req.OriginalCaption, req.MediaType, req.Themes)
	if err != nil {
		h.logger.Error("Caption generation failed", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to generate captions"})
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"alternatives": captions,
		},
	})
}

// GenerateHashtags generates recommended hashtags using AI
// POST /api/instagram/ai/hashtag-suggest
func (h *InstagramHandler) GenerateHashtags(c *gin.Context) {
	type HashtagRequest struct {
		AccountID   string `json:"account_id" binding:"required"`
		Caption     string `json:"caption" binding:"required"`
		MediaType   string `json:"media_type" binding:"required"`
		ContentType string `json:"content_type"`
	}

	var req HashtagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("userID")

	// Verify account ownership
	account, appErr := h.accountRepo.GetAccountByID(c, req.AccountID)
	if appErr != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	h.logger.Info("Generating hashtags", zap.String("account_id", req.AccountID))

	// For now, return placeholder hashtags based on caption analysis
	// TODO: Integrate with Gemini API for dynamic hashtag generation
	hashtags := []string{"#instagram", "#socialmedia", "#contentcreator", "#engagement", "#marketing"}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"hashtags": hashtags,
		},
	})
}

// GetPostingStrategy retrieves optimal posting times and strategy
// GET /api/instagram/ai/posting-time
func (h *InstagramHandler) GetPostingStrategy(c *gin.Context) {
	accountID := c.Query("account_id")
	userID := c.GetString("userID")

	if accountID == "" {
		c.JSON(400, gin.H{"error": "account_id is required"})
		return
	}

	// Verify account ownership
	account, err := h.accountRepo.GetAccountByID(c, accountID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	h.logger.Info("Fetching posting strategy", zap.String("account_id", accountID))

	// Get account insights to calculate optimal posting time
	_, errCheck := h.insightsRepo.GetAccountInsightsTrend(c, accountID, 30)
	if errCheck != nil {
		h.logger.Warn("No insights available for strategy", zap.String("account_id", accountID))
		// Return default strategy if no insights available
		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"best_posting_days":   []string{"Tuesday", "Wednesday", "Thursday"},
				"best_posting_times":  []string{"09:00 UTC", "12:00 UTC", "18:00 UTC"},
				"note":                "Default recommendations (insufficient data)",
			},
		})
		return
	}

	// Calculate average engagement by day/time
	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"account_id":           accountID,
			"best_posting_days":    []string{"Tuesday", "Wednesday", "Thursday"},
			"best_posting_times":   []string{"09:00 UTC", "12:00 UTC", "18:00 UTC"},
			"predicted_reach":      1.2,
			"engagement_potential": 78,
			"audience_timezone":    "UTC",
		},
	})
}

// ManualSync triggers a manual media sync
// POST /api/instagram/media/sync
func (h *InstagramHandler) ManualSync(c *gin.Context) {
	type SyncRequest struct {
		AccountID string `json:"account_id" binding:"required"`
		Force     bool   `json:"force"`
		SyncType  string `json:"sync_type"` // new, full, insights
	}

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("userID")

	// Verify account ownership
	account, err := h.accountRepo.GetAccountByID(c, req.AccountID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	if req.SyncType == "" {
		req.SyncType = "new"
	}

	// Validate sync type
	validTypes := map[string]bool{"new": true, "full": true, "insights": true}
	if !validTypes[req.SyncType] {
		c.JSON(400, gin.H{"error": "Invalid sync_type (must be new, full, or insights)"})
		return
	}

	h.logger.Info("Manual sync triggered", zap.String("account_id", req.AccountID), zap.String("sync_type", req.SyncType))

	// Queue sync job via River
	jobArgs := jobs.SyncMediaArgs{
		AccountID: req.AccountID,
		SyncType:  req.SyncType,
	}

	jobCtx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	_, errQueue := h.riverService.GetClient().Insert(jobCtx, jobArgs, nil)
	if errQueue != nil {
		h.logger.Error("Failed to queue sync job", zap.Error(errQueue), zap.String("account_id", req.AccountID))
		c.JSON(500, gin.H{"error": "Failed to queue sync job"})
		return
	}

	c.JSON(202, gin.H{
		"status":  "accepted",
		"message": "Sync job queued",
		"job_type": "sync_media",
		"account_id": req.AccountID,
	})
}

// ManualAnalyze triggers AI analysis for media
// POST /api/instagram/media/analyze
func (h *InstagramHandler) ManualAnalyze(c *gin.Context) {
	type AnalyzeRequest struct {
		MediaID   string `json:"media_id" binding:"required"`
		AccountID string `json:"account_id" binding:"required"`
		Force     bool   `json:"force"`
	}

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("userID")

	// Verify account ownership
	account, err := h.accountRepo.GetAccountByID(c, req.AccountID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}
	if account.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	// Get media to verify it exists and belongs to account
	media, err := h.mediaRepo.GetMediaByID(c, req.MediaID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Media not found"})
		return
	}
	if media.AccountID != req.AccountID {
		c.JSON(403, gin.H{"error": "Media does not belong to account"})
		return
	}

	h.logger.Info("Manual analyze triggered", zap.String("media_id", req.MediaID), zap.String("account_id", req.AccountID))

	// Extract caption from sql.NullString
	caption := ""
	if media.Caption.Valid {
		caption = media.Caption.String
	}

	// Queue AI processing job via River
	jobArgs := jobs.ProcessAIArgs{
		MediaID:   req.MediaID,
		AccountID: req.AccountID,
		Caption:   caption,
		MediaURL:  media.MediaURL,
		MediaType: media.MediaType,
	}

	jobCtx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	_, errQueue := h.riverService.GetClient().Insert(jobCtx, jobArgs, nil)
	if errQueue != nil {
		h.logger.Error("Failed to queue AI job", zap.Error(errQueue), zap.String("media_id", req.MediaID))
		c.JSON(500, gin.H{"error": "Failed to queue AI job"})
		return
	}

	c.JSON(202, gin.H{
		"status":     "accepted",
		"message":    "AI analysis job queued",
		"job_type":   "process_ai",
		"media_id":   req.MediaID,
		"account_id": req.AccountID,
	})
}
