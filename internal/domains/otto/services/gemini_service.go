package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/internal/domains/otto/repository"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type ottoAssistantService struct {
	messageRepo      repository.OttoMessageRepository
	conversationRepo repository.OttoConversationRepository
	apiKey           string
	model            string
	httpClient       *http.Client
	logger           *zap.Logger
}

// NewOttoAssistantService creates an OttoAssistantService backed by Gemini
func NewOttoAssistantService(
	messageRepo repository.OttoMessageRepository,
	conversationRepo repository.OttoConversationRepository,
) OttoAssistantService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash"
	}

	logger := logging.GetServiceLogger("OttoAssistantService")
	if apiKey == "" {
		logger.Warn("GEMINI_API_KEY not set — Otto will return degraded responses")
	}

	return &ottoAssistantService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		apiKey:           apiKey,
		model:            model,
		httpClient:       &http.Client{Timeout: 60 * time.Second},
		logger:           logger,
	}
}

// ProcessMessage fetches conversation history, calls Gemini, persists and returns the assistant reply.
// The caller is responsible for already having saved the user message before calling this.
func (s *ottoAssistantService) ProcessMessage(ctx context.Context, conversationID, userMessage string) (*models.OttoMessage, error) {
	// Fetch last 30 messages for multi-turn context
	history, err := s.messageRepo.GetConversationMessages(ctx, conversationID, 30, 0)
	if err != nil {
		s.logger.Error("Failed to fetch conversation history", zap.Error(err))
		return nil, fmt.Errorf("fetch history: %w", err)
	}

	// Fetch optional ConversationContext (Instagram account, metrics scope, etc.)
	var convCtx *models.ConversationContext
	conv, convErr := s.conversationRepo.GetConversation(ctx, conversationID)
	if convErr == nil && conv.Context != "" {
		var parsed models.ConversationContext
		if jsonErr := json.Unmarshal([]byte(conv.Context), &parsed); jsonErr == nil {
			convCtx = &parsed
		}
	}

	// Generate response (gracefully degrade if Gemini is down)
	responseText, genErr := s.GenerateResponse(ctx, history, convCtx)
	if genErr != nil {
		s.logger.Error("Gemini call failed — using fallback response", zap.Error(genErr))
		responseText = "I'm having trouble reaching my AI backend right now. Please try again in a moment."
	}

	// Persist assistant reply
	assistantMsg := models.NewOttoMessage(conversationID, "", "assistant", responseText)
	assistantMsg.ModelUsed = s.model

	if err := s.messageRepo.CreateMessage(ctx, assistantMsg); err != nil {
		s.logger.Error("Failed to persist assistant message", zap.Error(err))
		return nil, fmt.Errorf("save reply: %w", err)
	}

	s.logger.Info("Otto reply saved",
		zap.String("conversation_id", conversationID),
		zap.Int("response_len", len(responseText)),
	)

	return assistantMsg, nil
}

// GetConversationHistory returns paginated messages.
func (s *ottoAssistantService) GetConversationHistory(ctx context.Context, conversationID string, limit int) ([]*models.OttoMessage, error) {
	return s.messageRepo.GetConversationMessages(ctx, conversationID, limit, 0)
}

// GenerateResponse builds the Gemini multi-turn payload and returns the model's reply text.
func (s *ottoAssistantService) GenerateResponse(ctx context.Context, messages []*models.OttoMessage, convCtx *models.ConversationContext) (string, error) {
	if s.apiKey == "" {
		return "Otto AI is not configured. Please set the GEMINI_API_KEY environment variable.", nil
	}

	type geminiPart struct {
		Text string `json:"text"`
	}
	type geminiContent struct {
		Role  string       `json:"role"`
		Parts []geminiPart `json:"parts"`
	}

	var contents []geminiContent

	// Inject Instagram context as a priming exchange if present
	systemCtx := buildOttoSystemContext(convCtx)
	if systemCtx != "" {
		contents = append(contents,
			geminiContent{Role: "user", Parts: []geminiPart{{Text: systemCtx}}},
			geminiContent{Role: "model", Parts: []geminiPart{{Text: "Got it. I have your Instagram context and will tailor my advice accordingly."}}},
		)
	}

	// Add conversation history (user / assistant turns)
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}

	if len(contents) == 0 {
		return "Hello! I'm Otto, your Instagram strategy assistant. How can I help you today?", nil
	}

	requestBody := map[string]interface{}{
		"contents": contents,
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": "You are Otto, an expert AI assistant for the Refyne platform specialising in Instagram growth, content strategy, analytics interpretation, and audience engagement. Be concise, actionable, and specific. When discussing analytics, give concrete recommendations. Keep responses under 300 words unless asked for detail."},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"topP":            0.95,
			"topK":            64,
			"maxOutputTokens": 2048,
		},
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.model, s.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini http call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Gemini API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return "", fmt.Errorf("gemini returned %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty gemini response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// EnrichContext deserializes the stored JSON context for a conversation.
func (s *ottoAssistantService) EnrichContext(ctx context.Context, conversationID string) (*models.ConversationContext, error) {
	conv, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conv.Context == "" {
		return &models.ConversationContext{}, nil
	}
	var c models.ConversationContext
	if err := json.Unmarshal([]byte(conv.Context), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// buildOttoSystemContext creates a context primer for Gemini from optional Instagram account data.
func buildOttoSystemContext(c *models.ConversationContext) string {
	if c == nil {
		return ""
	}
	var parts []string
	if c.AccountID != "" {
		parts = append(parts, fmt.Sprintf("Instagram Account ID: %s", c.AccountID))
	}
	if c.PlatformType != "" {
		parts = append(parts, fmt.Sprintf("Platform: %s", c.PlatformType))
	}
	if c.MetricsScope != "" {
		parts = append(parts, fmt.Sprintf("Analytics scope: %s", c.MetricsScope))
	}
	if len(parts) == 0 {
		return ""
	}
	result := "Here is the context for this conversation:\n"
	for _, p := range parts {
		result += "- " + p + "\n"
	}
	return result
}
