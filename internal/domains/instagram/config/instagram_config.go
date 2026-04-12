package config

import (
	"os"
	"time"

	"go.uber.org/zap"
)

// InstagramConfig holds Instagram API configuration
type InstagramConfig struct {
	AppID              string
	AppSecret          string
	AccessToken        string
	RefreshToken       string
	Environment        string // sandbox or production
	OAuthRedirectURI   string
	WebhookVerifyToken string
	Logger             *zap.Logger
}

// GeminiConfig holds Gemini API configuration
type GeminiConfig struct {
	APIKey            string
	Model             string
	MaxTokens         int
	Temperature       float32
	TopP              float32
	TopK              int
	TimeoutSeconds    int
	RequestsPerMinute int
	Logger            *zap.Logger
}

// NewInstagramConfig creates a new Instagram configuration from environment variables
func NewInstagramConfig(logger *zap.Logger) (*InstagramConfig, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "sandbox"
	}

	var appID, appSecret, accessToken, refreshToken string

	if env == "production" {
		appID = os.Getenv("INSTAGRAM_APP_ID")
		appSecret = os.Getenv("INSTAGRAM_APP_SECRET")
		accessToken = os.Getenv("INSTAGRAM_ACCESS_TOKEN")
		refreshToken = os.Getenv("INSTAGRAM_REFRESH_TOKEN")
	} else {
		// Sandbox credentials
		appID = os.Getenv("INSTAGRAM_SANDBOX_APP_ID")
		appSecret = os.Getenv("INSTAGRAM_SANDBOX_APP_SECRET")
		accessToken = os.Getenv("INSTAGRAM_SANDBOX_ACCESS_TOKEN")
		refreshToken = os.Getenv("INSTAGRAM_SANDBOX_REFRESH_TOKEN")
	}

	if appID == "" {
		logger.Info("INSTAGRAM_APP_ID not configured, using stub credentials for sandbox mode")
		appID = "stub-sandbox-app-id"
	}

	if appSecret == "" {
		logger.Info("INSTAGRAM_APP_SECRET not configured, using stub credentials for sandbox mode")
		appSecret = "stub-sandbox-app-secret"
	}

	oauthRedirectURI := os.Getenv("INSTAGRAM_OAUTH_REDIRECT_URI")
	if oauthRedirectURI == "" {
		oauthRedirectURI = "http://localhost:8080/api/instagram/auth/callback"
	}

	webhookVerifyToken := os.Getenv("INSTAGRAM_WEBHOOK_VERIFY_TOKEN")
	if webhookVerifyToken == "" {
		webhookVerifyToken = "refyne_webhook_token"
	}

	return &InstagramConfig{
		AppID:              appID,
		AppSecret:          appSecret,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		Environment:        env,
		OAuthRedirectURI:   oauthRedirectURI,
		WebhookVerifyToken: webhookVerifyToken,
		Logger:             logger,
	}, nil
}

// NewGeminiConfig creates a new Gemini configuration from environment variables
func NewGeminiConfig(logger *zap.Logger) (*GeminiConfig, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Return nil config if not configured - AI features will be optional
		logger.Warn("GEMINI_API_KEY not configured, AI features will be disabled")
		return nil, nil
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash"
	}

	maxTokens := 4096
	temperature := float32(0.7)
	topP := float32(0.95)
	topK := 64
	timeoutSeconds := 30
	requestsPerMinute := 60000

	return &GeminiConfig{
		APIKey:            apiKey,
		Model:             model,
		MaxTokens:         maxTokens,
		Temperature:       temperature,
		TopP:              topP,
		TopK:              topK,
		TimeoutSeconds:    timeoutSeconds,
		RequestsPerMinute: requestsPerMinute,
		Logger:            logger,
	}, nil
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	GlobalLimit      int           // 200 calls/hour per account
	Buffer           int           // Reserve (keep below 180)
	HourlyWindow     time.Duration // 1 hour
	TokenRefreshDays int           // Refresh every 23 days (55 days before 60-day expiry)
}

// NewRateLimitConfig creates a rate limit configuration for Instagram API
func NewRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		GlobalLimit:      200,
		Buffer:           180,
		HourlyWindow:     time.Hour,
		TokenRefreshDays: 23,
	}
}
