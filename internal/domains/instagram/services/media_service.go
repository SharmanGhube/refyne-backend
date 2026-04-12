package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramMediaService handles fetching and caching Instagram media
type InstagramMediaService interface {
	// FetchMedia fetches media from Instagram API with rate limiting
	FetchMedia(ctx context.Context, accountID, accessToken, syncType string) ([]*models.InstagramMedia, error)

	// GetCachedMedia retrieves cached media from Redis
	GetCachedMedia(ctx context.Context, accountID string) ([]*models.InstagramMedia, error)

	// CacheMedia stores media in Redis with 1-hour TTL
	CacheMedia(ctx context.Context, accountID string, media []*models.InstagramMedia) error

	// FetchMediaInsights fetches engagement metrics for media
	FetchMediaInsights(ctx context.Context, accountID, accessToken, mediaID string) (*models.MediaInsights, error)
}

type instagramMediaService struct {
	config      *config.InstagramConfig
	oauthSvc    InstagramOAuthService
	rateLimiter RateLimitChecker
	redis       *redis.Client
	httpClient  *http.Client
	logger      *zap.Logger
}

// NewInstagramMediaService creates a new media service
func NewInstagramMediaService(
	cfg *config.InstagramConfig,
	oauthSvc InstagramOAuthService,
	rateLimiter RateLimitChecker,
	redis *redis.Client,
) InstagramMediaService {
	return &instagramMediaService{
		config:      cfg,
		oauthSvc:    oauthSvc,
		rateLimiter: rateLimiter,
		redis:       redis,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logging.GetServiceLogger("InstagramMediaService"),
	}
}

