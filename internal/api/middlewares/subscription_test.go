package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRequireSubscription_NoAuth verifies that missing auth returns 401.
func TestRequireSubscription_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	// NOTE: RequireSubscription needs a *sqlx.DB but we pass nil here because
	// the middleware should bail out at the auth check before touching DB.
	r.Use(RequireSubscription(nil))
	r.GET("/premium", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"feature": "unlocked"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/premium", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

// TestRequireSubscription_EmptyUserID verifies that an empty user ID returns 401.
func TestRequireSubscription_EmptyUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	// Simulate auth middleware setting an empty user ID
	r.Use(func(c *gin.Context) {
		c.Set(UserIDKey, "")
		c.Next()
	})
	r.Use(RequireSubscription(nil))
	r.GET("/premium", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"feature": "unlocked"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/premium", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRequireSubscription_ResponseFormat verifies error response structure.
func TestRequireSubscription_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequireSubscription(nil))
	r.GET("/premium", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"feature": "unlocked"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/premium", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "error")
	assert.Contains(t, body, "message")
	assert.Contains(t, body, "request_id")
}
