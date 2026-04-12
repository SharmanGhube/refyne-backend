package services

import (
	"context"
	"encoding/json"
	"fmt"
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
	FetchMedia(ctx context.Context, accountID string, syncType string) ([]*models.InstagramMedia, error)

	// GetCachedMedia retrieves cached media from Redis
	GetCachedMedia(ctx context.Context, accountID string) ([]*models.InstagramMedia, error)

	// CacheMedia stores media in Redis with 1-hour TTL
	CacheMedia(ctx context.Context, accountID string, media []*models.InstagramMedia) error

	// FetchMediaInsights fetches engagement metrics for media
	FetchMediaInsights(ctx context.Context, accountID, mediaID string) (*models.MediaInsights, error)
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
func (s *instagramMediaService) FetchMedia(ctx context.Context, accountID string, syncType string) ([]*models.InstagramMedia, error) {
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

	// Get account to fetch access token
	// Note: In real implementation, we would retrieve the account from repository
	// For now, we'll return a placeholder error
	s.logger.Info("Fetching media from Instagram",
		zap.String("account_id", accountID),
		zap.String("sync_type", syncType),
	)

	// Record API call for rate limiting
	if err := s.rateLimiter.RecordCall(ctx, accountID); err != nil {
		s.logger.Error("Failed to record API call", zap.Error(err))
		return nil, err
	}

	// TODO: Implement actual media fetching
	// 1. Decrypt access token from database
	// 2. Call Instagram API based on sync_type:
	//    - "new": GET /me/media?fields=id,caption,media_type,media_url,timestamp,like.limit(0),comments.limit(0)
	//    - "full": Paginate through all media (use after/before cursors)
	//    - "insights": GET /media/{id}/insights?metric=impressions,reach,engagement
	// 3. Parse and transform response to InstagramMedia models
	// 4. Return media list

	return nil, fmt.Errorf("media fetching not yet implemented")
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
func (s *instagramMediaService) FetchMediaInsights(ctx context.Context, accountID, mediaID string) (*models.MediaInsights, error) {
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

	// Record API call for rate limiting
	if err := s.rateLimiter.RecordCall(ctx, accountID); err != nil {
		s.logger.Error("Failed to record API call", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Fetching media insights",
		zap.String("account_id", accountID),
		zap.String("media_id", mediaID),
	)

	// TODO: Implement actual insights fetching from Instagram API
	// GET /media/{mediaID}/insights?metric=impressions,reach,engagement,profile_views,likes,comments,saves
	// Parse response and return MediaInsights struct

	return nil, fmt.Errorf("media insights fetching not yet implemented")
}
