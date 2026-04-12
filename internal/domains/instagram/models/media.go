package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// InstagramMedia represents a cached Instagram post/media
type InstagramMedia struct {
	ID               string         `db:"id" json:"id"`
	AccountID        string         `db:"account_id" json:"account_id"`
	InstagramMediaID string         `db:"instagram_media_id" json:"instagram_media_id"`
	MediaType        string         `db:"media_type" json:"media_type"` // PHOTO, VIDEO, CAROUSEL, REELS, STORY
	Caption          sql.NullString `db:"caption" json:"caption"`
	MediaURL         string         `db:"media_url" json:"media_url"`
	Permalink        sql.NullString `db:"permalink" json:"permalink"`
	ThumbnailURL     sql.NullString `db:"thumbnail_url" json:"thumbnail_url"`

	// Engagement metrics
	LikeCount    int `db:"like_count" json:"like_count"`
	CommentCount int `db:"comment_count" json:"comment_count"`
	SharesCount  int `db:"shares_count" json:"shares_count"`
	Impressions  int `db:"impressions" json:"impressions"`
	Reach        int `db:"reach" json:"reach"`

	// Timestamps
	PostedAt time.Time `db:"posted_at" json:"posted_at"`
	SyncedAt time.Time `db:"synced_at" json:"synced_at"`

	// AI Analysis (stored as JSONB)
	AIAnalysis sql.NullString `db:"ai_analysis" json:"ai_analysis"`

	// Timestamps
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

// AIAnalysis represents the structure of AI-generated analysis
type AIAnalysis struct {
	Sentiment    SentimentAnalysis `json:"sentiment,omitempty"`
	ContentType  string            `json:"content_type,omitempty"`
	Themes       []string          `json:"content_themes,omitempty"`
	Hashtags     []string          `json:"recommended_hashtags,omitempty"`
	Topics       []string          `json:"topics_to_engage,omitempty"`
	QualityScore int               `json:"quality_score,omitempty"`
	Engagement   int               `json:"engagement_potential,omitempty"`
	Alignment    string            `json:"sentiment_alignment,omitempty"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Overall string `json:"overall"` // positive, neutral, negative
	Score   int    `json:"score"`   // 0-10
}

// ParseAIAnalysis parses the AI analysis JSON
func (m *InstagramMedia) ParseAIAnalysis() (*AIAnalysis, error) {
	if !m.AIAnalysis.Valid {
		return nil, nil
	}

	var analysis AIAnalysis
	err := json.Unmarshal([]byte(m.AIAnalysis.String), &analysis)
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}

// CreateInstagramMediaInput represents input for creating Instagram media
type CreateInstagramMediaInput struct {
	AccountID        string
	InstagramMediaID string
	MediaType        string
	Caption          string
	MediaURL         string
	Permalink        string
	ThumbnailURL     string
	PostedAt         time.Time
	LikeCount        int
	CommentCount     int
}

// UpdateInstagramMediaInput represents input for updating Instagram media
type UpdateInstagramMediaInput struct {
	LikeCount    int
	CommentCount int
	SharesCount  int
	Impressions  int
	Reach        int
	AIAnalysis   *AIAnalysis
	SyncedAt     time.Time
}

// MediaInsights represents engagement metrics for a media
type MediaInsights struct {
	ID               string    `db:"id" json:"id"`
	MediaID          string    `db:"media_id" json:"media_id"`
	AccountID        string    `db:"account_id" json:"account_id"`
	Impressions      int       `db:"impressions" json:"impressions"`
	Reach            int       `db:"reach" json:"reach"`
	ProfileVisits    int       `db:"profile_visits" json:"profile_visits"`
	Shares           int       `db:"shares" json:"shares"`
	Saves            int       `db:"saves" json:"saves"`
	Clicks           int       `db:"clicks" json:"clicks"`
	EngagementRate   float64   `db:"engagement_rate" json:"engagement_rate"`
	MetricDate       time.Time `db:"metric_date" json:"metric_date"`
	CollectedAt      time.Time `db:"collected_at" json:"collected_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// AccountInsights represents account-level analytics
type AccountInsights struct {
	ID             string    `db:"id" json:"id"`
	AccountID      string    `db:"account_id" json:"account_id"`
	Impressions    int       `db:"impressions" json:"impressions"`
	Reach          int       `db:"reach" json:"reach"`
	ProfileVisits  int       `db:"profile_visits" json:"profile_visits"`
	FollowerCount  int       `db:"follower_count" json:"follower_count"`
	EngagementRate float64   `db:"engagement_rate" json:"engagement_rate"`
	GrowthRate     float64   `db:"growth_rate" json:"growth_rate"`
	MetricDate     time.Time `db:"metric_date" json:"metric_date"`
	CollectedAt    time.Time `db:"collected_at" json:"collected_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// CaptionSuggestion represents an alternative caption
type CaptionSuggestion struct {
	Caption string `json:"caption"`
	Reason  string `json:"reason"`
}

// PostingStrategy represents optimal posting recommendations
type PostingStrategy struct {
	BestPostingDays    []string `json:"best_posting_days"`
	BestPostingTimes   []string `json:"best_posting_times_utc"`
	ReachedMultiplier  float64  `json:"predicted_reach_multiplier"`
	GrowthStrategy     string   `json:"growth_strategy"`
	TrendAlignment     int      `json:"trend_alignment_score"`
}

// AIAnalysisResult represents the AI analysis result
type AIAnalysisResult struct {
	Sentiment           string   `json:"sentiment"`
	SentimentScore      int      `json:"sentiment_score"`
	ContentType         string   `json:"content_type"`
	ContentThemes       []string `json:"content_themes"`
	QualityScore        int      `json:"quality_score"`
	EngagementPotential int      `json:"engagement_potential"`
	RecommendedHashtags []string `json:"recommended_hashtags"`
	TopicsToEngage      []string `json:"topics_to_engage"`
}

// AIRecommendations represents comprehensive AI-generated recommendations
type AIRecommendations struct {
	ID                   string              `db:"id" json:"id"`
	MediaID              string              `db:"media_id" json:"media_id"`
	AccountID            string              `db:"account_id" json:"account_id"`
	Analysis             *AIAnalysisResult   `db:"analysis" json:"analysis"`
	CaptionSuggestions   []*CaptionSuggestion `db:"caption_suggestions" json:"caption_suggestions"`
	PostingStrategy      *PostingStrategy    `db:"posting_strategy" json:"posting_strategy"`
	ConfidenceScore      float64             `db:"confidence_score" json:"confidence_score"`
	GeneratedAt          time.Time           `db:"generated_at" json:"generated_at"`
	ExpiresAt            time.Time           `db:"expires_at" json:"expires_at"`
	UpdatedAt            time.Time           `db:"updated_at" json:"updated_at"`
}
