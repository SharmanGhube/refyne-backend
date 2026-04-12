package repository

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramAccountRepository defines operations for Instagram accounts
type InstagramAccountRepository interface {
	// Create a new Instagram account connection
	CreateAccount(c *gin.Context, input *models.CreateInstagramAccountInput) (*models.InstagramAccount, *errors.AppError)

	// Get account by ID
	GetAccountByID(c *gin.Context, accountID string) (*models.InstagramAccount, *errors.AppError)

	// Get account by user ID
	GetAccountsByUserID(c *gin.Context, userID string) ([]*models.InstagramAccount, *errors.AppError)

	// Get account by Instagram user ID
	GetAccountByInstagramUserID(c *gin.Context, instagramUserID string) (*models.InstagramAccount, *errors.AppError)

	// Update account
	UpdateAccount(c *gin.Context, accountID string, input *models.UpdateInstagramAccountInput) *errors.AppError

	// Update sync status
	UpdateSyncStatus(c *gin.Context, accountID, status string, errorMessage *string) *errors.AppError

	// Delete account (soft delete)
	DeleteAccount(c *gin.Context, accountID string) *errors.AppError

	// Check if account exists
	AccountExists(c *gin.Context, accountID string) (bool, *errors.AppError)

	// Get all active accounts for a user
	GetActiveAccountsByUserID(c *gin.Context, userID string) ([]*models.InstagramAccount, *errors.AppError)

	// Check if user has Instagram account connected
	HasInstagramAccount(c *gin.Context, userID string) (bool, *errors.AppError)
}

type instagramAccountRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewInstagramAccountRepository creates a new Instagram account repository
func NewInstagramAccountRepository(db *sqlx.DB) InstagramAccountRepository {
	return &instagramAccountRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("InstagramAccountRepository"),
	}
}