// mediaAPIResponse represents the response from Instagram media endpoint
type mediaAPIResponse struct {
	Data   []mediaItem `json:"data"`
	Paging struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"paging"`
}

type mediaItem struct {
	ID        string `json:"id"`
	Caption   string `json:"caption"`
	MediaType string `json:"media_type"` // PHOTO, VIDEO, CAROUSEL, REELS, STORY
	MediaURL  string `json:"media_url"`
	Permalink string `json:"permalink"`
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	Like      struct {
		Data []interface{} `json:"data"`
	} `json:"like"`
	Comments struct {
		Data []interface{} `json:"data"`
	} `json:"comments"`
}

// FetchMedia fetches media from Instagram API with rate limiting
func (s *instagramMediaService) FetchMedia(ctx context.Context, accountID, accessToken, syncType string) ([]*models.InstagramMedia, error) {
	// Check rate limit
	canCall, _, err := s.rateLimiter.CanMakeCall(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to check rate limit", zap.Error(err))
		return nil, err
	}

	if !canCall {
		s.logger.Warn("Rate limit reached, cannot fetch media",
			zap.String("account_id", accountID),
		)
		return nil, fmt.Errorf("rate limit reached")
	}

	if accessToken == "" {
		s.logger.Error("Access token is required to fetch media",
			zap.String("account_id", accountID),
		)
		return nil, fmt.Errorf("access token is required")
	}

	// Build Instagram Graph API endpoint
	apiURL := "https://graph.instagram.com/me/media"
	fields := "id,caption,media_type,media_url,timestamp,like.limit(0).summary(true),comments.limit(0).summary(true),insights.metric(impressions,engagement)"

	// Make API request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	q := req.URL.Query()
	q.Add("fields", fields)
	q.Add("access_token", accessToken)
	q.Add("limit", "25") // Paginated results
	req.URL.RawQuery = q.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to call Instagram API", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Instagram API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("instagram api returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp mediaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		s.logger.Error("Failed to decode response", zap.Error(err))
		return nil, err
	}

	// Transform to our models
	var media []*models.InstagramMedia
	for _, item := range apiResp.Data {
		m := &models.InstagramMedia{
			AccountID:        accountID,
			InstagramMediaID: item.ID,
			Caption: sql.NullString{
				String: item.Caption,
				Valid:  item.Caption != "",
			},
			MediaType:    item.MediaType,
			MediaURL:     item.MediaURL,
			Permalink:    sql.NullString{String: item.Permalink, Valid: true},
			PostedAt:     parseTimestamp(item.Timestamp),
			LikeCount:    len(item.Like.Data),
			CommentCount: len(item.Comments.Data),
			SyncedAt:     time.Now(),
		}
		media = append(media, m)
	}

	// Record API call for rate limiting
	if err := s.rateLimiter.RecordCall(ctx, accountID); err != nil {
		s.logger.Error("Failed to record API call", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Fetched media from Instagram",
		zap.String("account_id", accountID),
		zap.Int("count", len(media)),
		zap.String("sync_type", syncType),
	)

	return media, nil
}

// GetCachedMedia retrieves cached media from Redis
func (s *instagramMediaService) GetCachedMedia(ctx context.Context, accountID string) ([]*models.InstagramMedia, error) {
	cacheKey := fmt.Sprintf("instagram:media:%s", accountID)

	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			s.logger.Debug("No cached media found", zap.String("account_id", accountID))
			return nil, nil // Cache miss is not an error
		}
		s.logger.Error("Failed to get cached media", zap.Error(err), zap.String("account_id", accountID))
		return nil, err
	}

	var media []*models.InstagramMedia
	if err := json.Unmarshal([]byte(val), &media); err != nil {
		s.logger.Error("Failed to unmarshal cached media", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Retrieved cached media",
		zap.String("account_id", accountID),
		zap.Int("count", len(media)),
	)

	return media, nil
}

// CacheMedia stores media in Redis with 1-hour TTL
func (s *instagramMediaService) CacheMedia(ctx context.Context, accountID string, media []*models.InstagramMedia) error {
	cacheKey := fmt.Sprintf("instagram:media:%s", accountID)

	data, err := json.Marshal(media)
	if err != nil {
		s.logger.Error("Failed to marshal media for caching", zap.Error(err))
		return err
	}

	if err := s.redis.Set(ctx, cacheKey, data, time.Hour).Err(); err != nil {
		s.logger.Error("Failed to cache media", zap.Error(err), zap.String("account_id", accountID))
		return err
	}

	s.logger.Debug("Cached media",
		zap.String("account_id", accountID),
		zap.Int("count", len(media)),
	)

	return nil
}

// FetchMediaInsights fetches engagement metrics for media
func (s *instagramMediaService) FetchMediaInsights(ctx context.Context, accountID, accessToken, mediaID string) (*models.MediaInsights, error) {
	// Check rate limit
	canCall, _, err := s.rateLimiter.CanMakeCall(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to check rate limit", zap.Error(err))
		return nil, err
	}

	if !canCall {
		s.logger.Warn("Rate limit reached, cannot fetch insights",
			zap.String("account_id", accountID),
			zap.String("media_id", mediaID),
		)
		return nil, fmt.Errorf("rate limit reached")
	}

	if accessToken == "" {
		s.logger.Error("Access token is required to fetch insights",
			zap.String("account_id", accountID),
		)
		return nil, fmt.Errorf("access token is required")
	}

	// Build Instagram Graph API endpoint for insights
	apiURL := fmt.Sprintf("https://graph.instagram.com/%s/insights", mediaID)
	metrics := "impressions,reach,engagement,clicks,saves,video_views"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	q := req.URL.Query()
	q.Add("metric", metrics)
	q.Add("access_token", accessToken)
	req.URL.RawQuery = q.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to call Instagram insights API", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Warn("Instagram insights API returned non-OK status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		// Return nil for insights if not available (some media types may not have insights)
		return nil, nil
	}

	// Parse response
	var insightsResp struct {
		Data []struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&insightsResp); err != nil {
		s.logger.Error("Failed to decode insights response", zap.Error(err))
		return nil, err
	}

	// Extract metrics into MediaInsights struct
	insights := &models.MediaInsights{
		MediaID:     mediaID,
		AccountID:   accountID,
		MetricDate:  time.Now(),
		CollectedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Parse metrics
	for _, item := range insightsResp.Data {
		switch item.Name {
		case "impressions":
			insights.Impressions = item.Value
		case "reach":
			insights.Reach = item.Value
		case "engagement":
			insights.Shares = item.Value // Using shares field for engagement
		case "clicks":
			insights.Clicks = item.Value
		case "saves":
			insights.Saves = item.Value
		}
	}

	// Calculate engagement rate
	if insights.Impressions > 0 {
		engagementRate := float64(insights.Shares) / float64(insights.Impressions) * 100
		insights.EngagementRate = engagementRate
	}

	// Record API call for rate limiting
	if err := s.rateLimiter.RecordCall(ctx, accountID); err != nil {
		s.logger.Error("Failed to record API call", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Fetched media insights from Instagram",
		zap.String("account_id", accountID),
		zap.String("media_id", mediaID),
		zap.Int("impressions", insights.Impressions),
		zap.Int("reach", insights.Reach),
	)

	return insights, nil
}

// Helper function to parse Instagram timestamp
func parseTimestamp(timestamp string) time.Time {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// If parsing fails, return current time
		return time.Now()
	}
	return t
}
