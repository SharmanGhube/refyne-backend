package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	// SubscriptionTierKey is the context key for the user's subscription tier
	SubscriptionTierKey = "subscription_tier"
	// SubscriptionStatusKey is the context key for the user's subscription status
	SubscriptionStatusKey = "subscription_status"
)

// RequireSubscription is a middleware that enforces an active Pro subscription.
// It must be placed AFTER AuthMiddleware in the middleware chain so that
// UserIDKey is available in the gin context.
//
// Flow: extract userID → query DB for subscription fields → allow or reject.
func RequireSubscription(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Retrieve authenticated user ID (set by AuthMiddleware)
		userID, ok := GetUserID(c)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Authentication required",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Fetch subscription fields from the database
		var sub struct {
			Tier   string `db:"subscription_tier"`
			Status string `db:"subscription_status"`
		}

		err := db.QueryRowx(
			"SELECT subscription_tier, subscription_status FROM users WHERE id = $1 AND deleted_at IS NULL",
			userID,
		).StructScan(&sub)

		if err != nil {
			logger.Error("Failed to fetch subscription status",
				zap.String("requestID", requestID),
				zap.String("userID", userID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Internal Server Error",
				"message":    "Unable to verify subscription status",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Enforce: only "pro" tier with "active" or "trialing" status
		if sub.Tier != "pro" || (sub.Status != "active" && sub.Status != "trialing") {
			logger.Warn("Subscription required — access denied",
				zap.String("requestID", requestID),
				zap.String("userID", userID),
				zap.String("tier", sub.Tier),
				zap.String("status", sub.Status))
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Forbidden",
				"message":    "An active Pro subscription is required to access this feature",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Inject subscription info into context for downstream handlers
		c.Set(SubscriptionTierKey, sub.Tier)
		c.Set(SubscriptionStatusKey, sub.Status)

		c.Next()
	}
}
