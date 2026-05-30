package middlewares

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Cookie names
const (
	RefreshTokenCookieName = "refresh_token"
)

// SetRefreshTokenCookie sets the refresh token as an httpOnly, Secure cookie.
// The cookie is NOT accessible via JavaScript, mitigating XSS token theft.
func SetRefreshTokenCookie(c *gin.Context, refreshToken string, expiry time.Duration) {
	secure := os.Getenv("APP_ENV") == "production"

	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/api/auth",          // Scoped to auth endpoints only
		MaxAge:   int(expiry.Seconds()), // e.g. 7 days
		HttpOnly: true,                  // Not accessible via JS — critical for XSS protection
		Secure:   secure,                // Only sent over HTTPS in production
		SameSite: sameSite,              // CSRF protection
	})
}

// ClearRefreshTokenCookie removes the refresh token cookie.
func ClearRefreshTokenCookie(c *gin.Context) {
	secure := os.Getenv("APP_ENV") == "production"

	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// GetRefreshTokenFromCookie extracts the refresh token from the httpOnly cookie.
func GetRefreshTokenFromCookie(c *gin.Context) (string, bool) {
	cookie, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || cookie == "" {
		return "", false
	}
	return cookie, true
}