// CreateAccount creates a new Instagram account connection
func (r *instagramAccountRepository) CreateAccount(c *gin.Context, input *models.CreateInstagramAccountInput) (*models.InstagramAccount, *errors.AppError) {
	query := `
		INSERT INTO instagram_accounts (
			user_id, instagram_user_id, username, access_token,
			refresh_token, token_expires_at, profile_picture_url,
			biography, followers_count, connected_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		) RETURNING *
	`

	var account models.InstagramAccount
	err := r.db.QueryRowxContext(c, query,
		input.UserID,
		input.InstagramUserID,
		input.Username,
		input.AccessToken,
		input.RefreshToken,
		input.TokenExpiresAt,
		input.ProfilePictureURL,
		input.Biography,
		input.FollowersCount,
	).StructScan(&account)

	if err != nil {
		r.logger.Error("Failed to create Instagram account", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_CREATE_FAILED",
			"Failed to create Instagram account connection",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	r.logger.Info("Instagram account created", zap.String("account_id", account.ID), zap.String("user_id", input.UserID))
	return &account, nil
}

// GetAccountByID retrieves an Instagram account by ID
func (r *instagramAccountRepository) GetAccountByID(c *gin.Context, accountID string) (*models.InstagramAccount, *errors.AppError) {
	query := `SELECT * FROM instagram_accounts WHERE id = $1 AND deleted_at IS NULL`

	var account models.InstagramAccount
	err := r.db.QueryRowxContext(c, query, accountID).StructScan(&account)

	if err == sql.ErrNoRows {
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_NOT_FOUND",
			"Instagram account not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	if err != nil {
		r.logger.Error("Failed to get Instagram account", zap.Error(err), zap.String("account_id", accountID))
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_FETCH_FAILED",
			"Failed to retrieve Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return &account, nil
}

// GetAccountsByUserID retrieves all Instagram accounts for a user
func (r *instagramAccountRepository) GetAccountsByUserID(c *gin.Context, userID string) ([]*models.InstagramAccount, *errors.AppError) {
	query := `SELECT * FROM instagram_accounts WHERE user_id = $1 AND deleted_at IS NULL ORDER BY connected_at DESC`

	var accounts []*models.InstagramAccount
	err := r.db.SelectContext(c, &accounts, query, userID)

	if err != nil {
		r.logger.Error("Failed to get Instagram accounts for user", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNTS_FETCH_FAILED",
			"Failed to retrieve Instagram accounts",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	if accounts == nil {
		accounts = []*models.InstagramAccount{}
	}

	return accounts, nil
}

// GetAccountByInstagramUserID retrieves an account by Instagram's user ID
func (r *instagramAccountRepository) GetAccountByInstagramUserID(c *gin.Context, instagramUserID string) (*models.InstagramAccount, *errors.AppError) {
	query := `SELECT * FROM instagram_accounts WHERE instagram_user_id = $1 AND deleted_at IS NULL`

	var account models.InstagramAccount
	err := r.db.QueryRowxContext(c, query, instagramUserID).StructScan(&account)

	if err == sql.ErrNoRows {
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_NOT_FOUND",
			"Instagram account not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	if err != nil {
		r.logger.Error("Failed to get Instagram account by Instagram ID", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_FETCH_FAILED",
			"Failed to retrieve Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return &account, nil
}

// UpdateAccount updates an Instagram account
func (r *instagramAccountRepository) UpdateAccount(c *gin.Context, accountID string, input *models.UpdateInstagramAccountInput) *errors.AppError {
	query := `
		UPDATE instagram_accounts SET
			access_token = $1,
			refresh_token = $2,
			token_expires_at = $3,
			profile_picture_url = $4,
			biography = $5,
			followers_count = $6,
			last_sync_at = NOW(),
			updated_at = NOW()
		WHERE id = $7 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(c, query,
		input.AccessToken,
		input.RefreshToken,
		input.TokenExpiresAt,
		input.ProfilePictureURL,
		input.Biography,
		input.FollowersCount,
		accountID,
	)

	if err != nil {
		r.logger.Error("Failed to update Instagram account", zap.Error(err), zap.String("account_id", accountID))
		return errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_UPDATE_FAILED",
			"Failed to update Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_NOT_FOUND",
			"Instagram account not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	r.logger.Info("Instagram account updated", zap.String("account_id", accountID))
	return nil
}

// UpdateSyncStatus updates the sync status of an account
func (r *instagramAccountRepository) UpdateSyncStatus(c *gin.Context, accountID, status string, errorMessage *string) *errors.AppError {
	query := `
		UPDATE instagram_accounts SET
			sync_status = $1,
			sync_error_message = $2,
			updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(c, query, status, errorMessage, accountID)

	if err != nil {
		r.logger.Error("Failed to update sync status", zap.Error(err), zap.String("account_id", accountID))
		return errors.NewAppError(
			c,
			"SYNC_STATUS_UPDATE_FAILED",
			"Failed to update sync status",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return nil
}

// DeleteAccount soft deletes an Instagram account
func (r *instagramAccountRepository) DeleteAccount(c *gin.Context, accountID string) *errors.AppError {
	query := `
		UPDATE instagram_accounts SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(c, query, accountID)

	if err != nil {
		r.logger.Error("Failed to delete Instagram account", zap.Error(err), zap.String("account_id", accountID))
		return errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_DELETE_FAILED",
			"Failed to delete Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_NOT_FOUND",
			"Instagram account not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	r.logger.Info("Instagram account deleted", zap.String("account_id", accountID))
	return nil
}

// AccountExists checks if an account exists
func (r *instagramAccountRepository) AccountExists(c *gin.Context, accountID string) (bool, *errors.AppError) {
	query := `SELECT EXISTS(SELECT 1 FROM instagram_accounts WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(c, query, accountID).Scan(&exists)

	if err != nil {
		r.logger.Error("Failed to check if account exists", zap.Error(err), zap.String("account_id", accountID))
		return false, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_CHECK_FAILED",
			"Failed to check Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return exists, nil
}

// GetActiveAccountsByUserID retrieves active Instagram accounts for a user
func (r *instagramAccountRepository) GetActiveAccountsByUserID(c *gin.Context, userID string) ([]*models.InstagramAccount, *errors.AppError) {
	query := `
		SELECT * FROM instagram_accounts
		WHERE user_id = $1
		AND deleted_at IS NULL
		AND token_expires_at > NOW()
		ORDER BY connected_at DESC
	`

	var accounts []*models.InstagramAccount
	err := r.db.SelectContext(c, &accounts, query, userID)

	if err != nil {
		r.logger.Error("Failed to get active Instagram accounts", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNTS_FETCH_FAILED",
			"Failed to retrieve Instagram accounts",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	if accounts == nil {
		accounts = []*models.InstagramAccount{}
	}

	return accounts, nil
}

// HasInstagramAccount checks if user has any Instagram account connected
func (r *instagramAccountRepository) HasInstagramAccount(c *gin.Context, userID string) (bool, *errors.AppError) {
	query := `SELECT EXISTS(SELECT 1 FROM instagram_accounts WHERE user_id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(c, query, userID).Scan(&exists)

	if err != nil {
		r.logger.Error("Failed to check if user has Instagram account", zap.Error(err), zap.String("user_id", userID))
		return false, errors.NewAppError(
			c,
			"INSTAGRAM_ACCOUNT_CHECK_FAILED",
			"Failed to check Instagram account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return exists, nil
}
