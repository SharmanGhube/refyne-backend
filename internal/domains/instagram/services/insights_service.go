package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramInsightsService handles fetching account-level analytics
type InstagramInsightsService interface {
	// FetchAccountInsights fetches account-level metrics from Instagram API
	FetchAccountInsights(ctx context.Context, accountID, accessToken string) (*models.AccountInsights, error)

	// FetchAccountInsightsTrend fetches analytics trend for last N days
	FetchAccountInsightsTrend(ctx context.Context, accountID, accessToken string, days int) ([]*models.AccountInsights, error)
}

type instagramInsightsService struct {
	config      *config.InstagramConfig
	rateLimiter RateLimitChecker
	httpClient  *http.Client
	logger      *zap.Logger
}

// NewInstagramInsightsService creates a new insights service
func NewInstagramInsightsService(
	cfg *config.InstagramConfig,
	rateLimiter RateLimitChecker,
) InstagramInsightsService {
	return &instagramInsightsService{
		config:      cfg,
		rateLimiter: rateLimiter,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logging.GetServiceLogger("InstagramInsightsService"),
	}
}

// insightsAPIResponse represents the response from Instagram insights endpoint
type insightsAPIResponse struct {
	Data []struct {
		Name  string        `json:"name"`
		Period string        `json:"period"`
		Values []struct {
			Value int `json:"value"`
		} `json:"values"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"data"`
}

// FetchAccountInsights fetches account-level metrics from Instagram API
func (s *instagramInsightsService) FetchAccountInsights(ctx context.Context, accountID, accessToken string) (*models.AccountInsights, error) {
	// Check rate limit
	canCall, _, err := s.rateLimiter.CanMakeCall(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to check rate limit", zap.Error(err))
		return nil, err
	}

	if !canCall {
		s.logger.Warn("Rate limit reached, cannot fetch account insights",
			zap.String("account_id", accountID),
		)
		return nil, fmt.Errorf("rate limit reached")
	}

	if accessToken == "" {
		s.logger.Error("Access token is required to fetch insights",
			zap.String("account_id", accountID),
		)
		return nil, fmt.Errorf("access token is required")
	}

	// Build Instagram Graph API endpoint for account insights
	// Metrics: impressions, reach, profile_views, follower_count
	apiURL := "https://graph.instagram.com/me/insights"
	metrics := "impressions,reach,profile_views,follower_count,get_directions_clicks,website_clicks,phone_call_clicks,text_message_clicks"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	q := req.URL.Query()
	q.Add("metric", metrics)
	q.Add("period", "day")
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
		s.logger.Warn("Instagram API returned non-OK status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("instagram api returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp insightsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		s.logger.Error("Failed to decode response", zap.Error(err))
		return nil, err
	}

	// Extract metrics into AccountInsights struct
	insights := &models.AccountInsights{
		AccountID:   accountID,
		MetricDate:  time.Now(),
		CollectedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Parse metrics
	for _, metric := range apiResp.Data {
		if len(metric.Values) == 0 {
			continue
		}

		// Get the most recent value
		value := metric.Values[len(metric.Values)-1].Value

		switch metric.Name {
		case "impressions":
			insights.Impressions = value
		case "reach":
			insights.Reach = value
		case "profile_views":
			insights.ProfileVisits = value
		case "follower_count":
			insights.FollowerCount = value
		}
	}

	// Calculate engagement rate
	if insights.Impressions > 0 {
		engagementRate := float64(insights.Reach) / float64(insights.Impressions) * 100
		insights.EngagementRate = engagementRate
	}

	// Calculate growth rate (simple: profile visits / impressions)
	if insights.Impressions > 0 {
		growthRate := float64(insights.ProfileVisits) / float64(insights.Impressions) * 100
		insights.GrowthRate = growthRate
	}

	// Record API call for rate limiting
	if err := s.rateLimiter.RecordCall(ctx, accountID); err != nil {
		s.logger.Error("Failed to record API call", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Fetched account insights from Instagram",
		zap.String("account_id", accountID),
		zap.Int("impressions", insights.Impressions),
		zap.Int("reach", insights.Reach),
		zap.Int("follower_count", insights.FollowerCount),
	)

	return insights, nil
}

// FetchAccountInsightsTrend fetches analytics trend for last N days
func (s *instagramInsightsService) FetchAccountInsightsTrend(ctx context.Context, accountID, accessToken string, days int) ([]*models.AccountInsights, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}

	if days > 90 {
		days = 90 // Cap at 90 days (Instagram API limit)
	}

	// For simplicity, we'll fetch daily insights
	// In a real production system, you'd want to batch these calls or use a different approach
	var trends []*models.AccountInsights

	// Fetch insights for each day in the range
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)

		insight, err := s.FetchAccountInsights(ctx, accountID, accessToken)
		if err != nil {
			s.logger.Warn("Failed to fetch insights for date",
				zap.String("account_id", accountID),
				zap.String("date", date.Format(time.RFC3339)),
				zap.Error(err),
			)
			continue
		}

		insight.MetricDate = date
		trends = append(trends, insight)
	}

	s.logger.Info("Fetched account insights trend",
		zap.String("account_id", accountID),
		zap.Int("days", days),
		zap.Int("data_points", len(trends)),
	)

	return trends, nil
}
