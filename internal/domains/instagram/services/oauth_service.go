package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/repository"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramOAuthService handles OAuth flow for Instagram
type InstagramOAuthService interface {
	// GenerateAuthURL generates the OAuth authorization URL for Instagram
	GenerateAuthURL(state string) string

	// HandleCallback handles the OAuth callback and exchanges code for token
	HandleCallback(c *gin.Context, userID, code, state string) ([]*models.InstagramAccount, *errors.AppError)

	// RefreshToken refreshes an expired access token
	RefreshToken(c *gin.Context, accountID string) *errors.AppError

	// DisconnectAccount disconnects an Instagram account
	DisconnectAccount(c *gin.Context, accountID string) *errors.AppError

	// GetDecryptedAccessToken returns the decrypted access token for an account
	GetDecryptedAccessToken(ctx context.Context, accountID string) (string, error)
}

type instagramOAuthService struct {
	config      *config.InstagramConfig
	accountRepo repository.InstagramAccountRepository
	db          *sqlx.DB
	httpClient  *http.Client
	logger      *zap.Logger
}

// NewInstagramOAuthService creates a new Instagram OAuth service
func NewInstagramOAuthService(
	cfg *config.InstagramConfig,
	accountRepo repository.InstagramAccountRepository,
	db *sqlx.DB,
) InstagramOAuthService {
	return &instagramOAuthService{
		config:      cfg,
		accountRepo: accountRepo,
		db:          db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logging.GetServiceLogger("InstagramOAuthService"),
	}
}

