package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// GeminiService handles AI analysis using Google's Gemini API
type GeminiService interface {
	// AnalyzeMedia performs comprehensive AI analysis on media
	AnalyzeMedia(ctx context.Context, mediaType, caption, mediaURL string) (*models.AIAnalysisResult, error)

	// GenerateCaptions generates alternative captions for media
	GenerateCaptions(ctx context.Context, originalCaption, mediaType string, themes []string) ([]*models.CaptionSuggestion, error)

	// GeneratePostingStrategy generates optimal posting recommendations
	GeneratePostingStrategy(ctx context.Context, analysis *models.AIAnalysisResult, accountID string) (*models.PostingStrategy, error)
}

type geminiService struct {
	config     *config.InstagramConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewGeminiService creates a new Gemini AI service
func NewGeminiService(cfg *config.InstagramConfig) GeminiService {
	return &geminiService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logging.GetServiceLogger("GeminiService"),
	}
}

// AnalyzeMedia performs comprehensive AI analysis on media
func (s *geminiService) AnalyzeMedia(ctx context.Context, mediaType, caption, mediaURL string) (*models.AIAnalysisResult, error) {
	prompt := fmt.Sprintf(`
Analyze this Instagram %s post and provide structured insights in valid JSON format.

Media URL: %s
Original Caption: "%s"

Provide ONLY valid JSON without any markdown formatting:
{
  "sentiment": "positive|neutral|negative",
  "sentiment_score": 0-10,
  "content_type": "product|lifestyle|educational|entertainment|behind-the-scenes|other",
  "content_themes": ["theme1", "theme2", "theme3"],
  "quality_score": 0-10,
  "engagement_potential": 0-100,
  "recommended_hashtags": ["hashtag1", "hashtag2", "hashtag3", "hashtag4", "hashtag5"],
  "topics_to_engage": ["topic1", "topic2"]
}
`, mediaType, mediaURL, caption)

	response, err := s.callGeminiAPI(ctx, prompt)
	if err != nil {
		s.logger.Error("Failed to call Gemini API", zap.Error(err))
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}

	var analysis models.AIAnalysisResult
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		s.logger.Error("Failed to parse Gemini response", zap.Error(err), zap.String("response", response))
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	s.logger.Debug("Media analyzed",
		zap.String("sentiment", analysis.Sentiment),
		zap.Int("quality_score", analysis.QualityScore),
	)

	return &analysis, nil
}

// GenerateCaptions generates alternative captions for media
func (s *geminiService) GenerateCaptions(ctx context.Context, originalCaption, mediaType string, themes []string) ([]*models.CaptionSuggestion, error) {
	themesStr := ""
	for _, t := range themes {
		themesStr += t + ", "
	}

	prompt := fmt.Sprintf(`
Original caption: "%s"
Media type: %s
Content themes: %s

Generate 3 alternative Instagram captions that:
1. Preserve the original sentiment and key message
2. Include better keyword placement for searchability
3. Add compelling call-to-action
4. Include strategic emoji placement
5. Have optimal hashtag ordering

ONLY return valid JSON without markdown:
{
  "captions": [
    {"caption": "...", "reason": "..."},
    {"caption": "...", "reason": "..."},
    {"caption": "...", "reason": "..."}
  ]
}
`, originalCaption, mediaType, themesStr)

	response, err := s.callGeminiAPI(ctx, prompt)
	if err != nil {
		s.logger.Error("Failed to generate captions", zap.Error(err))
		return nil, fmt.Errorf("caption generation failed: %w", err)
	}

	var result struct {
		Captions []*models.CaptionSuggestion `json:"captions"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		s.logger.Error("Failed to parse captions response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse captions: %w", err)
	}

	s.logger.Debug("Captions generated", zap.Int("count", len(result.Captions)))
	return result.Captions, nil
}

// GeneratePostingStrategy generates optimal posting recommendations
func (s *geminiService) GeneratePostingStrategy(ctx context.Context, analysis *models.AIAnalysisResult, accountID string) (*models.PostingStrategy, error) {
	prompt := fmt.Sprintf(`
Based on this Instagram post analysis:
- Content themes: %v
- Quality: %d/10
- Sentiment: %s
- Engagement potential: %d/100

Provide ONLY valid JSON for optimal posting strategy:
{
  "best_posting_days": ["Monday", "Wednesday", "Friday"],
  "best_posting_times_utc": ["09:00", "14:00", "19:00"],
  "predicted_reach_multiplier": 1.2-3.5,
  "growth_strategy": "clear description",
  "trend_alignment_score": 0-100
}
`, analysis.ContentThemes, analysis.QualityScore, analysis.Sentiment, analysis.EngagementPotential)

	response, err := s.callGeminiAPI(ctx, prompt)
	if err != nil {
		s.logger.Error("Failed to generate posting strategy", zap.Error(err))
		return nil, fmt.Errorf("strategy generation failed: %w", err)
	}

	var strategy models.PostingStrategy
	if err := json.Unmarshal([]byte(response), &strategy); err != nil {
		s.logger.Error("Failed to parse strategy response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse strategy: %w", err)
	}

	s.logger.Debug("Strategy generated",
		zap.Float64("reach_multiplier", strategy.ReachedMultiplier),
		zap.Int("trend_alignment", strategy.TrendAlignment),
	)

	return &strategy, nil
}

// callGeminiAPI makes a request to the Gemini API
func (s *geminiService) callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	// TODO: Implement actual Gemini API call
	// For now, return placeholder
	s.logger.Info("Gemini API call placeholder", zap.String("prompt_len", fmt.Sprintf("%d", len(prompt))))

	return "{}", fmt.Errorf("gemini api integration not yet implemented")
}
