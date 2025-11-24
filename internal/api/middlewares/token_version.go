package middlewares

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// TokenVersionCache provides methods to validate token versions
type TokenVersionCache struct {
	db *sqlx.DB
}

// NewTokenVersionCache creates a new token version cache
func NewTokenVersionCache(db *sqlx.DB) *TokenVersionCache {
	return &TokenVersionCache{db: db}
}

// GetUserTokenVersion fetches the current token version for a user
func (tvc *TokenVersionCache) GetUserTokenVersion(ctx context.Context, userID string) (int, error) {
	var tokenVersion int
	query := `SELECT token_version FROM users WHERE id = $1`

	err := tvc.db.GetContext(ctx, &tokenVersion, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, err
	}

	return tokenVersion, nil
}