// GenerateAuthURL generates the OAuth authorization URL for Facebook Login (to manage Instagram)
func (s *instagramOAuthService) GenerateAuthURL(state string) string {
	baseURL := "https://www.facebook.com/v25.0/dialog/oauth"

	params := url.Values{}
	params.Set("client_id", s.config.AppID)
	params.Set("redirect_uri", s.config.OAuthRedirectURI)
	// Facebook Login scopes required to manage Instagram DMs and Comments, plus business_management in case the page is in a Business Portfolio
	params.Set("scope", "instagram_basic,instagram_manage_messages,instagram_manage_comments,pages_show_list,pages_read_engagement,business_management")
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("extras", `{"setup":{"channel":"IG_API_ONBOARDING"}}`)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// tokenExchangeResponse represents the response from Facebook's token endpoint
type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// fbAccountsResponse represents Facebook Pages and their linked Instagram accounts
type fbAccountsResponse struct {
	Data []struct {
		ID                       string `json:"id"`
		Name                     string `json:"name"`
		AccessToken              string `json:"access_token"`
		InstagramBusinessAccount struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"instagram_business_account,omitempty"`
	} `json:"data"`
}

// userInfoResponse represents the extracted Instagram Business Account details
type userInfoResponse struct {
	ID                string
	Username          string
	Name              string
	ProfilePictureURL string
	Biography         string
	Website           string
	FollowersCount    int
	FollowsCount      int
	MediaCount        int
}

// HandleCallback handles the OAuth callback and exchanges code for token
func (s *instagramOAuthService) HandleCallback(c *gin.Context, userID, code, state string) ([]*models.InstagramAccount, *errors.AppError) {
	if code == "" {
		s.logger.Warn("OAuth callback missing authorization code")
		return nil, errors.NewAppError(
			c,
			"OAUTH_MISSING_CODE",
			"Missing authorization code in OAuth callback",
			errors.ErrorTypeValidation,
			errors.SeverityMedium,
			"instagram",
		)
	}

	// Exchange code for access token
	tokenResp, err := s.exchangeCodeForToken(code)
	if err != nil {
		s.logger.Error("Failed to exchange code for token", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"OAUTH_TOKEN_EXCHANGE_FAILED",
			"Failed to exchange authorization code for access token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	// Get user info (username, profile details) for all linked IG accounts
	userInfos, err := s.getUserInfo(tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to fetch user info", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"OAUTH_USER_INFO_FAILED",
			"Failed to fetch Instagram user information",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	// Encrypt tokens for storage
	encryptedToken, encErr := s.encryptToken(tokenResp.AccessToken)
	if encErr != nil {
		s.logger.Error("Failed to encrypt access token", zap.Error(encErr))
		return nil, errors.NewAppError(
			c,
			"TOKEN_ENCRYPTION_FAILED",
			"Failed to encrypt access token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	encryptedRefreshToken := ""

	// Facebook Graph API long-lived tokens expire in 60 days
	tokenExpiresAt := time.Now().AddDate(0, 0, 60)
	if tokenResp.ExpiresIn > 0 {
		tokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	var savedAccounts []*models.InstagramAccount

	for _, info := range userInfos {
		// Create account input
		accountInput := &models.CreateInstagramAccountInput{
			UserID:            userID,
			InstagramUserID:   info.ID,
			Username:          info.Username,
			AccessToken:       encryptedToken,
			RefreshToken:      encryptedRefreshToken,
			TokenExpiresAt:    tokenExpiresAt,
			ProfilePictureURL: info.ProfilePictureURL,
			Biography:         info.Biography,
			FollowersCount:    info.FollowersCount,
		}

		// Store account in database
		account, repoErr := s.accountRepo.CreateAccount(c, accountInput)
		if repoErr != nil {
			s.logger.Error("Failed to create Instagram account", zap.Error(repoErr), zap.String("ig_user_id", info.ID))
			continue
		}

		savedAccounts = append(savedAccounts, account)
		s.logger.Info("Instagram account connected successfully",
			zap.String("user_id", userID),
			zap.String("instagram_user_id", info.ID),
			zap.String("username", info.Username),
		)
	}

	if len(savedAccounts) == 0 {
		return nil, errors.NewAppError(
			c,
			"OAUTH_NO_ACCOUNTS_SAVED",
			"Failed to save any linked Instagram accounts",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return savedAccounts, nil
}

// exchangeCodeForToken exchanges the authorization code for an access token via Facebook Graph API
func (s *instagramOAuthService) exchangeCodeForToken(code string) (*tokenExchangeResponse, error) {
	// Facebook token endpoint
	tokenURL := "https://graph.facebook.com/v25.0/oauth/access_token"

	// Prepare request body
	data := url.Values{}
	data.Set("client_id", s.config.AppID)
	data.Set("client_secret", s.config.AppSecret)
	data.Set("redirect_uri", s.config.OAuthRedirectURI)
	data.Set("code", code)

	// Make request
	resp, err := s.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make token exchange request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check for errors in response
	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("Facebook token exchange failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("facebook API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp tokenExchangeResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	// Step 2: Exchange the short-lived token for a long-lived token (60 days)
	longLivedURL := "https://graph.facebook.com/v25.0/oauth/access_token"
	llData := url.Values{}
	llData.Set("grant_type", "fb_exchange_token")
	llData.Set("client_id", s.config.AppID)
	llData.Set("client_secret", s.config.AppSecret)
	llData.Set("fb_exchange_token", tokenResp.AccessToken)

	llResp, llErr := s.httpClient.Get(fmt.Sprintf("%s?%s", longLivedURL, llData.Encode()))
	if llErr == nil {
		defer llResp.Body.Close()
		if llResp.StatusCode == http.StatusOK {
			llBody, _ := io.ReadAll(llResp.Body)
			var llTokenResp tokenExchangeResponse
			if json.Unmarshal(llBody, &llTokenResp) == nil && llTokenResp.AccessToken != "" {
				s.logger.Info("Successfully exchanged for long-lived Facebook token")
				return &llTokenResp, nil
			}
		} else {
			llBody, _ := io.ReadAll(llResp.Body)
			s.logger.Warn("Failed to exchange for long-lived token", zap.String("response", string(llBody)))
		}
	} else {
		s.logger.Warn("Failed request to exchange for long-lived token", zap.Error(llErr))
	}

	// Fallback to short-lived token if long-lived exchange failed
	return &tokenResp, nil
}

// getUserInfo fetches user information from Facebook Graph API to find the linked Instagram Business Account
func (s *instagramOAuthService) getUserInfo(accessToken string) ([]*userInfoResponse, error) {
	// Facebook accounts endpoint to get pages and linked instagram accounts
	userURL := fmt.Sprintf("https://graph.facebook.com/v25.0/me/accounts?fields=id,name,access_token,instagram_business_account{id,username}&access_token=%s", url.QueryEscape(accessToken))

	resp, err := s.httpClient.Get(userURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Warn("Failed to get user info from Facebook",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("facebook API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var accountsResp fbAccountsResponse
	if err := json.Unmarshal(body, &accountsResp); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	totalPages := len(accountsResp.Data)
	var igAccounts []*userInfoResponse

	// Fetch detailed info for each linked Instagram Business Account
	for _, page := range accountsResp.Data {
		if page.InstagramBusinessAccount.ID != "" && page.AccessToken != "" {
			igInfoUrl := fmt.Sprintf("https://graph.facebook.com/v25.0/%s?fields=id,username,name,profile_picture_url,biography,website,followers_count,follows_count,media_count&access_token=%s", 
                page.InstagramBusinessAccount.ID, 
                url.QueryEscape(page.AccessToken),
            )
            
            igResp, err := s.httpClient.Get(igInfoUrl)
            if err == nil && igResp.StatusCode == http.StatusOK {
                bodyBytes, _ := io.ReadAll(igResp.Body)
                var igData struct {
                    ID                string `json:"id"`
                    Username          string `json:"username"`
                    Name              string `json:"name"`
                    ProfilePictureURL string `json:"profile_picture_url"`
                    Biography         string `json:"biography"`
                    Website           string `json:"website"`
                    FollowersCount    int    `json:"followers_count"`
                    FollowsCount      int    `json:"follows_count"`
                    MediaCount        int    `json:"media_count"`
                }
                if err := json.Unmarshal(bodyBytes, &igData); err == nil {
                    igAccounts = append(igAccounts, &userInfoResponse{
                        ID:                igData.ID,
                        Username:          igData.Username,
                        Name:              igData.Name,
                        ProfilePictureURL: igData.ProfilePictureURL,
                        Biography:         igData.Biography,
                        Website:           igData.Website,
                        FollowersCount:    igData.FollowersCount,
                        FollowsCount:      igData.FollowsCount,
                        MediaCount:        igData.MediaCount,
                    })
                }
            }
            if igResp != nil {
                igResp.Body.Close()
            }
		}
	}

	s.logger.Info("Facebook accounts retrieved",
		zap.Int("total_pages", totalPages),
		zap.Int("instagram_accounts_connected", len(igAccounts)),
	)

	if len(igAccounts) > 0 {
		return igAccounts, nil
	}

	s.logger.Warn("No instagram business account found in Facebook response",
		zap.String("response_body", string(body)),
	)
	return nil, fmt.Errorf("no linked instagram business account found for this facebook user")
}

// encryptToken encrypts a token using AES-256-GCM
func (s *instagramOAuthService) encryptToken(token string) (string, error) {
	// Get encryption key - in production, this should come from environment/secrets
	// For now, use a derived key from config
	key := s.deriveEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken decrypts a token encrypted with AES-256-GCM
func (s *instagramOAuthService) decryptToken(encryptedToken string) (string, error) {
	key := s.deriveEncryptionKey()

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted token: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	token, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return string(token), nil
}

// GetDecryptedAccessToken returns the decrypted access token for an account
func (s *instagramOAuthService) GetDecryptedAccessToken(ctx context.Context, accountID string) (string, error) {
	var encryptedToken string
	query := "SELECT access_token FROM instagram_accounts WHERE id = $1 AND deleted_at IS NULL"
	err := s.db.QueryRowContext(ctx, query, accountID).Scan(&encryptedToken)
	if err != nil {
		return "", fmt.Errorf("failed to get account access token: %w", err)
	}

	if encryptedToken == "" {
		return "", fmt.Errorf("account has no access token")
	}

	return s.decryptToken(encryptedToken)
}

// deriveEncryptionKey derives a 32-byte key from the app secret (for AES-256)
func (s *instagramOAuthService) deriveEncryptionKey() []byte {
	// In production, use a proper key derivation function (PBKDF2, Argon2, etc.)
	// For now, use SHA256 of the app secret plus a constant salt
	hash := sha256.Sum256([]byte(s.config.AppSecret + "instagram-token-encryption-salt"))
	return hash[:]
}

// RefreshToken refreshes an expired access token
func (s *instagramOAuthService) RefreshToken(c *gin.Context, accountID string) *errors.AppError {
	account, err := s.accountRepo.GetAccountByID(c, accountID)
	if err != nil {
		return err
	}

	// Check if refresh token exists and is available
	if !account.RefreshToken.Valid || account.RefreshToken.String == "" {
		s.logger.Warn("Account has no refresh token available",
			zap.String("account_id", accountID),
		)
		return errors.NewAppError(
			c,
			"NO_REFRESH_TOKEN",
			"Refresh token not available for this account",
			errors.ErrorTypeValidation,
			errors.SeverityMedium,
			"instagram",
		)
	}

	// Decrypt the refresh token
	decryptedRefreshToken, decErr := s.decryptToken(account.RefreshToken.String)
	if decErr != nil {
		s.logger.Error("Failed to decrypt refresh token", zap.Error(decErr))
		return errors.NewAppError(
			c,
			"TOKEN_DECRYPTION_FAILED",
			"Failed to decrypt refresh token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	// Call Instagram's token refresh endpoint
	tokenURL := "https://graph.instagram.com/v18.0/oauth/access_token"
	data := url.Values{}
	data.Set("grant_type", "refresh_access_token")
	data.Set("access_token", decryptedRefreshToken)

	resp, httpErr := s.httpClient.PostForm(tokenURL, data)
	if httpErr != nil {
		s.logger.Error("Failed to refresh token", zap.Error(httpErr))
		return errors.NewAppError(
			c,
			"TOKEN_REFRESH_FAILED",
			"Failed to refresh Instagram access token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("Instagram token refresh failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return errors.NewAppError(
			c,
			"TOKEN_REFRESH_FAILED",
			"Instagram API failed to refresh token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	var tokenResp tokenExchangeResponse
	if parseErr := json.Unmarshal(body, &tokenResp); parseErr != nil {
		s.logger.Error("Failed to parse token refresh response", zap.Error(parseErr))
		return errors.NewAppError(
			c,
			"TOKEN_REFRESH_PARSE_FAILED",
			"Failed to parse token refresh response",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	// Encrypt new token
	encryptedToken, encErr := s.encryptToken(tokenResp.AccessToken)
	if encErr != nil {
		s.logger.Error("Failed to encrypt new access token", zap.Error(encErr))
		return errors.NewAppError(
			c,
			"TOKEN_ENCRYPTION_FAILED",
			"Failed to encrypt new access token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	// Update account with new token
	tokenExpiresAt := time.Now().AddDate(0, 0, 60)
	updateInput := &models.UpdateInstagramAccountInput{
		AccessToken:    encryptedToken,
		TokenExpiresAt: tokenExpiresAt,
	}

	if updateErr := s.accountRepo.UpdateAccount(c, accountID, updateInput); updateErr != nil {
		s.logger.Error("Failed to update account with new token", zap.Error(updateErr))
		return updateErr
	}

	s.logger.Info("Token refreshed successfully",
		zap.String("account_id", accountID),
		zap.String("new_expiry", tokenExpiresAt.String()),
	)

	return nil
}

// DisconnectAccount disconnects an Instagram account
func (s *instagramOAuthService) DisconnectAccount(c *gin.Context, accountID string) *errors.AppError {
	account, err := s.accountRepo.GetAccountByID(c, accountID)
	if err != nil {
		return err
	}

	// TODO: Optionally revoke the access token on Instagram's side
	// This would involve calling Instagram's revoke endpoint with the app token

	// Delete account locally
	if delErr := s.accountRepo.DeleteAccount(c, accountID); delErr != nil {
		return delErr
	}

	s.logger.Info("Instagram account disconnected",
		zap.String("account_id", accountID),
		zap.String("instagram_user_id", account.InstagramUserID),
	)

	return nil
}
